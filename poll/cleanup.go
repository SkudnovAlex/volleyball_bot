package poll

import (
	"log"
	"time"

	"volleyball_bot/storage"
)

// CleanupPolls открепляет опросы, день игры которых прошёл, и удаляет опросы
// возрастом >= deleteAfterDays дней. Файл хранилища затем перезаписывается.
func (s *Service) CleanupPolls() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	polls, err := s.repo.Load()
	if err != nil {
		log.Printf("Ошибка чтения хранилища: %v", err)
		return
	}
	if len(polls) == 0 {
		log.Println("Нет сохранённых опросов для очистки")
		return
	}

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	deleteThreshold := now.AddDate(0, 0, -deleteAfterDays).Unix()

	remaining := make([]storage.PinnedPoll, 0, len(polls))

	for _, p := range polls {
		// 1. Удаление: возраст >= deleteAfterDays дней.
		if p.CreatedAt <= deleteThreshold {
			if s.deleteMessage(p) {
				continue
			}
			// Не удалось удалить (нет прав) — оставляем для повторной попытки.
			remaining = append(remaining, p)
			continue
		}

		// 2. Открепление: день игры уже прошёл, а опрос ещё закреплён.
		// Записи без GameDate не открепляем, их уберёт удаление по возрасту.
		game := time.Unix(p.GameDate, 0).In(now.Location())
		if p.Pinned && p.GameDate != 0 && game.Before(startOfToday) {
			if s.unpinMessage(p) {
				p.Pinned = false
			}
		}
		remaining = append(remaining, p)
	}

	if err := s.repo.Save(remaining); err != nil {
		log.Printf("Не удалось обновить хранилище: %v", err)
	}
}
