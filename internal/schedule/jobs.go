package schedule

import (
	"content-maestro/internal/models"
	"content-maestro/internal/store"
	"time"

	"github.com/go-co-op/gocron"
)

func InitJobs(store store.StoreInterface) models.JobRegistry {
	return models.JobRegistry{
		"collect": func(s *gocron.Scheduler) {
			CollectJob(s, store)
		},
		"message": func(s *gocron.Scheduler) {
			MessageJob(s, store)
		},
	}
}

func NewScheduler(store store.StoreInterface, name string, defaultSchedule string) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	setting, err := store.GetCronSetting(name)
	if err != nil || setting == nil {
		log.Debugf("%s cron setting not found in database, using default schedule", name)
		if job, exists := InitJobs(store)[name]; exists {
			s.Cron(defaultSchedule).Do(job, s)
			s.StartAsync()
			log.Debugf("Scheduler started for %s with default schedule: %s", name, defaultSchedule)
		}
		return s
	}

	if !setting.IsActive {
		log.Debugf("%s cron is disabled in database", name)
		return s
	}

	if job, exists := InitJobs(store)[name]; exists && setting.Schedule != "" {
		s.Cron(setting.Schedule).Do(job, s)
		s.StartAsync()
		log.Debugf("Scheduler started for %s with schedule: %s", name, setting.Schedule)
	}

	return s
}
