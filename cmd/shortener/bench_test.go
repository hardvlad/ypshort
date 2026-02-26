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
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
)

func BenchmarkTestAdd(b *testing.B) {

	b.StopTimer()

	observer := audit.InitObserver()

	myLogger, _ := logger.InitLogger()
	defer myLogger.Sync()
	sugarLogger := myLogger.Sugar()

	conf := config.NewConfig("http://localhost:8080/", "", 6)
	storage, _ := repository.NewStorage(conf.FileName, sugarLogger)
	mux := handler.NewHandlers(conf, storage, sugarLogger, observer)

	b.StartTimer()
	for b.Loop() {
		b.StopTimer()
		domain := "https://" + getRandomString(8) + "/"
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(domain))

		w := httptest.NewRecorder()
		b.StartTimer()
		mux.ServeHTTP(w, request)
		b.StopTimer()

		res := w.Result()
		res.Body.Close()
		b.StartTimer()
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
