package storage

import "sync"

// PinnedPoll хранит информацию о созданном опросе.
type PinnedPoll struct {
	MessageID int   `json:"message_id"`
	CreatedAt int64 `json:"created_at"` // unix-время создания
	GameDate  int64 `json:"game_date"`  // unix-время игры, на которую создан опрос
	Pinned    bool  `json:"pinned"`     // сообщение сейчас закреплено
}

// PollStore — потокобезопасное JSON-хранилище закреплённых опросов.
type PollStore struct {
	path string
	mu   sync.Mutex
}
