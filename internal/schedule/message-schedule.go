package schedule

import (
	"content-maestro/internal/logger"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/store"
	"content-maestro/internal/utils"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var log = logger.NewLogger()

func MessageJob(s *gocron.Scheduler) {
	log.Debug("cron job started")

	repo, err := repository.GetRepository(1, false, "ASC", "date_added")
	if err != nil {
		log.Error("Error getting repository: %v", err)
		return
	}

	item := repo.Data.Items[0]
	username_repo := strings.TrimPrefix(item.URL, "https://github.com/")
	image_name := "./tmp/gh_project_img/image.png"

	err = socialify.Socialify(username_repo)
	if err != nil {
		log.Error(err)
		err := utils.CopyFile("./assets/banner.jpg", image_name)
		if err != nil {
			log.Error("Failed to copy file: %v", err)
			return
		}
	}

	if _, err := repository.UpdateRepositoryPosted(item.URL, true); err != nil {
		log.Error("Error updating repository posted status: %v", err)
	}

	err = utils.RemoveAllFilesInFolder("./tmp/gh_project_img")
	if err != nil {
		log.Error(err)
	}
}

func MessageCron(store *store.Store) *gocron.Scheduler {
	setting, err := store.GetCronSetting("message")
	if err != nil || setting == nil || !setting.IsActive {
		log.Debug("Message cron is disabled")
		return gocron.NewScheduler(time.UTC)
	}

	s := gocron.NewScheduler(time.UTC)
	s.Cron(setting.Schedule).Do(MessageJob, s)
	s.StartAsync()
	log.Debug("scheduler started successfully")
	return s
}
