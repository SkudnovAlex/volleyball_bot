package poll

const (
	// gameHour — час начала игры (для текста опроса).
	gameHour = 19
	// createDaysAhead — на сколько дней вперёд создаётся опрос.
	createDaysAhead = 7
	// deleteAfterDays — опросы возрастом >= этого числа дней удаляются.
	deleteAfterDays = 7
)

var pollOptions = []string{
	"✅ Да, я в игре",
	"👤 Я приведу друга(+1)",
	"🚶‍♂️🚶‍♂️ Я приведу друзей(+2)",
	"❌ Нет, я не смогу",
	"🤔 Возможно, пока не могу сказать",
}

var weekdays = map[string]string{
	"Monday":    "Понедельник",
	"Tuesday":   "Вторник",
	"Wednesday": "Среда",
	"Thursday":  "Четверг",
	"Friday":    "Пятница",
	"Saturday":  "Суббота",
	"Sunday":    "Воскресенье",
}
