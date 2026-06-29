package scheduler

import (
	"testing"
	"time"
)

func TestStart_InvalidSchedule(t *testing.T) {
	c, err := Start("not a cron", time.UTC, func() {})
	if err == nil {
		t.Fatal("ожидалась ошибка для некорректного расписания")
	}
	if c != nil {
		c.Stop()
	}
}

func TestStart_ValidSchedule(t *testing.T) {
	c, err := Start("* * * * *", time.UTC, func() {})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if c == nil {
		t.Fatal("cron не должен быть nil при валидном расписании")
	}
	if len(c.Entries()) != 1 {
		t.Errorf("ожидалась 1 задача, получено: %d", len(c.Entries()))
	}
	c.Stop()
}
