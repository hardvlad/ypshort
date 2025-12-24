package handler

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Conf   *config.Config
	Store  repository.StorageInterface
	Logger *zap.SugaredLogger
}

type shortenerResponse struct {
	isError     bool
	message     string
	redirectURL string
	code        int
}

type URLRequest struct {
	URL string `json:"url"`
}

type BatchURLRequest struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type BatchURLResponseObject struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

func createPostHandler(data Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeResponse(w, r, shortenerResponse{
				isError: true,
				message: "can't read body",
				code:    http.StatusBadRequest,
			})
			return
		}

		writeResponse(w, r, processNewURL(data, string(bodyBytes)))
	}
}

func createGetHandler(data Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, r, processRedirect(data, chi.URLParam(r, "code")))
	}
}

func createPostJSONHandler(data Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req URLRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeResponse(w, r, shortenerResponse{
				isError: true,
				message: "can't decode JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		if req.URL == "" {
			writeResponse(w, r, shortenerResponse{
				isError: true,
				message: "please post URL in JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		resp := processNewURL(data, req.URL)
		if resp.isError {
			writeResponse(w, r, resp)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.code)
			json.NewEncoder(w).Encode(map[string]string{"result": resp.message})
		}
	}
}

func createPostJSONBatchHandler(data Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp []BatchURLResponseObject

		var req []BatchURLRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeResponse(w, r, shortenerResponse{
				isError: true,
				message: "can't decode JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		if len(req) == 0 {
			writeResponse(w, r, shortenerResponse{
				isError: true,
				message: "please post correct JSON",
				code:    http.StatusBadRequest,
			})
			return
		}

		for _, urlData := range req {
			maxAttempts := 5
			var shortLink string

			success := false
			for i := 0; i < maxAttempts; i++ {
				shortLink = GenerateRandomString(data.Conf)
				code, err := data.Store.Set(shortLink, urlData.URL)
				if err != nil {
					if errors.Is(err, repository.ErrorKeyExists) {
						continue
					} else {
						data.Logger.Debugw(err.Error(), "event", "добавление URL", "url", urlData.URL)
						break
					}
				} else {
					success = true
					shortLink = code
					break
				}
			}

			if success {
				fullURL, _ := url.JoinPath(data.Conf.ServerAddress, shortLink)
				resp = append(resp, BatchURLResponseObject{ID: urlData.ID, URL: fullURL})
			}
		}

		if len(resp) == 0 {
			writeResponse(w, r, shortenerResponse{
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

func createPingDBHandler(data Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, r, pingDB(data))
	}
}

func NewHandlers(conf *config.Config, store repository.StorageInterface, sugarLogger *zap.SugaredLogger) http.Handler {

	mux := chi.NewRouter()

	handlersData := Handlers{
		Conf:   conf,
		Store:  store,
		Logger: sugarLogger,
	}

	mux.Post(`/`, createPostHandler(handlersData))
	mux.Get(`/{code}`, createGetHandler(handlersData))
	mux.Post(`/api/shorten`, createPostJSONHandler(handlersData))
	mux.Get(`/ping`, createPingDBHandler(handlersData))
	mux.Post(`/api/shorten/batch`, createPostJSONBatchHandler(handlersData))

	return mux
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp shortenerResponse) {
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

func pingDB(data Handlers) shortenerResponse {
	database, err := data.Conf.DBConfig.InitDB()
	if err != nil {
		if data.Logger != nil {
			data.Logger.Errorw(err.Error(), "event", "соединение с базой данных")
		}
		return shortenerResponse{
			isError: false,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	database.Close()

	return shortenerResponse{
		isError: false,
		message: http.StatusText(http.StatusOK),
		code:    http.StatusOK,
	}
}

func processRedirect(data Handlers, path string) shortenerResponse {
	if urlRedirect, ok := data.Store.Get(path); ok {
		return shortenerResponse{
			isError:     false,
			redirectURL: urlRedirect,
			code:        http.StatusTemporaryRedirect,
		}
	}

	return shortenerResponse{
		isError: true,
		message: "short link does not exist",
		code:    http.StatusBadRequest,
	}
}

func processNewURL(data Handlers, body string) shortenerResponse {

	success := false
	maxAttempts := 5
	var shortLink string

	for i := 0; i < maxAttempts; i++ {
		shortLink = GenerateRandomString(data.Conf)
		code, err := data.Store.Set(shortLink, body)
		if err != nil {
			if errors.Is(err, repository.ErrorKeyExists) {
				continue
			} else {
				data.Logger.Debugw(err.Error(), "event", "добавление URL", "url", body)
				break
			}
		} else {
			success = true
			shortLink = code
			break
		}
	}

	if !success {
		return shortenerResponse{
			isError: true,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	fullURL, err := url.JoinPath(data.Conf.ServerAddress, shortLink)
	if err != nil {
		return shortenerResponse{
			isError: false,
			message: http.StatusText(http.StatusInternalServerError),
			code:    http.StatusInternalServerError,
		}
	}

	return shortenerResponse{
		isError: false,
		message: fullURL,
		code:    http.StatusCreated,
	}
}

func GenerateRandomString(conf *config.Config) string {
	b := make([]byte, conf.ShortLinkLength)
	for i := 0; i < conf.ShortLinkLength; i++ {
		b[i] = conf.Charset[rand.Intn(len(conf.Charset))]
	}
	return string(b[:])
}
