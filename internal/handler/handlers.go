package handler

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"
)

type Handlers struct {
	Conf      *config.Config
	Store     *repository.Storage
	Mux       *chi.Mux
	LockMutex sync.Mutex
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

var HandlersData Handlers

func NewHandlers(conf *config.Config, store *repository.Storage) http.Handler {

	mux := chi.NewRouter()

	mux.Post(`/`, PostHandler)
	mux.Get(`/{code}`, GetHandler)
	mux.Post(`/api/shorten`, PostJSONHandler)

	HandlersData = Handlers{
		Conf:  conf,
		Store: store,
		Mux:   mux,
	}

	return mux
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := processNewURL(req.URL)
	if resp.isError {
		writeResponse(w, r, resp)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.code)
		json.NewEncoder(w).Encode(map[string]string{"result": resp.message})
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeResponse(w, r, shortenerResponse{
			isError: true,
			message: "can't read body",
			code:    http.StatusBadRequest,
		})
		return
	}

	writeResponse(w, r, processNewURL(string(bodyBytes)))
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r, processRedirect(chi.URLParam(r, "code")))
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp shortenerResponse) {
	if resp.isError {
		http.Error(w, resp.message, resp.code)
	} else {
		if resp.redirectURL != "" {
			http.Redirect(w, r, resp.redirectURL, resp.code)
		} else {
			w.WriteHeader(resp.code)
			_, err := w.Write([]byte(resp.message))
			if err != nil {
				return
			}
		}
	}
}

func processRedirect(path string) shortenerResponse {
	if url, ok := HandlersData.Store.Get(path); ok {
		return shortenerResponse{
			isError:     false,
			redirectURL: url,
			code:        http.StatusTemporaryRedirect,
		}
	}

	return shortenerResponse{
		isError: true,
		message: "short link does not exist",
		code:    http.StatusBadRequest,
	}
}

func processNewURL(body string) shortenerResponse {

	success := false
	maxAttempts := 5
	var shortLink string

	for i := 0; i < maxAttempts; i++ {
		shortLink = GenerateRandomString(HandlersData.Conf)
		if _, ok := HandlersData.Store.Get(shortLink); !ok {
			success = true
			break
		}
	}

	if !success {
		return shortenerResponse{
			isError: true,
			message: "failed to generate unique short link",
			code:    http.StatusInternalServerError,
		}
	}

	HandlersData.LockMutex.Lock()
	HandlersData.Store.Set(shortLink, body)
	HandlersData.LockMutex.Unlock()

	fullURL := HandlersData.Conf.ServerAddress
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	fullURL += shortLink

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
