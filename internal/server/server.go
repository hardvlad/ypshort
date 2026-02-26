// Package server contains the HTTP server start logic.
package server

import (
	"net/http"
	_ "net/http/pprof"
)

func StartServer(addr string, mux http.Handler) error {
	if addr == "" {
		addr = ":8080"
	}
	return http.ListenAndServe(addr, mux)
}
