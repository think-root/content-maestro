package scheduler

import (
	"content-maestro/internal/schedule"
	"content-maestro/internal/store"

	"github.com/go-co-op/gocron"
)

func UpdateScheduler(cronName string, store *store.Store) *gocron.Scheduler {
	switch cronName {
	case "message":
		return schedule.MessageCron(store)
	case "collect":
		return schedule.CollectCron(store)
	default:
		return nil
	}
}
