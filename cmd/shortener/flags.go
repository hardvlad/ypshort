package main

import (
	"flag"
)

var runAddress string
var serverAddress string

func parseFlags() {

	flag.StringVar(&runAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&serverAddress, "b", "http://localhost:8080/", "базовый адрес результирующего сокращённого URL")

	flag.Parse()
}
