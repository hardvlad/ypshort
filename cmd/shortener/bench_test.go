package main

import (
	"context"
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
	"github.com/hardvlad/ypshort/internal/service"
)

func BenchmarkTestAdd(b *testing.B) {

	b.StopTimer()

	observer := audit.InitObserver()

	myLogger, _ := logger.InitLogger()
	defer myLogger.Sync()
	sugarLogger := myLogger.Sugar()

	conf := config.NewConfig("http://localhost:8080/", "", 6, "")
	storage, _ := repository.NewStorage(conf.FileName, sugarLogger)
	shortenerService := service.NewShortenerService(context.Background(), conf, storage, sugarLogger, observer)
	mux := handler.NewHandlers(context.Background(), conf, storage, sugarLogger, observer, shortenerService)

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
