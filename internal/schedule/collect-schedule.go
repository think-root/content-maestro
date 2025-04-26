package schedule

import (
	"bytes"
	"content-maestro/internal/store"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

type generateRequest struct {
	MaxRepos           int    `json:"max_repos"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
}

type generateResponse struct {
	Status    string   `json:"status"`
	Added     []string `json:"added"`
	DontAdded []string `json:"dont_added"`
}

func CollectJob(s *gocron.Scheduler) {
	log.Debug("Collecting posts...")

	store, err := store.NewStore("./data/badger")
	if err != nil {
		log.Error("Error connecting to store: %v", err)
		return
	}
	defer store.Close()

	settings, err := store.GetCollectSettings()
	if err != nil {
		log.Error("Error getting collect settings: %v", err)
		return
	}

	payload := generateRequest{
		MaxRepos:           settings.MaxRepos,
		Since:              settings.Since,
		SpokenLanguageCode: settings.SpokenLanguageCode,
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

	var response generateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		return
	}

	if response.Status == "ok" {
		log.Debugf("Successfully collected %d new repositories", len(response.Added))
	}
}

func CollectCron(store *store.Store) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	setting, err := store.GetCronSetting("collect")
	if err != nil || setting == nil || !setting.IsActive {
		log.Debug("Collect cron is disabled")
		return s
	}

	s.Cron(setting.Schedule).Do(CollectJob, s)
	s.StartAsync()
	log.Debug("scheduler started successfully")
	return s
}
