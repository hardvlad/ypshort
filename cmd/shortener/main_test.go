package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBefore(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "negative test before #1",
			method: http.MethodGet,
			target: "/xxxxxxxxx",
			body:   "",
			want: want{
				code:     400,
				response: `short link does not exist` + "\n",
			},
		},
	}

	conf := config.NewConfig("http://localhost:8080/", "")
	storage, err := repository.NewStorage(conf.FileName, nil)
	require.NoError(t, err)
	mux := handler.NewHandlers(conf, storage, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}

func TestAdd(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "add test #1",
			method: http.MethodPost,
			target: "/",
			body:   "https://ya.ru",
			want: want{
				code: 201,
			},
		},
	}

	conf := config.NewConfig("http://localhost:8080/", "")
	storage, err := repository.NewStorage(conf.FileName, nil)
	require.NoError(t, err)
	mux := handler.NewHandlers(conf, storage, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Contains(t, string(resBody), conf.ServerAddress)
		})
	}
}

func TestExisting(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "existing test #1",
			method: http.MethodGet,
			target: "/xxxxxxxxxx",
			want: want{
				code:     307,
				response: "https://ya.ru",
			},
		},
	}

	conf := config.NewConfig("http://localhost:8080/", "")
	storage, err := repository.NewStorage(conf.FileName, nil)
	require.NoError(t, err)
	mux := handler.NewHandlers(conf, storage, nil)
	_, _, err = storage.Set(`xxxxxxxxxx`, "https://ya.ru")
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			location := res.Header.Get("Location")
			assert.Equal(t, test.want.response, location)
			res.Body.Close()
		})
	}
}

func TestAddJson(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "add test JSON #1",
			method: http.MethodPost,
			target: "/api/shorten",
			body:   `{"url":"https://ya.ru"}`,
			want: want{
				code: 201,
			},
		},
	}

	conf := config.NewConfig("http://localhost:8080/", "")
	storage, err := repository.NewStorage(conf.FileName, nil)
	require.NoError(t, err)
	mux := handler.NewHandlers(conf, storage, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)
			assert.Contains(t, string(resBody), "http://localhost:8080/")
		})
	}
}

func TestAddJsonBatch(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "add test JSON Batch #1",
			method: http.MethodPost,
			target: "/api/shorten/batch",
			body:   `[{"correlation_id": "1","original_url": "https://ya.ru"},{"correlation_id": "2","original_url": "https://yandex.ru"}]`,
			want: want{
				code: 201,
			},
		},
	}

	conf := config.NewConfig("http://localhost:8080/", "")
	storage, err := repository.NewStorage(conf.FileName, nil)
	require.NoError(t, err)
	mux := handler.NewHandlers(conf, storage, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)

			var resp []handler.BatchURLResponseObject

			err = json.Unmarshal(resBody, &resp)
			require.NoError(t, err)
			assert.Equal(t, 2, len(resp))
			assert.Equal(t, "1", resp[0].ID)
			assert.Equal(t, "2", resp[1].ID)
		})
	}
}
