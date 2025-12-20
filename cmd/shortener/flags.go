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
}

func parseFlags() programFlags {

	var flags programFlags

	flag.StringVar(&flags.RunAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
		flags.RunAddress = envRunAddr
	}

	flag.StringVar(&flags.ServerAddress, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")
	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		flags.ServerAddress = envServAddr
	}

	flag.StringVar(&flags.FileName, "f", "shortener_db.json", "файл данных сервиса")
	if envFileName := os.Getenv("FILE_STORAGE_PATH"); envFileName != "" {
		flags.FileName = envFileName
	}

	flag.Parse()

	if !strings.HasSuffix(flags.ServerAddress, "/") {
		flags.ServerAddress += "/"
	}

	return flags
}
