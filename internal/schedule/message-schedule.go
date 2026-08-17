package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/models"
	"content-maestro/internal/notification"
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

	var status int
	var logMessage string

	var successfulAPIs []string
	var failedAPIs []string
	var errorMessages []string
	var updatedURL string

	// Assembled at exit rather than at the end of the publishing loop, so an
	// early return - or a panic - still records which item the run consumed and
	// which connectors missed it. That is the data a manual retry needs, and it
	// matters most exactly when the run died mid-publish.
	runDetails := func() *models.MessageRunDetails {
		if updatedURL == "" && len(successfulAPIs) == 0 && len(failedAPIs) == 0 {
			return nil
		}
		return &models.MessageRunDetails{
			URL:    updatedURL,
			Sent:   successfulAPIs,
			Failed: failedAPIs,
		}
	}

	defer func() {
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, logMessage)
			log.Error("Message job panic: %v", r)
			if err := store.LogCronExecutionDetails("message", 0, panicMessage, runDetails()); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			notification.NotifyCronResult("message", 0, panicMessage)
			panic(r)
		}

		if err := store.LogCronExecutionDetails("message", status, logMessage, runDetails()); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
		notification.NotifyCronResult("message", status, logMessage)
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

			repo, err := repository.GetRepository(1, false, "ASC", "publication_queue", textLanguage)
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
			image_name = fmt.Sprintf("%s/%s", imageDir, imageFilename)

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

	for apiName, endpoint := range apiConfigs.APIs {
		if !endpoint.Enabled {
			continue
		}

		textLanguage := endpoint.TextLanguage
		if textLanguage == "" {
			textLanguage = "en"
		}

		repo, err := repository.GetRepository(1, false, "ASC", "publication_queue", textLanguage)
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

		// Repositories whose URL no longer resolves are dropped and the next
		// candidate is fetched. The loop reports its outcome through this flag:
		// reading the fetch response afterwards would dereference nil once a
		// re-fetch fails, which crashes the job and, through the deferred
		// re-panic, the whole process.
		itemAvailable := true

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

			nextRepo, err := repository.GetRepository(1, false, "ASC", "publication_queue", textLanguage)
			if err != nil {
				log.Error("Error getting next repository for %s API: %v", apiName, err)
				failedAPIs = append(failedAPIs, apiName)
				errorMessages = append(errorMessages, fmt.Sprintf("%s API error: failed to get next repository: %v", apiName, err))
				itemAvailable = false
				break
			}

			if len(nextRepo.Data.Items) == 0 {
				log.Debugf("No more valid repositories available for %s API", apiName)
				failedAPIs = append(failedAPIs, apiName)
				errorMessages = append(errorMessages, fmt.Sprintf("%s API error: no valid repositories available", apiName))
				itemAvailable = false
				break
			}

			item = nextRepo.Data.Items[0]
		}

		if !itemAvailable {
			continue
		}

		if updatedURL == "" {
			updatedURL = item.URL
		}

		resp, err := publishItem(apiName, endpoint, item, image_name)
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

	err := utils.RemoveAllFilesInFolder(imageDir)
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
