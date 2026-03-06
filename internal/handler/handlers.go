// Package handler creates handlers to handle incoming requests.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hardvlad/ypshort/internal/audit"
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type deleteChannelRequest struct {
	UserID int
	URLs   []string
}

type handlers struct {
	Conf   *config.Config
	Store  repository.StorageInterface
	Logger *zap.SugaredLogger
}

type ShortenerResponse struct {
	isError     bool
	message     string
	redirectURL string
	code        int
}

type urlRequest struct {
	URL string `json:"url"`
}

type batchURLRequest struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type BatchURLResponseObject struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

type StatsResponseObject struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewHandlers initializes and returns an HTTP handler with all defined routes and dependencies injected.
func NewHandlers(ctx context.Context, conf *config.Config, store repository.StorageInterface, sugarLogger *zap.SugaredLogger, observer *audit.Event) http.Handler {

	mux := chi.NewRouter()

	handlersData := handlers{
		Conf:   conf,
		Store:  store,
		Logger: sugarLogger,
	}

	ch := make(chan deleteChannelRequest, 100)
	go deleteWorker(handlersData, ch)

	mux.Post(`/`, createPostHandler(handlersData, observer))
	mux.Get(`/{code}`, CreateGetHandler(handlersData, observer))
	mux.Post(`/api/shorten`, createPostJSONHandler(handlersData, observer))
	mux.Get(`/ping`, createPingDBHandler(handlersData))
	mux.Post(`/api/shorten/batch`, createPostJSONBatchHandler(handlersData))
	mux.Get(`/api/user/urls`, createGetUserURLSHandler(handlersData))
	mux.Delete(`/api/user/urls`, createDeleteUserURLSHandler(handlersData, ch))

	mux.Get(`/api/internal/stats`, createGetStatistics(handlersData))

	return mux
}

func createPostHandler(data handlers, observer *audit.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "can't read body",
				code:    http.StatusBadRequest,
			})
			return
		}

		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		writeResponse(w, r, ProcessNewURL(data, string(bodyBytes), userID))

		updateObserver(observer, "shorten", string(bodyBytes), userID)
	}
}

func updateObserver(observer *audit.Event, action string, url string, userID int) {
	if userID == 0 {
		observer.Update(audit.AuditorEvent{TS: time.Now().Unix(), Action: action, URL: url})
	} else {
		observer.Update(audit.AuditorEvent{TS: time.Now().Unix(), Action: "follow", UserID: strconv.Itoa(userID), URL: url})
	}
}

func CreateGetHandler(data handlers, observer *audit.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		p := processRedirect(data, chi.URLParam(r, "code"))
		writeResponse(w, r, p)

		updateObserver(observer, "shorten", p.redirectURL, userID)
	}
}

func createGetStatistics(data handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		if data.Conf.TrustedSubnet == "" || !isTrustedIP(ip, data.Conf.TrustedSubnet) {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "Forbidden from IP: " + ip,
				code:    http.StatusForbidden,
			})
			return
		}

		a := StatsResponseObject{}
		urls, err := data.Store.GetURLsCount()
		if err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: http.StatusText(http.StatusInternalServerError),
				code:    http.StatusInternalServerError,
			})
			return
		}

		a.URLs = urls

		users, err := data.Store.GetUsersCount()
		if err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: http.StatusText(http.StatusInternalServerError),
				code:    http.StatusInternalServerError,
			})
			return
		}
		a.Users = users

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(a)
	}
}

func isTrustedIP(ip string, subnet string) bool {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil || cidr == nil {
		return false
	}

	return cidr.Contains(ipAddr)
}

func createPostJSONHandler(data handlers, observer *audit.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req urlRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "can't decode JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		if req.URL == "" {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "please post URL in JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		resp := ProcessNewURL(data, req.URL, userID)
		if resp.isError {
			writeResponse(w, r, resp)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.code)
			json.NewEncoder(w).Encode(map[string]string{"result": resp.message})
		}

		updateObserver(observer, "shorten", req.URL, userID)
	}
}

func createPostJSONBatchHandler(data handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp []BatchURLResponseObject

		var req []batchURLRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "can't decode JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		if len(req) == 0 {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "please post correct JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		for _, urlData := range req {
			success, shortLink, _, err := getShortCode(data, urlData.URL, 5, userID)
			if err != nil {
				data.Logger.Debugw(err.Error(), "event", "добавление URL", "url", urlData.URL)
			}

			if success {
				fullURL, _ := url.JoinPath(data.Conf.ServerAddress, shortLink)
				resp = append(resp, BatchURLResponseObject{ID: urlData.ID, URL: fullURL})
			}
		}

		if len(resp) == 0 {
			writeResponse(w, r, ShortenerResponse{
				isError: false,
				message: "no data in response",
				code:    http.StatusBadRequest,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func createPingDBHandler(data handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, r, pingDB(data))
	}
}

func createGetUserURLSHandler(data handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		userURLs, err := data.Store.GetUserData(userID)
		if err != nil {
			data.Logger.Debugw(err.Error(), "event", "получение данных пользователя", "user_id", userID)
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: http.StatusText(http.StatusInternalServerError),
				code:    http.StatusInternalServerError,
			})
			return
		}

		if len(userURLs) == 0 {
			writeResponse(w, r, ShortenerResponse{
				isError: false,
				message: "no URLs for this user",
				code:    http.StatusNoContent,
			})
			return
		}

		type UserURLResponse struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		var resp []UserURLResponse
		for shortCode, originalURL := range userURLs {
			fullURL, _ := url.JoinPath(data.Conf.ServerAddress, shortCode)
			resp = append(resp, UserURLResponse{
				ShortURL:    fullURL,
				OriginalURL: originalURL,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func createDeleteUserURLSHandler(data handlers, ch chan deleteChannelRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var urlsToDelete []string
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&urlsToDelete); err != nil {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "can't decode JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		if len(urlsToDelete) == 0 {
			writeResponse(w, r, ShortenerResponse{
				isError: true,
				message: "please post correct JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok {
			userID = 0
		}

		ch <- deleteChannelRequest{
			UserID: userID,
			URLs:   urlsToDelete,
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func deleteWorker(data handlers, ch chan deleteChannelRequest) {
	for req := range ch {
		err := data.Store.DeleteURLs(req.URLs, req.UserID)
		if err != nil {
			data.Logger.Debugw(err.Error(), "event", "удаление URL", "user_id", req.UserID, "urls", req.URLs)
		}
	}
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp ShortenerResponse) {
	if resp.isError {
		http.Error(w, resp.message, resp.code)
	} else {
		if resp.redirectURL != "" {
			http.Redirect(w, r, resp.redirectURL, resp.code)
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(resp.code)
			_, err := w.Write([]byte(resp.message))
			if err != nil {
				return
			}
		}
	}
}

func pingDB(data handlers) ShortenerResponse {
	database, err := data.Conf.DBConfig.InitDB()
	if err != nil {
		if data.Logger != nil {
			data.Logger.Errorw(err.Error(), "event", "соединение с базой данных")
		}
		return ShortenerResponse{
			isError: false,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	database.Close()

	return ShortenerResponse{
		isError: false,
		message: http.StatusText(http.StatusOK),
		code:    http.StatusOK,
	}
}

func processRedirect(data handlers, path string) ShortenerResponse {
	urlRedirect, isDeleted, ok := data.Store.Get(path)
	if isDeleted {
		return ShortenerResponse{
			isError: true,
			message: "short link was deleted",
			code:    http.StatusGone,
		}
	}

	if ok {
		return ShortenerResponse{
			isError:     false,
			redirectURL: urlRedirect,
			code:        http.StatusTemporaryRedirect,
		}
	}

	return ShortenerResponse{
		isError: true,
		message: "short link does not exist",
		code:    http.StatusBadRequest,
	}
}

func ProcessNewURL(data handlers, body string, userID int) ShortenerResponse {

	success, shortLink, urlAlreadyExisted, err := getShortCode(data, body, 5, userID)
	if err != nil {
		data.Logger.Debugw(err.Error(), "event", "добавление URL", "url", body)
	}

	if !success {
		return ShortenerResponse{
			isError: true,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	fullURL, err := url.JoinPath(data.Conf.ServerAddress, shortLink)
	if err != nil {
		return ShortenerResponse{
			isError: true,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	if urlAlreadyExisted {
		return ShortenerResponse{
			isError: false,
			message: fullURL,
			code:    http.StatusConflict,
		}
	}

	return ShortenerResponse{
		isError: false,
		message: fullURL,
		code:    http.StatusCreated,
	}
}

func generateRandomString(conf *config.Config) string {
	b := make([]byte, conf.ShortLinkLength)
	for i := 0; i < conf.ShortLinkLength; i++ {
		b[i] = conf.Charset[rand.Intn(len(conf.Charset))]
	}
	return string(b[:])
}

func getShortCode(data handlers, body string, maxAttempts int, userID int) (success bool, code string, urlExisted bool, err error) {
	success = false
	var shortLink string
	var urlAlreadyExisted bool

	for i := 0; i < maxAttempts; i++ {
		shortLink = generateRandomString(data.Conf)
		code, urlExisted, err := data.Store.Set(shortLink, body, userID)
		if err != nil {
			if errors.Is(err, repository.ErrorKeyExists) {
				continue
			} else {
				return false, "", false, err
			}
		} else {
			success = true
			shortLink = code
			urlAlreadyExisted = urlExisted
			break
		}
	}
	return success, shortLink, urlAlreadyExisted, err
}
