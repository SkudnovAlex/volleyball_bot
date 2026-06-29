package storage

import "sync"

// PinnedPoll хранит информацию о созданном закреплённом опросе.
type PinnedPoll struct {
	MessageID int   `json:"message_id"`
	CreatedAt int64 `json:"created_at"` // unix-время создания
}

// PollStore — потокобезопасное JSON-хранилище закреплённых опросов.
type PollStore struct {
	path string
	mu   sync.Mutex
}
