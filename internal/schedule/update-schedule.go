package schedule

import (
	"content-maestro/internal/store"

	"github.com/go-co-op/gocron"
)

func UpdateScheduler(cronName string, store store.StoreInterface) *gocron.Scheduler {
	switch cronName {
	case "message":
		return NewScheduler(store, cronName, "12 12 * * *")
	case "collect":
		return NewScheduler(store, cronName, "13 13 * * 6")
	default:
		return nil
	}
}
