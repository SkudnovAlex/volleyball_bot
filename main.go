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

	c := scheduler.New(cfg.Location)
	if err := scheduler.Add(c, cfg.Schedule, svc.CreatePoll); err != nil {
		log.Panic("Ошибка настройки задания создания:", err)
	}
	if err := scheduler.Add(c, cfg.CleanupSchedule, svc.CleanupPolls); err != nil {
		log.Panic("Ошибка настройки задания очистки:", err)
	}
	c.Start()
	log.Printf("Планировщик запущен (создание: %q, очистка: %q, зона: %s)",
		cfg.Schedule, cfg.CleanupSchedule, cfg.Location)

	scheduler.WaitForShutdown(c)
}
