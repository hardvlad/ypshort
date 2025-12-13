package main

import (
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/server"
)

func main() {
	server.StartServer(`:8080`, handler.NewHandlers(config.NewConfig(), repository.NewStorage()))
}
