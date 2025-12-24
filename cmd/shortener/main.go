package main

import (
	"log"

	"github.com/hardvlad/ypshort/internal/config"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/repository/pg"
	"github.com/hardvlad/ypshort/internal/server"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
		store, err = repository.NewStorage(flags.FileName, sugarLogger)
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "init storage, file: "+flags.FileName)
		}
	} else {
		store = pg.NewPGStorage(db, sugarLogger)
		defer db.Close()

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "подготовка к миграции")
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://./migrations",
			"postgres", driver)
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "подготовка к миграции 2")
		}
		err = m.Up()
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "применение миграции")
		}
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
