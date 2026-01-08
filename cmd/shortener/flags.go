package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type programFlags struct {
	RunAddress    string
	ServerAddress string
	FileName      string
	Length        int
	Dsn           string
	AuditFile     string
	AuditURL      string
}

func parseFlags() programFlags {

	var flags programFlags

	flag.StringVar(&flags.RunAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	if envRunAddr, ok := os.LookupEnv("BASE_URL"); ok {
		flags.RunAddress = envRunAddr
	}

	flag.IntVar(&flags.Length, "l", 6, "длина сокращённой части URL")
	if envLength, ok := os.LookupEnv("SHORT_LENGTH"); ok {
		var err error
		flags.Length, err = strconv.Atoi(envLength)
		if err != nil {
			flags.Length = 6
		}
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

	flag.StringVar(&flags.AuditFile, "audit-file", "", "путь к файлу-приёмнику, в который сохраняются логи аудита")
	if envAuditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		flags.AuditFile = envAuditFile
	}

	flag.StringVar(&flags.AuditURL, "audit-url", "", "полный URL удаленного сервера-приёмника, куда отправляются логи аудита")
	if envAuditURL, ok := os.LookupEnv("AUDIT_FILE"); ok {
		flags.AuditURL = envAuditURL
	}

	flag.Parse()

	if !strings.HasSuffix(flags.ServerAddress, "/") {
		flags.ServerAddress += "/"
	}

	return flags
}
