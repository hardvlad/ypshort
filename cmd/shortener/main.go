package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hardvlad/ypshort/internal/audit"
	"github.com/hardvlad/ypshort/internal/config"
	ypgrpc "github.com/hardvlad/ypshort/internal/grpc"
	"github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/logger"
	"github.com/hardvlad/ypshort/internal/repository"
	"github.com/hardvlad/ypshort/internal/repository/pg"
	"github.com/hardvlad/ypshort/internal/server"
	"github.com/hardvlad/ypshort/internal/service"
	pb "github.com/hardvlad/ypshort/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {

	printBuildInfo()

	myLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatal(err)
	}

	defer myLogger.Sync()

	flags := parseFlags()

	sugarLogger := myLogger.Sugar()

	conf := config.NewConfig(flags.ServerAddress, flags.Dsn, flags.Length, flags.TrustedSubnet)

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
			sugarLogger.Fatalw(err.Error(), "event", "создание объекта миграции")
		}
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			sugarLogger.Fatalw(err.Error(), "event", "применение миграции")
		}
	}

	if store != nil {
		defer store.Close()
	}

	observer := audit.InitObserver()
	if flags.AuditFile != "" {
		fileAuditor, err := audit.InitAuditFile(flags.AuditFile)
		if err != nil {
			sugarLogger.Fatalw(err.Error(), "event", "файл аудита не открылся")
		}
		observer.Register(fileAuditor)
		defer func(fileAuditor *audit.AuditorFile) {
			err := fileAuditor.Close()
			if err != nil {
				sugarLogger.Debugw(err.Error(), "event", "ошибка закрытия файла аудита")
			}
		}(fileAuditor)
	}

	if flags.AuditURL != "" {
		urlAuditor := audit.InitAuditURL(flags.AuditURL)
		observer.Register(urlAuditor)
	}

	ctx, cancel := context.WithCancel(context.Background())
	idleConnsClosed := make(chan struct{})

	shortenerService := service.NewShortenerService(ctx, conf, store, sugarLogger, observer)

	mux := logger.WithLogging(
		handler.AuthorizationMiddleware(
			handler.RequestDecompressHandle(
				handler.ResponseCompressHandle(
					handler.NewHandlers(ctx, conf, store, sugarLogger, observer, shortenerService),
					sugarLogger,
				),
				sugarLogger,
			),
			sugarLogger, conf.CookieName, conf.TokenSecret, db,
		),
		sugarLogger,
	)

	addr := flags.RunAddress
	if addr == "" {
		addr = ":8080"
	}

	// создание HTTP сервера
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-c

		sugarLogger.Infow("Завершение работы сервиса")
		if err := srv.Shutdown(context.Background()); err != nil {
			sugarLogger.Debugw(err.Error(), "event", "shutdown server")
		}

		cancel()
		sugarLogger.Infow("Завершение работы сервиса 2")
		close(idleConnsClosed)
	}()

	grpcAddr := flags.GRPCAddress
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1:8090"
	}

	sugarLogger.Infow("Старт gRPC сервера", "addr", grpcAddr)

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		sugarLogger.Fatalw(err.Error(), "event", "ошибка запуска gRPC")
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(ypgrpc.AuthInterceptor(sugarLogger, conf.TokenSecret, db)),
	)
	pb.RegisterShortenerServiceServer(grpcServer, ypgrpc.New(shortenerService, sugarLogger))
	reflection.Register(grpcServer)

	go func() {
		sugarLogger.Infow("запуск gRPC сервера")
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			sugarLogger.Fatalw(err.Error(), "event", "gRPC сервер не запустился")
		}
	}()

	sugarLogger.Infow("Старт сервера", "addr", addr)

	// старт сервера на адресе
	err = server.StartServer(flags.EnableHTTPS, flags.SSLCertPath, flags.SSLKeyPath, srv)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugarLogger.Infow(err.Error(), "event", "start server")
	}

	<-idleConnsClosed

	sugarLogger.Infow("HTTP сервер остановлен")
}
