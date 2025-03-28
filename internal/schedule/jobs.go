package schedule

import (
	"content-maestro/internal/models"
	"content-maestro/internal/store"
	"time"

	"github.com/go-co-op/gocron"
)

func InitJobs() models.JobRegistry {
	return models.JobRegistry{
		"collect": CollectJob,
		"message": MessageJob,
	}
}

func NewScheduler(store *store.Store, name string, schedule string) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	if job, exists := InitJobs()[name]; exists && schedule != "" {
		s.Cron(schedule).Do(job, s)
		s.StartAsync()
	}

	return s
}
