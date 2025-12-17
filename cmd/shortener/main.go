package main

import (
	"log"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/server"
)

func main() {

	myLogger, err1 := logger.InitLogger()
	if err1 != nil {
		log.Fatal(err1)
	}

	defer myLogger.Sync()

	flags := parseFlags()

	logger.Sugar = *myLogger.Sugar()
	logger.Sugar.Infow("Старт сервера", "addr", flags.RunAddress)

	err2 := server.StartServer(flags.RunAddress, logger.WithLogging(handler.NewHandlers(config.NewConfig(flags.ServerAddress), repository.NewStorage())))
	if err2 != nil {
		logger.Sugar.Fatalw(err2.Error(), "event", "start server")
	}
}
