package main

import (
	"log"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/repository/pg"
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

	conf := config.NewConfig(flags.ServerAddress, flags.Dsn)

	var store repository.StorageInterface

	db, err := conf.DBConfig.InitDB()
	if err != nil {
		sugarLogger.Infow(err.Error(), "storage", "DB недоступна, используем файловое/in-memory хранилище")
		storage, err := repository.NewStorage(flags.FileName, sugarLogger)
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "init storage, file: "+flags.FileName)
		}
		store = storage
	} else {
		store = pg.NewPGStorage(db, sugarLogger)
		defer db.Close()
	}

	err = server.StartServer(flags.RunAddress, logger.WithLogging(
		handler.RequestDecompressHandle(
			handler.ResponseCompressHandle(
				handler.NewHandlers(conf, store, sugarLogger),
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
