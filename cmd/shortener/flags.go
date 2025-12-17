package main

import (
	"flag"
)

type programFlags struct {
	RunAddress    string
	ServerAddress string
}

func parseFlags() programFlags {

	var flags programFlags

	flag.StringVar(&flags.RunAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&flags.ServerAddress, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")

	flag.Parse()

	return flags
}
