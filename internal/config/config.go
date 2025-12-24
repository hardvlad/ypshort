package config

import "github.com/hardvlad/ypshort/internal/config/db"

type Config struct {
	ServerAddress   string
	ShortLinkLength int
	Charset         string
	FileName        string
	DBConfig        *db.Config
}

func NewConfig(serverAddress string, dsn string) *Config {
	if serverAddress == "" {
		serverAddress = "http://localhost:8080/"
	}

	return &Config{
		ServerAddress:   serverAddress,
		ShortLinkLength: 6,
		Charset:         "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		DBConfig:        db.NewConfig(dsn),
	}
}
