package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"
)

type Handlers struct {
	Conf  *config.Config
	Store *repository.Storage
	Mux   *http.ServeMux
}

type shortenerResponse struct {
	isError     bool
	message     string
	redirectUrl string
	code        int
}

var HandlersData Handlers

func NewHandlers(conf *config.Config, store *repository.Storage) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, ShortenerHandler)

	HandlersData = Handlers{
		Conf:  conf,
		Store: store,
		Mux:   mux,
	}

	return mux
}

func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
	} else {
		if r.Method == http.MethodGet {
			writeResponse(w, r, processRedirect(strings.TrimPrefix(r.URL.Path, "/")))
		} else {
			http.Error(w, "method not allowed", http.StatusBadRequest)
		}
	}
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp shortenerResponse) {
	if resp.isError {
		http.Error(w, resp.message, resp.code)
	} else {
		if resp.redirectUrl != "" {
			http.Redirect(w, r, resp.redirectUrl, resp.code)
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
			redirectUrl: url,
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
	shortLink := HandlersData.Conf.GenerateRandomString()

	if _, ok := HandlersData.Store.Get(shortLink); ok {
		return shortenerResponse{
			isError: true,
			message: "short link already exists",
			code:    http.StatusBadRequest,
		}
	}

	HandlersData.Store.Set(shortLink, body)

	return shortenerResponse{
		isError: false,
		message: HandlersData.Conf.ServerAddress + shortLink,
		code:    http.StatusCreated,
	}
}
