// Package server contains the HTTP server start logic.
package server

import (
	"net/http"
	_ "net/http/pprof"
)

func StartServer(addr string, enableHTTPS bool, mux http.Handler) error {
	if addr == "" {
		addr = ":8080"
	}

	if enableHTTPS {
		return http.ListenAndServeTLS(addr, "cert.pem", "key.pem", mux)
	}
	return http.ListenAndServe(addr, mux)
}
