package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var shortenerStorage = make(map[string]string)
var shortLinkLength = 6
var serverAddress = "http://localhost:8080/"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func shortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		processNewUrl(w, r)
	}
	if r.Method == http.MethodGet {
		processRedirect(w, r)
	}
}

func processRedirect(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	if url, ok := shortenerStorage[path]; ok {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, "short link does not exist", http.StatusBadRequest)
	return
}

func processNewUrl(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	shortLink := generateRandomString(shortLinkLength)

	if _, ok := shortenerStorage[shortLink]; ok {
		http.Error(w, "short link already exists", http.StatusBadRequest)
		return
	}

	shortenerStorage[shortLink] = string(bodyBytes)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(serverAddress + shortLink))
}

func generateRandomString(length int) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortenerHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
