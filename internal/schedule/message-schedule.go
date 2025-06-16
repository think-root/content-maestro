package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/store"
	"content-maestro/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)


func MessageJob(s *gocron.Scheduler, store store.StoreInterface) {
	log.Debug("cron job started")
	
	var success bool
	var logMessage string
	startTime := time.Now()
	
	defer func() {
		duration := time.Since(startTime)
		finalMessage := fmt.Sprintf("%s (duration: %v)", logMessage, duration)
		
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, finalMessage)
			log.Error("Message job panic: %v", r)
			if err := store.LogCronExecution("message", false, panicMessage); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			panic(r)
		}
		
		if err := store.LogCronExecution("message", success, finalMessage); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
	}()

	repo, err := repository.GetRepository(1, false, "ASC", "date_added")
	if err != nil {
		log.Error("Error getting repository: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error getting repository: %v", err)
		return
	}

	if len(repo.Data.Items) == 0 {
		log.Debug("No items found in repository")
		success = false
		logMessage = "No items found in repository"
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
			success = false
			logMessage = fmt.Sprintf("Failed to copy fallback banner file: %v", err)
			return
		}
	}

	err = api.LoadAPIConfigs("./internal/api/apis-config.yml")
	if err != nil {
		log.Error("Failed to load API configurations: %v", err)
		success = false
		logMessage = fmt.Sprintf("Failed to load API configurations: %v", err)
		return
	}

	var successfulAPIs []string
	var failedAPIs []string
	var errorMessages []string

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
				JSONBody: map[string]any{"text": item.Text, "url": item.URL},
			}
		default:
			req = api.RequestConfig{
				APIName:  apiName,
				JSONBody: map[string]any{"text": item.Text, "url": item.URL},
			}
		}

		resp, err := api.ExecuteRequest(req)
		if err != nil {
			log.Errorf("%s API error: %v", apiName, err)
			failedAPIs = append(failedAPIs, apiName)
			errorMessages = append(errorMessages, fmt.Sprintf("%s API error: %v", apiName, err))
		} else if resp.Success {
			log.Debugf("%s post created successfully!", apiName)
			successfulAPIs = append(successfulAPIs, apiName)
		}
	}

	if len(successfulAPIs) > 0 {
		if _, err := repository.UpdateRepositoryPosted(item.URL, true); err != nil {
			log.Error("Error updating repository posted status: %v", err)
			success = false
			logMessage = fmt.Sprintf("Error updating repository posted status: %v", err)
			return
		}
	}

	err = utils.RemoveAllFilesInFolder("./tmp/gh_project_img")
	if err != nil {
		log.Error(err)
		success = false
		logMessage = fmt.Sprintf("Error cleaning up temporary files: %v", err)
		return
	}

	if len(successfulAPIs) == 0 {
		success = false
		logMessage = fmt.Sprintf("No messages sent successfully. Repository: %s. Errors: %s", item.URL, strings.Join(errorMessages, "; "))
	} else if len(failedAPIs) > 0 {
		success = true
		logMessage = fmt.Sprintf("Message sent to: %s. Failed: %s. Repository: %s. Errors: %s",
			strings.Join(successfulAPIs, ", "),
			strings.Join(failedAPIs, ", "),
			item.URL,
			strings.Join(errorMessages, "; "))
	} else {
		success = true
		logMessage = fmt.Sprintf("Message sent successfully to: %s. Repository: %s",
			strings.Join(successfulAPIs, ", "),
			item.URL)
	}
}

func MessageCron(store store.StoreInterface) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	
	setting, err := store.GetCronSetting("message")
	if err != nil || setting == nil || !setting.IsActive {
		log.Debug("Message cron is disabled")
		return s
	}

	log.Debugf("Message cron is enabled with schedule: %s", setting.Schedule)
	s.Cron(setting.Schedule).Do(MessageJob, s, store)
	s.StartAsync()
	log.Debug("scheduler started successfully")
	return s
}
