package cron

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"content-maestro/internal/logger"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/store"
	"content-maestro/internal/utils"

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

	req, err := http.NewRequest("POST", os.Getenv("CONTENT_ALCHEMIST_URL")+"/think-root/api/social-media/", nil)
	if err != nil {
		log.Error("Error creating request: %v", err)
		return
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		return
	}

	var response struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		return
	}

	if response.Status == "ok" {
		log.Debug("Message sent successfully")
	}

	if _, err := repository.UpdateRepositoryPosted(item.URL, true); err != nil {
		log.Error("Error updating repository posted status: %v", err)
	}

	err = utils.RemoveAllFilesInFolder("./tmp/gh_project_img")
	if err != nil {
		log.Error(err)
	}
}

func CollectJob(s *gocron.Scheduler) {
	log.Debug("Collecting posts...")

	payload := struct {
		MaxRepos           int    `json:"max_repos"`
		Since              string `json:"since"`
		SpokenLanguageCode string `json:"spoken_language_code"`
	}{
		MaxRepos:           5,
		Since:              "daily",
		SpokenLanguageCode: "en",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error marshaling request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("CONTENT_ALCHEMIST_URL")+"/think-root/api/auto-generate/", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Error creating request: %v", err)
		return
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		return
	}

	var response struct {
		Status    string   `json:"status"`
		Added     []string `json:"added"`
		DontAdded []string `json:"dont_added"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		return
	}

	if response.Status == "ok" {
		log.Debugf("Successfully collected %d new repositories", len(response.Added))
	}
}

func InitScheduler(store *store.Store, cronName string) *gocron.Scheduler {
	setting, err := store.GetCronSetting(cronName)
	if err != nil {
		log.Error("Error getting cron setting: %v", err)
		return gocron.NewScheduler(time.UTC)
	}

	if setting == nil || !setting.IsActive {
		log.Debug("%s cron is disabled", cronName)
		return gocron.NewScheduler(time.UTC)
	}

	var job Job
	switch cronName {
	case "message":
		job = MessageJob
	case "collect":
		job = CollectJob
	default:
		return gocron.NewScheduler(time.UTC)
	}

	s := NewScheduler(setting.Schedule, job)
	log.Debug("scheduler started successfully")
	return s
}
