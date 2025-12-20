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

	storage, err2 := repository.NewStorage(flags.FileName)
	if err2 != nil {
		logger.Sugar.Fatalw(err2.Error(), "event", "init storage, file: "+flags.FileName)
	}

	err3 := server.StartServer(flags.RunAddress, handler.RequestDecompressHandle(handler.ResponseCompressHandle(logger.WithLogging(handler.NewHandlers(config.NewConfig(flags.ServerAddress), storage)))))
	if err3 != nil {
		logger.Sugar.Fatalw(err3.Error(), "event", "start server")
	}
}
