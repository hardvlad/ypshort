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

	myLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatal(err)
	}

	defer myLogger.Sync()

	flags := parseFlags()

	sugarLogger := myLogger.Sugar()
	sugarLogger.Infow("Старт сервера", "addr", flags.RunAddress)

	storage, err := repository.NewStorage(flags.FileName, sugarLogger)
	if err != nil {
		sugarLogger.Fatalw(err.Error(), "event", "init storage, file: "+flags.FileName)
	}

	err = server.StartServer(flags.RunAddress, logger.WithLogging(
		handler.RequestDecompressHandle(
			handler.ResponseCompressHandle(
				handler.NewHandlers(config.NewConfig(flags.ServerAddress, flags.Dsn), storage, sugarLogger),
				sugarLogger,
			),
			sugarLogger,
		),
		sugarLogger,
	),
	)

	if err != nil {
		sugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}
