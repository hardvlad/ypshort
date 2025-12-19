package handler

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Conf  *config.Config
	Store *repository.Storage
}

type shortenerResponse struct {
	isError     bool
	message     string
	redirectURL string
	code        int
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

func NewHandlers(conf *config.Config, store *repository.Storage) http.Handler {

	mux := chi.NewRouter()

	handlersData := Handlers{
		Conf:  conf,
		Store: store,
	}

	mux.Post(`/`, createPostHandler(handlersData))
	mux.Get(`/{code}`, createGetHandler(handlersData))

	return mux
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
		err := data.Store.Set(shortLink, body)
		if err != nil {
			if errors.Is(err, repository.ErrorKeyExists) {
				continue
			}
		} else {
			success = true
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
