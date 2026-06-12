package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerAddress       string
	DatabaseURL         string
	CityGarbageURL      string
	TelegramSecretToken string
	TelegramBotToken    string
}

func LoadConfig() Config {
	return Config{
		ServerAddress:       getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		CityGarbageURL:      getEnv("CITY_GARBAGE_URL", ""),
		TelegramSecretToken: getEnv("TELEGRAM_SECRET_TOKEN", ""),
		TelegramBotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
	}
}

func (c Config) Validate() error {
	required := map[string]string{
		"SERVER_ADDRESS":        c.ServerAddress,
		"DATABASE_URL":          c.DatabaseURL,
		"CITY_GARBAGE_URL":      c.CityGarbageURL,
		"TELEGRAM_SECRET_TOKEN": c.TelegramSecretToken,
		"TELEGRAM_BOT_TOKEN":    c.TelegramBotToken,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("environment variable %s is not set or is empty", key)
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
