package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/logger"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/utils"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

var log = logger.NewLogger()

func MessageCron() *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	s.Cron("* * * * *").Do(func() {
		log.Debug("cron job started")

		repo, err := repository.GetRepository(1, false)
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

		err = api.LoadAPIConfigs("./internal/api/apis-config.yml")
		if err != nil {
			log.Fatalf("Failed to load API configurations: %v", err)
		}

		xPostReq := api.RequestConfig{
			APIName: "twitter",
			FormFields: map[string]string{
				"text": item.Text,
				"url":  item.URL,
			},
			FileFields: map[string]string{
				"image": image_name,
			},
		}

		xResp, err := api.ExecuteRequest(xPostReq)
		if err != nil {
			log.Debug("X API error: %v", err)
		} else if xResp.Success {
			log.Debug("X post created successfully!")
		}

		if _, err := repository.UpdateRepositoryPosted(item.URL, true); err != nil {
			log.Error("Error updating repository posted status: %v", err)
		}

		err = utils.RemoveAllFilesInFolder("./tmp/gh_project_img")
		if err != nil {
			log.Error(err)
		}
	})

	s.StartAsync()

	log.Debug("scheduler started successfully")
	return s
}
