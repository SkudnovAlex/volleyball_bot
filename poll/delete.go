package poll

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
)

// deleteMessage удаляет сообщение опроса. Возвращает true, если запись
// следует убрать из хранилища (удалено успешно или сообщения уже нет).
func (s *Service) deleteMessage(p storage.PinnedPoll) bool {
	createdStr := time.Unix(p.CreatedAt, 0).Format("02.01.2006")
	log.Printf("Старый опрос ID %d от %s — удаляем", p.MessageID, createdStr)

	del := tgbotapi.DeleteMessageConfig{
		ChatID:    s.chatID,
		MessageID: p.MessageID,
	}
	if _, err := s.sender.Request(del); err != nil {
		if isInsufficientRights(err) {
			log.Printf("Нет прав на удаление сообщения %d — оставляю запись для повторной попытки", p.MessageID)
			return false
		}
		// Иное (например, сообщение уже удалено) — убираем из хранилища.
		log.Printf("Ошибка удаления опроса %d (запись будет удалена): %v", p.MessageID, err)
		return true
	}

	log.Printf("Опрос %d от %s удалён!", p.MessageID, createdStr)
	return true
}
