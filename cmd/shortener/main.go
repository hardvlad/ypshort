package main

import (
	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/server"
)

func main() {
	parseFlags()
	server.StartServer(runAddress, handler.NewHandlers(config.NewConfig(serverAddress), repository.NewStorage()))
}
