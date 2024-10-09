package config

import "os"

type Config struct {
	TelegramToken string
	OpenAIKey     string
}

func LoadConfig() *Config {
	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
	}
}
