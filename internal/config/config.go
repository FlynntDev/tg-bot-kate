package config

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken        string
	DatabasePath    string
	ChannelUsername string
}

func LoadConfig() (*Config, error) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("не удалось найти переменную окружения BOT_TOKEN")
	}

	channelUsername := os.Getenv("CHANNEL_USERNAME")
	if channelUsername == "" {
		return nil, fmt.Errorf("не удалось найти переменную окружения CHANNEL_USERNAME")
	}

	databasePath := "database.db" // Путь к базе данных можно также получить из переменной окружения при необходимости

	return &Config{
		BotToken:        botToken,
		DatabasePath:    databasePath,
		ChannelUsername: channelUsername,
	}, nil
}
