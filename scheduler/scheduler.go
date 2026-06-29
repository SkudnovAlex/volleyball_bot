package scheduler

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

// Start настраивает и запускает cron с заданным расписанием, таймзоной и задачей.
func Start(schedule string, loc *time.Location, job func()) (*cron.Cron, error) {
	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc(schedule, job); err != nil {
		return nil, err
	}
	c.Start()
	return c, nil
}

// WaitForShutdown блокируется до SIGINT/SIGTERM и корректно останавливает cron.
func WaitForShutdown(c *cron.Cron) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Получен сигнал завершения. Бот останавливается...")

	ctx := c.Stop()
	<-ctx.Done()

	log.Println("Бот остановлен")
}
