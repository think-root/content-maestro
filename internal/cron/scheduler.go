package cron

import (
	"time"

	"github.com/go-co-op/gocron"
)

type Job func(s *gocron.Scheduler)

func NewScheduler(schedule string, job Job) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	
	if schedule != "" {
		s.Cron(schedule).Do(func() {
			job(s)
		})
	}
	
	s.StartAsync()
	time.Sleep(time.Second)
	
	return s
}
