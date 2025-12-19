package main

import (
	"log"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/server"
)

func main() {
	flags := parseFlags()
	err := server.StartServer(flags.RunAddress, handler.NewHandlers(config.NewConfig(flags.ServerAddress), repository.NewStorage()))
	if err != nil {
		log.Fatal(err)
	}
}
