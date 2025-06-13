package schedule

import (
	"bytes"
	"content-maestro/internal/store"
	"encoding/json"
	"fmt"
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
	UseDirectURL       bool   `json:"use_direct_url"`
	LlmProvider        string `json:"llm_provider"`
	LlmConfig          struct {
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	} `json:"llm_config"`
}

type generateResponse struct {
	Status    string   `json:"status"`
	Added     []string `json:"added"`
	DontAdded []string `json:"dont_added"`
}

func CollectJob(s *gocron.Scheduler, store store.StoreInterface) {
	log.Debug("Collecting posts...")

	settings, err := store.GetCollectSettings()
	if err != nil {
		log.Error("Error getting collect settings: %v", err)
		store.LogCronExecution("collect", false, err.Error())
		return
	}

	promptSettings, err := store.GetPromptSettings()
	if err != nil {
		log.Error("Error getting prompt settings: %v", err)
		store.LogCronExecution("collect", false, err.Error())
		return
	}

	payload := generateRequest{
		MaxRepos:           settings.MaxRepos,
		Since:              settings.Since,
		SpokenLanguageCode: settings.SpokenLanguageCode,
		UseDirectURL:       promptSettings.UseDirectURL,
		LlmProvider:        promptSettings.LlmProvider,
		LlmConfig: struct {
			Model       string  `json:"model"`
			Temperature float64 `json:"temperature"`
			Messages    []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}{
			Model:       "openai/gpt-4o-mini-search-preview",
			Temperature: promptSettings.Temperature,
			Messages: []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{
					Role:    "system",
					Content: promptSettings.Content,
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error marshaling request: %v", err)
		store.LogCronExecution("collect", false, err.Error())
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("CONTENT_ALCHEMIST_URL")+"/think-root/api/auto-generate/", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Error creating request: %v", err)
		store.LogCronExecution("collect", false, err.Error())
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
		store.LogCronExecution("collect", false, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		store.LogCronExecution("collect", false, err.Error())
		return
	}

	log.Debugf("API response body: %s", string(body))

	var response generateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		store.LogCronExecution("collect", false, fmt.Sprintf("Error unmarshaling response: %v. Response body: %s", err, string(body)))
		return
	}

	if response.Status == "ok" {
		log.Debugf("Successfully collected %d new repositories", len(response.Added))
		msg := fmt.Sprintf("Collected repos: %d", len(response.Added))
		store.LogCronExecution("collect", true, msg)
	} else {
		store.LogCronExecution("collect", false, "API returned non-ok status")
	}
}

func CollectCron(store store.StoreInterface) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	setting, err := store.GetCronSetting("collect")
	if err != nil || setting == nil || !setting.IsActive {
		log.Debug("Collect cron is disabled")
		return s
	}

	log.Debugf("Collect cron is enabled with schedule: %s", setting.Schedule)
	s.Cron(setting.Schedule).Do(CollectJob, s, store)
	s.StartAsync()
	log.Debug("Scheduler started successfully for collect cron")
	return s
}
