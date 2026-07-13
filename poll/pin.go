package poll

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
)

// pinMessage закрепляет сообщение опроса. Возвращает true при успехе.
func (s *Service) pinMessage(messageID int) bool {
	pin := tgbotapi.PinChatMessageConfig{
		ChatID:    s.chatID,
		MessageID: messageID,
	}
	if _, err := s.sender.Request(pin); err != nil {
		log.Printf("Ошибка закрепления опроса: %v", err)
		return false
	}
	log.Printf("Опрос %d закреплён!", messageID)
	return true
}

// unpinMessage открепляет сообщение опроса. Возвращает true при успехе.
func (s *Service) unpinMessage(p storage.PinnedPoll) bool {
	unpin := tgbotapi.UnpinChatMessageConfig{
		ChatID:    s.chatID,
		MessageID: p.MessageID,
	}
	if _, err := s.sender.Request(unpin); err != nil {
		log.Printf("Не удалось открепить опрос %d: %v", p.MessageID, err)
		return false
	}
	log.Printf("Опрос %d откреплён", p.MessageID)
	return true
}
