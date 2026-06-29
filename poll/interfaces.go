package poll

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
)

// Sender абстрагирует методы Telegram-бота, нужные сервису.
// Им удовлетворяет *tgbotapi.BotAPI.
type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// Repository абстрагирует хранилище закреплённых опросов.
// Ему удовлетворяет *storage.PollStore.
type Repository interface {
	Load() ([]storage.PinnedPoll, error)
	Save(polls []storage.PinnedPoll) error
	Add(p storage.PinnedPoll) error
}
