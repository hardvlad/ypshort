package main

import (
	"flag"
	"os"
	"strings"
)

type programFlags struct {
	RunAddress    string
	ServerAddress string
	FileName      string
	Dsn           string
}

func parseFlags() programFlags {

	var flags programFlags

	flag.StringVar(&flags.RunAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	if envRunAddr, ok := os.LookupEnv("BASE_URL"); ok {
		flags.RunAddress = envRunAddr
	}

	flag.StringVar(&flags.ServerAddress, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	if envServAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		flags.ServerAddress = envServAddr
	}

	flag.StringVar(&flags.FileName, "f", "shortener_db.json", "файл данных сервиса")
	if envFileName, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		flags.FileName = envFileName
	}

	flag.StringVar(&flags.Dsn, "d", "", "строка подключения к базе данных")
	if envDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		flags.Dsn = envDsn
	}

	flag.Parse()

	if !strings.HasSuffix(flags.ServerAddress, "/") {
		flags.ServerAddress += "/"
	}

	return flags
}
