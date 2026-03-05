// Package server contains the HTTP server start logic.
package server

import (
	"net/http"
	_ "net/http/pprof"
)

func StartServer(enableHTTPS bool, certFile string, keyFile string, srv *http.Server) error {
	if enableHTTPS {
		return srv.ListenAndServeTLS(certFile, keyFile)
	}
	return srv.ListenAndServe()
}
