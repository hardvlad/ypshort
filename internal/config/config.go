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

func NewConfig(serverAddress string) *Config {
	return &Config{
		ServerAddress:   serverAddress,
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
