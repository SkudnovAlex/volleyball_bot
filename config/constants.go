package config

const (
	defaultStorePath = "polls.json"
	// defaultSchedule — расписание создания опроса (ежедневно в 20:00).
	defaultSchedule = "0 20 * * *"
	// defaultCleanupSchedule — расписание очистки (ежедневно в 06:00).
	defaultCleanupSchedule = "0 6 * * *"
	defaultTimezone        = "Europe/Moscow"
	// defaultGameDays — игровые дни по умолчанию (англ. названия через запятую).
	defaultGameDays = "Tuesday,Thursday,Friday,Sunday"
)
