package poll

import (
	"errors"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func New(sender Sender, repo Repository, chatID int64, loc *time.Location, gameDays map[time.Weekday]bool) *Service {
	return &Service{
		sender:   sender,
		repo:     repo,
		chatID:   chatID,
		gameDays: gameDays,
		now:      func() time.Time { return time.Now().In(loc) },
	}
}

// isInsufficientRights определяет, что ошибка вызвана нехваткой прав бота.
func isInsufficientRights(err error) bool {
	var tgErr *tgbotapi.Error
	if errors.As(err, &tgErr) {
		return tgErr.Code == 400 && strings.Contains(tgErr.Message, "not enough rights")
	}
	return false
}
