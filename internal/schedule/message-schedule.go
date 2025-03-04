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
	s.Cron("10 12 * * *").Do(func() {
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

		for apiName, endpoint := range api.GetAPIConfigs().APIs { 
			if !endpoint.Enabled {
				continue
			}

			var req api.RequestConfig

			commonFields := map[string]string{
				"text": item.Text,
				"url":  item.URL,
			}

			switch strings.ToLower(endpoint.ContentType) {
			case "multipart":
				req = api.RequestConfig{
					APIName:    apiName,
					FormFields: commonFields,
					FileFields: map[string]string{
						"image": image_name,
					},
				}
			case "json":
				req = api.RequestConfig{
					APIName:  apiName,
					JSONBody: map[string]interface{}{"text": item.Text, "url": item.URL},
				}
			default:
				req = api.RequestConfig{
					APIName:  apiName,
					JSONBody: map[string]interface{}{"text": item.Text, "url": item.URL},
				}
			}

			resp, err := api.ExecuteRequest(req)
			if err != nil {
				log.Debug("%s API error: %v", apiName, err)
			} else if resp.Success {
				log.Debug("%s post created successfully!", apiName)
			}
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
