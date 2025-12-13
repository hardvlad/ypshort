package config

import (
	"math/rand"
	"strings"
)

type Config struct {
	ServerAddress   string
	ShortLinkLength int
	Charset         string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:   "http://localhost:8080/",
		ShortLinkLength: 6,
		Charset:         "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}

func (conf *Config) GenerateRandomString() string {
	sb := strings.Builder{}
	sb.Grow(conf.ShortLinkLength)
	for i := 0; i < conf.ShortLinkLength; i++ {
		sb.WriteByte(conf.Charset[rand.Intn(len(conf.Charset))])
	}
	return sb.String()
}
