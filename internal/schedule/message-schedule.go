package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/store"
	"content-maestro/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

func MessageJob(s *gocron.Scheduler, store store.StoreInterface) {
	log.Debug("cron job started")

	var status int
	var logMessage string

	defer func() {
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, logMessage)
			log.Error("Message job panic: %v", r)
			if err := store.LogCronExecution("message", 0, panicMessage); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			panic(r)
		}

		if err := store.LogCronExecution("message", status, logMessage); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
	}()

	apiConfigs := api.GetAPIConfigs()
	if apiConfigs == nil {
		log.Error("API configurations not loaded")
		status = 0
		logMessage = "API configurations not loaded"
		return
	}

	var image_name string

	needsImage := false
	for _, endpoint := range apiConfigs.APIs {
		if endpoint.Enabled && endpoint.SocialifyImage {
			needsImage = true
			break
		}
	}

	if needsImage {
		for _, endpoint := range apiConfigs.APIs {
			if !endpoint.Enabled || !endpoint.SocialifyImage {
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
				log.Debugf("No items found in repository for language %s", textLanguage)
				continue
			}

			item := repo.Data.Items[0]
			username_repo := strings.TrimPrefix(item.URL, "https://github.com/")
			timestamp := time.Now().UnixNano()
			imageFilename := fmt.Sprintf("image_%d.png", timestamp)
			image_name = fmt.Sprintf("./tmp/gh_project_img/%s", imageFilename)

			err = socialify.Socialify(username_repo, image_name)
			if err != nil {
				log.Error(err)
				err := utils.CopyFile("./assets/banner.jpg", image_name)
				if err != nil {
					log.Error("Failed to copy file: %v", err)
					status = 0
					logMessage = fmt.Sprintf("Failed to copy fallback banner file: %v", err)
					return
				}
			}
			break
		}
	}

	var successfulAPIs []string
	var failedAPIs []string
	var errorMessages []string
	var updatedURL string

	for apiName, endpoint := range apiConfigs.APIs {
		if !endpoint.Enabled {
			continue
		}

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
			log.Debugf("No items found in repository for %s API with language %s", apiName, textLanguage)
			failedAPIs = append(failedAPIs, apiName)
			errorMessages = append(errorMessages, fmt.Sprintf("%s API error: no items for language %s", apiName, textLanguage))
			continue
		}

		item := repo.Data.Items[0]

		for {
			statusCode, err := repository.ValidateRepositoryURL(item.URL)
			if err != nil {
				log.Error("Error validating repository URL %s: %v", item.URL, err)
				break
			}

			if statusCode == 200 {
				log.Debugf("Repository %s is valid (status %d)", item.URL, statusCode)
				break
			}

			log.Debugf("Repository %s returned status %d, deleting and getting next", item.URL, statusCode)

			if _, err := repository.DeleteRepository(item.URL); err != nil {
				log.Error("Error deleting repository %s: %v", item.URL, err)
			}

			repo, err = repository.GetRepository(1, false, "ASC", "date_added", textLanguage)
			if err != nil {
				log.Error("Error getting next repository for %s API: %v", apiName, err)
				failedAPIs = append(failedAPIs, apiName)
				errorMessages = append(errorMessages, fmt.Sprintf("%s API error: failed to get next repository: %v", apiName, err))
				break
			}

			if len(repo.Data.Items) == 0 {
				log.Debugf("No more valid repositories available for %s API", apiName)
				failedAPIs = append(failedAPIs, apiName)
				errorMessages = append(errorMessages, fmt.Sprintf("%s API error: no valid repositories available", apiName))
				break
			}

			item = repo.Data.Items[0]
		}

		if len(repo.Data.Items) == 0 {
			continue
		}

		if updatedURL == "" {
			updatedURL = item.URL
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
			}

			if endpoint.SocialifyImage && image_name != "" {
				req.FileFields = map[string]string{
					"image": image_name,
				}
			}
		case "json":
			req = api.RequestConfig{
				APIName:  apiName,
				JSONBody: map[string]any{"text": item.Text, "url": item.URL},
			}

			if endpoint.SocialifyImage && image_name != "" {
				publicURL := os.Getenv("PUBLIC_URL")
				if publicURL != "" {
					imageURL := fmt.Sprintf("%s/images/%s", publicURL, filepath.Base(image_name))
					req.JSONBody["image_url"] = imageURL
				} else {
					log.Error("PUBLIC_URL not set, cannot generate image_url for API %s", apiName)
				}
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
		} else {
			log.Errorf("%s API request failed (status %d): %s", apiName, resp.StatusCode, string(resp.Body))
			failedAPIs = append(failedAPIs, apiName)
			errorMessages = append(errorMessages, fmt.Sprintf("%s API failed (status %d)", apiName, resp.StatusCode))
		}
	}

	if len(successfulAPIs) > 0 && updatedURL != "" {
		if _, err := repository.UpdateRepositoryPosted(updatedURL, true); err != nil {
			log.Error("Error updating repository posted status: %v", err)
			status = 0
			logMessage = fmt.Sprintf("Error updating repository posted status: %v", err)
			return
		}
	}

	err := utils.RemoveAllFilesInFolder("./tmp/gh_project_img")
	if err != nil {
		log.Error(err)
		status = 0
		logMessage = fmt.Sprintf("Error cleaning up temporary files: %v", err)
		return
	}

	if len(successfulAPIs) == 0 {
		status = 0
		logMessage = fmt.Sprintf("No messages sent successfully. Errors: %s", strings.Join(errorMessages, "; "))
	} else if len(failedAPIs) > 0 {
		status = 2
		logMessage = fmt.Sprintf("Message sent to: %s. Failed: %s. Errors: %s",
			strings.Join(successfulAPIs, ", "),
			strings.Join(failedAPIs, ", "),
			strings.Join(errorMessages, "; "))
	} else {
		status = 1
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
