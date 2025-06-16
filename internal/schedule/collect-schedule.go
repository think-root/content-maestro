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
	Status       string   `json:"status"`
	Added        []string `json:"added"`
	DontAdded    []string `json:"dont_added"`
	ErrorMessage string   `json:"error_message"`
}

func CollectJob(s *gocron.Scheduler, store store.StoreInterface) {
	log.Debug("Collecting posts...")
	
	var success bool
	var logMessage string
	startTime := time.Now()
	
	defer func() {
		duration := time.Since(startTime)
		finalMessage := fmt.Sprintf("%s (duration: %v)", logMessage, duration)
		
		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, finalMessage)
			log.Error("Collect job panic: %v", r)
			if err := store.LogCronExecution("collect", false, panicMessage); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			panic(r)
		}
		
		if err := store.LogCronExecution("collect", success, finalMessage); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
	}()

	settings, err := store.GetCollectSettings()
	if err != nil {
		log.Error("Error getting collect settings: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error getting collect settings: %v", err)
		return
	}

	promptSettings, err := store.GetPromptSettings()
	if err != nil {
		log.Error("Error getting prompt settings: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error getting prompt settings: %v", err)
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
		success = false
		logMessage = fmt.Sprintf("Error marshaling request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("CONTENT_ALCHEMIST_URL")+"/think-root/api/auto-generate/", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Error creating request: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("API returned HTTP error: %d %s", resp.StatusCode, resp.Status)
		success = false
		logMessage = fmt.Sprintf("API returned HTTP error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error reading response: %v", err)
		return
	}

	log.Debugf("API response body: %s", string(body))

	var response generateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		success = false
		logMessage = fmt.Sprintf("Error unmarshaling response: %v. Response body: %s", err, string(body))
		return
	}

	switch response.Status {
	case "ok":
		log.Debugf("Successfully collected %d new repositories", len(response.Added))
		success = true
		logMessage = fmt.Sprintf("Successfully collected %d repositories. Added: %v", len(response.Added), response.Added)
		if len(response.DontAdded) > 0 {
			logMessage += fmt.Sprintf(". Skipped: %v", response.DontAdded)
		}
	case "error":
		log.Error("API returned error status: %s", response.ErrorMessage)
		success = false
		if response.ErrorMessage != "" {
			logMessage = fmt.Sprintf("API error: %s", response.ErrorMessage)
		} else {
			logMessage = "API returned error status without error message"
		}
		if len(response.DontAdded) > 0 {
			logMessage += fmt.Sprintf(". Failed repositories: %v", response.DontAdded)
		}
	default:
		log.Error("API returned unknown status: %s", response.Status)
		success = false
		logMessage = fmt.Sprintf("API returned unknown status: %s", response.Status)
		if response.ErrorMessage != "" {
			logMessage += fmt.Sprintf(". Error message: %s", response.ErrorMessage)
		}
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
