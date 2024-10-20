package config

import "os"

type Config struct {
	TelegramToken string
	OpenAIKey     string
	Owner         string
	BotUsername   string
	ChromaBaseUrl string
}

func LoadConfig() *Config {
	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
		Owner:         os.Getenv("RATATOSKR_OWNER"),
		BotUsername:   os.Getenv("RATATOSKR_BOT_USERNAME"),
		ChromaBaseUrl: os.Getenv("CHROMA_BASE_URL"),
	}
}
