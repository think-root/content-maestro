package schedule

import (
	"content-maestro/internal/store"

	"github.com/go-co-op/gocron"
)

func UpdateScheduler(cronName string, store *store.Store) *gocron.Scheduler {
	switch cronName {
	case "message":
		return MessageCron(store)
	case "collect":
		return CollectCron(store)
	default:
		return nil
	}
}
