package main

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"strings"
	"testing"

	"github.com/hardvlad/ypshort/internal/audit"
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/repository"
)

func BenchmarkTestAdd(b *testing.B) {

	observer := audit.InitObserver()

	conf := config.NewConfig("http://localhost:8080/", "", 6)
	storage, _ := repository.NewStorage(conf.FileName, nil)
	mux := handler.NewHandlers(conf, storage, nil, observer)

	for i := 0; i < b.N; i++ {
		domain := "https://" + getRandomString(8) + "/"
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(domain))

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, request)

		res := w.Result()
		res.Body.Close()
	}
}

func getRandomString(len int) string {
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	str := make([]byte, len)
	for i := 0; i < len; i++ {
		str[i] = alphabet[rand.Intn(26)]
	}
	return string(str)
}
