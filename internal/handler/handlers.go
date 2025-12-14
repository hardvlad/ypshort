package handler

import (
	"io"
	"net/http"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/repository"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Conf  *config.Config
	Store *repository.Storage
	Mux   *chi.Mux
}

type shortenerResponse struct {
	isError     bool
	message     string
	redirectURL string
	code        int
}

var HandlersData Handlers

func NewHandlers(conf *config.Config, store *repository.Storage) http.Handler {

	mux := chi.NewRouter()

	mux.Post(`/`, PostHandler)
	mux.Get(`/{code}`, GetHandler)

	HandlersData = Handlers{
		Conf:  conf,
		Store: store,
		Mux:   mux,
	}

	return mux
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
