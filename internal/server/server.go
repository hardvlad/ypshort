// Package server contains the HTTP server start logic.
package server

import (
	"net/http"
	_ "net/http/pprof"
)

func StartServer(enableHTTPS bool, srv *http.Server) error {
	if enableHTTPS {
		return srv.ListenAndServeTLS("cert.pem", "key.pem")
	}
	return srv.ListenAndServe()
}
