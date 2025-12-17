package server

import (
	"net/http"
)

func StartServer(addr string, mux http.Handler) error {
	return http.ListenAndServe(addr, mux)
}
