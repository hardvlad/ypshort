package server

import (
	"net/http"
)

func StartServer(addr string, mux http.Handler) {
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
