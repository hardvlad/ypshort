package main

import (
	"flag"
	"os"
	"strings"
)

type programFlags struct {
	RunAddress    string
	ServerAddress string
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

	if !strings.HasSuffix(flags.ServerAddress, "/") {
		flags.ServerAddress += "/"
	}

	flag.Parse()

	return flags
}
