package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/config"
	"volleyball_bot/poll"
	"volleyball_bot/scheduler"
	"volleyball_bot/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic("Ошибка создания бота:", err)
	}
	bot.Debug = false
	log.Printf("Бот @%s запущен!", bot.Self.UserName)

	store := storage.NewPollStore(cfg.StorePath)
	svc := poll.New(bot, store, cfg.ChatID, cfg.Location, cfg.GameDays)

	c, err := scheduler.Start(cfg.Schedule, cfg.Location, svc.CheckAndCreatePoll)
	if err != nil {
		log.Panic("Ошибка настройки cron:", err)
	}
	log.Printf("Планировщик запущен (расписание: %q, зона: %s)", cfg.Schedule, cfg.Location)

	scheduler.WaitForShutdown(c)
}
