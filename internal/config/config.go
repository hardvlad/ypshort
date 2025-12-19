package config

type Config struct {
	ServerAddress   string
	ShortLinkLength int
	Charset         string
}

func NewConfig(serverAddress string) *Config {
	if serverAddress == "" {
		serverAddress = "http://localhost:8080/"
	}

	return &Config{
		ServerAddress:   serverAddress,
		ShortLinkLength: 6,
		Charset:         "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}
