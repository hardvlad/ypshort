package config

import "github.com/hardvlad/ypshort/internal/config/db"

type Config struct {
	ServerAddress   string
	ShortLinkLength int
	Charset         string
	FileName        string
	DBConfig        *db.Config
	CookieName      string
	TokenSecret     string
}

func NewConfig(serverAddress string, dsn string, length int) *Config {
	if serverAddress == "" {
		serverAddress = "http://localhost:8080/"
	}

	return &Config{
		ServerAddress:   serverAddress,
		ShortLinkLength: length,
		Charset:         "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		DBConfig:        db.NewConfig(dsn),
		CookieName:      "yp_short_token",
		TokenSecret:     "superSecretKey",
	}
}
