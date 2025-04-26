package schedule

import (
	"content-maestro/internal/models"
	"content-maestro/internal/store"
	"time"

	"github.com/go-co-op/gocron"
)

func InitJobs(store *store.Store) models.JobRegistry {
	return models.JobRegistry{
		"collect": func(s *gocron.Scheduler) {
			CollectJob(s)
		},
		"message": MessageJob,
	}
}

func NewScheduler(store *store.Store, name string, schedule string) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	if job, exists := InitJobs(store)[name]; exists && schedule != "" {
		s.Cron(schedule).Do(job, s)
		s.StartAsync()
	}

	return s
}
