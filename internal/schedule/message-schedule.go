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
	
	defer func() {
		
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, logMessage)
			log.Error("Message job panic: %v", r)
			if err := store.LogCronExecution("message", false, panicMessage); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			panic(r)
		}
		
		if err := store.LogCronExecution("message", success, logMessage); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
	}()

	err := api.LoadAPIConfigs("./internal/api/apis-config.yml")
	if err != nil {
		log.Error("Failed to load API configurations: %v", err)
		success = false
		logMessage = fmt.Sprintf("Failed to load API configurations: %v", err)
		return
	}

	// Створюємо зображення один раз для всіх API
	// Для цього отримуємо перший доступний репозиторій
	var image_name string
	
	for _, endpoint := range api.GetAPIConfigs().APIs {
		if !endpoint.Enabled {
			continue
		}
		
		textLanguage := endpoint.TextLanguage
		if textLanguage == "" {
			textLanguage = "en"
		}
		
		repo, err := repository.GetRepository(1, false, "ASC", "date_added", textLanguage)
		if err != nil {
			log.Error("Error getting repository for language %s: %v", textLanguage, err)
			continue
		}
		
		if len(repo.Data.Items) == 0 {
			log.Debug("No items found in repository for language %s", textLanguage)
			continue
		}
		
		item := repo.Data.Items[0]
		username_repo := strings.TrimPrefix(item.URL, "https://github.com/")
		image_name = "./tmp/gh_project_img/image.png"
		
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
		break
	}
	
	if image_name == "" {
		log.Debug("No items found in repository for any language")
		success = false
		logMessage = "No items found in repository for any language"
		return
	}

	var successfulAPIs []string
	var failedAPIs []string
	var errorMessages []string
	var updatedURL string

	for apiName, endpoint := range api.GetAPIConfigs().APIs {
		if !endpoint.Enabled {
			continue
		}

		// Отримуємо текст для конкретної мови цього API
		textLanguage := endpoint.TextLanguage
		if textLanguage == "" {
			textLanguage = "en"
		}
		
		repo, err := repository.GetRepository(1, false, "ASC", "date_added", textLanguage)
		if err != nil {
			log.Error("Error getting repository for %s API with language %s: %v", apiName, textLanguage, err)
			failedAPIs = append(failedAPIs, apiName)
			errorMessages = append(errorMessages, fmt.Sprintf("%s API error (language %s): %v", apiName, textLanguage, err))
			continue
		}
		
		if len(repo.Data.Items) == 0 {
			log.Debug("No items found in repository for %s API with language %s", apiName, textLanguage)
			failedAPIs = append(failedAPIs, apiName)
			errorMessages = append(errorMessages, fmt.Sprintf("%s API error: no items for language %s", apiName, textLanguage))
			continue
		}
		
		item := repo.Data.Items[0]
		if updatedURL == "" {
			updatedURL = item.URL // Зберігаємо URL для оновлення статусу
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
			log.Debugf("%s post created successfully with language %s!", apiName, textLanguage)
			successfulAPIs = append(successfulAPIs, apiName)
		}
	}

	if len(successfulAPIs) > 0 && updatedURL != "" {
		if _, err := repository.UpdateRepositoryPosted(updatedURL, true); err != nil {
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
		logMessage = fmt.Sprintf("No messages sent successfully. Errors: %s", strings.Join(errorMessages, "; "))
	} else if len(failedAPIs) > 0 {
		success = true
		logMessage = fmt.Sprintf("Message sent to: %s. Failed: %s. Errors: %s",
			strings.Join(successfulAPIs, ", "),
			strings.Join(failedAPIs, ", "),
			strings.Join(errorMessages, "; "))
	} else {
		success = true
		logMessage = fmt.Sprintf("Message sent successfully to: %s",
			strings.Join(successfulAPIs, ", "))
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
