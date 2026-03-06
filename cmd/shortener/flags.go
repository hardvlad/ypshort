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
	SSLCertPath   string
	SSLKeyPath    string
	TrustedSubnet string
	GRPCAddress   string
}

type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     string `json:"enable_https"`
	SSLCertificate  string `json:"ssl_certificate"`
	SSLPrivateKey   string `json:"ssl_private_key"`
	TrustedSubnet   string `json:"trusted_subnet"`
	GRPCAddress     string `json:"grpc_address"`
}

func parseFlags() programFlags {

	var flags programFlags

	flag.StringVar(&flags.RunAddress, "a", ":8080", "адрес запуска HTTP-сервера")
	if envRunAddr, ok := os.LookupEnv("BASE_URL"); ok {
		flags.RunAddress = envRunAddr
	}

	flag.StringVar(&flags.GRPCAddress, "g", ":8080", "адрес запуска GRPC сервера")
	if envGRPCAddress, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		flags.GRPCAddress = envGRPCAddress
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

	flag.StringVar(&flags.SSLCertPath, "sc", "", "путь к HTTPS сертификату")
	if sslCert, ok := os.LookupEnv("HTTPS_CERT"); ok {
		flags.SSLCertPath = sslCert
	}

	flag.StringVar(&flags.SSLKeyPath, "sk", "", "путь к HTTPS приватному ключу")
	if sslKey, ok := os.LookupEnv("HTTPS_KEY"); ok {
		flags.SSLKeyPath = sslKey
	}

	flag.StringVar(&flags.TrustedSubnet, "t", "", "доверенная подсеть")
	if trustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		flags.TrustedSubnet = trustedSubnet
	}

	jsonConfigFile := ""
	flag.StringVar(&jsonConfigFile, "c", "", "имя файла конфигурации в формате JSON")
	if envJSONConfig, ok := os.LookupEnv("CONFIG"); ok {
		jsonConfigFile = envJSONConfig
	}

	flag.Parse()

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

					if config.BaseURL != "" && flags.ServerAddress == "" {
						flags.ServerAddress = config.BaseURL
					}

					if config.FileStoragePath != "" && flags.FileName == "" {
						flags.FileName = config.FileStoragePath
					}

					if config.DatabaseDsn != "" && flags.Dsn == "" {
						flags.Dsn = config.DatabaseDsn
					}

					if config.EnableHTTPS != "" && !flags.EnableHTTPS {
						flags.EnableHTTPS, _ = strconv.ParseBool(config.EnableHTTPS)
					}

					if config.SSLCertificate != "" && flags.SSLCertPath == "" {
						flags.SSLCertPath = config.SSLCertificate
					}

					if config.SSLPrivateKey != "" && flags.SSLKeyPath == "" {
						flags.SSLKeyPath = config.SSLPrivateKey
					}

					if config.TrustedSubnet != "" && flags.TrustedSubnet == "" {
						flags.TrustedSubnet = config.TrustedSubnet
					}

					if config.GRPCAddress != "" && flags.GRPCAddress == "" {
						flags.GRPCAddress = config.GRPCAddress
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
