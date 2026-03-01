package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

// generate:reset
type programFlags struct {
	RunAddress    string
	ServerAddress string
	FileName      string
	Length        int
	Dsn           string
	AuditFile     string
	AuditURL      string
	EnableHTTPS   bool
}

type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseUrl         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     string `json:"enable_https"`
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

	flag.BoolVar(&flags.EnableHTTPS, "s", false, "запустить сервер с поддержкой HTTPS")
	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		flags.EnableHTTPS = true
	}

	flag.Parse()

	jsonConfigFile := ""
	flag.StringVar(&jsonConfigFile, "c", "", "имя файла конфигурации в формате JSON")
	if envJSONConfig, ok := os.LookupEnv("CONFIG"); ok {
		jsonConfigFile = envJSONConfig
	}

	if jsonConfigFile != "" {
		var config jsonConfig

		_, err := os.Stat(jsonConfigFile)

		if err == nil || !errors.Is(err, os.ErrNotExist) {
			file, err := os.OpenFile(jsonConfigFile, os.O_RDONLY, 0x666)
			if err == nil {
				defer file.Close()
				if err := json.NewDecoder(file).Decode(&config); err == nil {
					if config.ServerAddress != "" && flags.RunAddress == "" {
						flags.RunAddress = config.ServerAddress
					}

					if config.BaseUrl != "" && flags.ServerAddress == "" {
						flags.ServerAddress = config.BaseUrl
					}

					if config.FileStoragePath != "" && flags.FileName == "" {
						flags.FileName = config.FileStoragePath
					}

					if config.DatabaseDsn != "" && flags.Dsn == "" {
						flags.Dsn = config.DatabaseDsn
					}

					if config.EnableHTTPS != "" && flags.EnableHTTPS == false {
						flags.EnableHTTPS, _ = strconv.ParseBool(config.EnableHTTPS)
					}
				}
			}
		}
	}

	if !strings.HasSuffix(flags.ServerAddress, "/") {
		flags.ServerAddress += "/"
	}

	return flags
}
