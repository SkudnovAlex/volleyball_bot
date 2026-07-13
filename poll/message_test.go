package poll

import (
	"strings"
	"testing"
	"time"
)

func TestGameDate(t *testing.T) {
	now := time.Date(2026, time.June, 30, 20, 30, 0, 0, time.UTC) // вторник
	got := gameDate(now)

	want := time.Date(2026, time.July, 5, gameHour, 0, 0, 0, time.UTC) // +5 дней, 19:00
	if !got.Equal(want) {
		t.Fatalf("gameDate = %v, ожидалось %v", got, want)
	}
}

func TestGetQuestion(t *testing.T) {
	date := time.Date(2026, time.July, 7, gameHour, 0, 0, 0, time.UTC) // вторник
	q := getQuestion(date)

	for _, sub := range []string{"Вторник", "07.07", "19:00"} {
		if !strings.Contains(q, sub) {
			t.Errorf("в тексте опроса отсутствует %q: %s", sub, q)
		}
	}
}
