package poll

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"volleyball_bot/storage"
)

// CreatePoll в игровой день создаёт новый опрос на createDaysAhead дней вперёд
// и закрепляет его, если опроса на эту дату ещё нет.
func (s *Service) CreatePoll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	todayWeekday := now.Weekday()

	log.Printf("Проверка: сегодня %s (%s)", now.Format("02.01.2006"), weekdays[todayWeekday.String()])

	if !s.gameDays[todayWeekday] {
		log.Printf("Сегодня (%s) игры нет, пропускаем", weekdays[todayWeekday.String()])
		return
	}

	// Дата игры через createDaysAhead дней в gameHour:00.
	date := gameDate(now)

	// Не создаём дубль, если опрос на эту дату уже есть.
	polls, err := s.repo.Load()
	if err != nil {
		log.Printf("Ошибка чтения хранилища: %v", err)
		return
	}
	if pollForDateExists(polls, date) {
		log.Printf("Опрос на %s уже создан, пропускаем", date.Format("02.01.2006"))
		return
	}

	log.Printf("Сегодня (%s) игра! Создаём опрос на %s...", weekdays[todayWeekday.String()], date.Format("02.01.2006"))

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

	pinned := s.pinMessage(sentPoll.MessageID)

	// Сохраняем для последующего открепления и удаления.
	if err := s.repo.Add(storage.PinnedPoll{
		MessageID: sentPoll.MessageID,
		CreatedAt: now.Unix(),
		GameDate:  date.Unix(),
		Pinned:    pinned,
	}); err != nil {
		log.Printf("Не удалось сохранить опрос: %v", err)
	}
}

// pollForDateExists сообщает, есть ли уже опрос на тот же календарный день игры.
func pollForDateExists(polls []storage.PinnedPoll, date time.Time) bool {
	for _, p := range polls {
		if p.GameDate == 0 {
			continue
		}
		g := time.Unix(p.GameDate, 0).In(date.Location())
		if g.Year() == date.Year() && g.YearDay() == date.YearDay() {
			return true
		}
	}
	return false
}
