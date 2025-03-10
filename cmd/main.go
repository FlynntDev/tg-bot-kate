package main

import (
	"log"

	"github.com/FlynntDev/tg-bot-kate/internal/bot"
	"github.com/FlynntDev/tg-bot-kate/internal/config"
	"github.com/FlynntDev/tg-bot-kate/internal/repository"
	"github.com/FlynntDev/tg-bot-kate/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Ошибка загрузки .env файла: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	repo, err := repository.NewRepository(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	srv := service.NewService(repo, cfg.BotToken, cfg.ChannelUsername)
	b, err := bot.NewBot(cfg.BotToken, srv)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	b.Start()
}
