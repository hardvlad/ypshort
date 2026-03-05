package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/hardvlad/ypshort/internal/audit"
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
)

func ExampleProcessNewURL() {
	observer := audit.InitObserver()

	myLogger, _ := logger.InitLogger()
	defer myLogger.Sync()
	sugarLogger := myLogger.Sugar()

	conf := config.NewConfig("http://localhost:8080/", "", 6)
	storage, _ := repository.NewStorage(conf.FileName, sugarLogger)
	mux := handler.NewHandlers(context.Background(), conf, storage, sugarLogger, observer)

	s := httptest.NewServer(mux)
	defer s.Close()

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://baha.com"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, request)

	res := w.Result()

	fmt.Println(res.StatusCode)
	res.Body.Close()

	// Output:
	// 201
}

func ExampleCreateGetHandler() {
	observer := audit.InitObserver()

	myLogger, _ := logger.InitLogger()
	defer myLogger.Sync()
	sugarLogger := myLogger.Sugar()

	conf := config.NewConfig("http://localhost:8080/", "", 6)
	storage, _ := repository.NewStorage(conf.FileName, sugarLogger)
	mux := handler.NewHandlers(context.Background(), conf, storage, sugarLogger, observer)

	_, _, _ = storage.Set(`zzzzzzzzzzz`, "https://baha.com", 0)

	s := httptest.NewServer(mux)
	defer s.Close()

	request := httptest.NewRequest(http.MethodGet, "/zzzzzzzzzzz", nil)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, request)

	res := w.Result()

	location := res.Header.Get("Location")
	res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(location)

	// Output:
	// 307
	// https://baha.com
}
