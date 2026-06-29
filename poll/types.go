package poll

import (
	"sync"
	"time"
)

// Service содержит бизнес-логику создания и удаления опросов.
type Service struct {
	sender   Sender
	repo     Repository
	chatID   int64
	gameDays map[time.Weekday]bool
	now      func() time.Time
	mu       sync.Mutex
}
