package poll

import (
	"fmt"
	"time"
)

// gameDate возвращает дату игры через daysAhead дней в gameHour:00.
func gameDate(now time.Time) time.Time {
	d := now.AddDate(0, 0, daysAhead)
	return time.Date(d.Year(), d.Month(), d.Day(), gameHour, 0, 0, 0, d.Location())
}

// getQuestion формирует текст опроса.
func getQuestion(date time.Time) string {
	dayName := weekdays[date.Weekday().String()]
	dateStr := date.Format("02.01")
	timeStr := date.Format("15:04")

	return fmt.Sprintf("🏐 %s %s в %s!\nКто идёт?", dayName, dateStr, timeStr)
}
