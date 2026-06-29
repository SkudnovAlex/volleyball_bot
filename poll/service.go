package poll

import (
	"errors"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
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

// CheckAndCreatePoll проверяет, игровой ли сегодня день, и если да —
// удаляет старые опросы и создаёт новый на неделю вперёд.
func (s *Service) CheckAndCreatePoll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	todayWeekday := now.Weekday()

	log.Printf("Проверка: сегодня %s (%s)", now.Format("02.01.2006"), weekdays[todayWeekday.String()])

	if !s.gameDays[todayWeekday] {
		log.Printf("Сегодня (%s) игры нет, пропускаем", weekdays[todayWeekday.String()])
		return
	}

	log.Printf("Сегодня (%s) игра! Создаём опрос на следующую неделю...", weekdays[todayWeekday.String()])

	// 1. Удаляем все опросы возрастом >= daysAhead дней.
	s.deleteOldPolls(now)

	// 2. Дата игры через неделю в gameHour:00.
	date := gameDate(now)

	// 3. Отправляем новый опрос.
	question := getQuestion(date)
	pollConfig := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: s.chatID,
		},
		Question:              question,
		Options:               pollOptions,
		IsAnonymous:           false,
		AllowsMultipleAnswers: true,
	}

	sentPoll, err := s.sender.Send(pollConfig)
	if err != nil {
		log.Printf("Ошибка отправки опроса: %v", err)
		return
	}

	log.Printf("Опрос создан! ID: %d на %s", sentPoll.MessageID, date.Format("02.01.2006"))

	// 4. Закрепляем новый опрос.
	pinConfig := tgbotapi.PinChatMessageConfig{
		ChatID:    s.chatID,
		MessageID: sentPoll.MessageID,
	}
	if _, err := s.sender.Request(pinConfig); err != nil {
		log.Printf("Ошибка закрепления опроса: %v", err)
	} else {
		log.Printf("Опрос %d закреплён!", sentPoll.MessageID)
	}

	// 5. Сохраняем в хранилище для последующего удаления.
	if err := s.repo.Add(storage.PinnedPoll{MessageID: sentPoll.MessageID, CreatedAt: now.Unix()}); err != nil {
		log.Printf("Не удалось сохранить опрос: %v", err)
	}
}

// deleteOldPolls читает хранилище и удаляет все опросы,
// возраст которых >= daysAhead дней. Файл затем перезаписывается.
func (s *Service) deleteOldPolls(now time.Time) {
	polls, err := s.repo.Load()
	if err != nil {
		log.Printf("Ошибка чтения хранилища: %v", err)
		return
	}
	if len(polls) == 0 {
		log.Println("Нет сохранённых опросов для проверки")
		return
	}

	thresholdUnix := now.AddDate(0, 0, -daysAhead).Unix()
	remaining := make([]storage.PinnedPoll, 0, len(polls))

	for _, p := range polls {
		// Опрос ещё актуален (моложе порога) — оставляем.
		if p.CreatedAt > thresholdUnix {
			remaining = append(remaining, p)
			continue
		}

		createdStr := time.Unix(p.CreatedAt, 0).Format("02.01.2006")
		log.Printf("Старый опрос ID %d от %s — удаляем", p.MessageID, createdStr)

		// Удаляем сообщение: для закреплённого это автоматически снимает закреп.
		del := tgbotapi.DeleteMessageConfig{
			ChatID:    s.chatID,
			MessageID: p.MessageID,
		}
		if _, err := s.sender.Request(del); err != nil {
			if isInsufficientRights(err) {
				log.Printf("Нет прав на удаление сообщения %d — оставляю запись для повторной попытки", p.MessageID)
				remaining = append(remaining, p)
				continue
			}
			// Иное (например, сообщение уже удалено) — убираем из хранилища.
			log.Printf("Ошибка удаления опроса %d (запись будет удалена): %v", p.MessageID, err)
			continue
		}

		log.Printf("Опрос %d от %s удалён!", p.MessageID, createdStr)
	}

	if err := s.repo.Save(remaining); err != nil {
		log.Printf("Не удалось обновить хранилище: %v", err)
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
