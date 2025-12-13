package main

import (
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

	handler.NewHandlers(config.NewConfig(), repository.NewStorage())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)

			w := httptest.NewRecorder()
			handler.ShortenerHandler(w, request)

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

	handler.NewHandlers(config.NewConfig(), repository.NewStorage())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))

			w := httptest.NewRecorder()
			handler.ShortenerHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, string(resBody), handler.HandlersData.Conf.ServerAddress)
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
			name:   "add test #1",
			method: http.MethodGet,
			target: "/xxxxxxxxxx",
			want: want{
				code:     307,
				response: "https://ya.ru",
			},
		},
	}

	handler.NewHandlers(config.NewConfig(), repository.NewStorage())
	handler.HandlersData.Store.Set(`xxxxxxxxxx`, "https://ya.ru")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)

			w := httptest.NewRecorder()
			handler.ShortenerHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			location := res.Header.Get("Location")
			assert.Equal(t, test.want.response, location)
		})
	}
}
