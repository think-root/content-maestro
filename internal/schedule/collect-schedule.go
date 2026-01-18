package schedule

import (
	"bytes"
	"content-maestro/internal/store"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
)

type generateRequest struct {
	MaxRepos           int    `json:"max_repos"`
	Resource           string `json:"resource"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
	Period             string `json:"period"`
	Language           string `json:"language"`
	UseDirectURL       bool   `json:"use_direct_url"`
	LlmProvider        string `json:"llm_provider"`
	LlmOutputLanguage  string `json:"llm_output_language"`
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

func getContentAlchemistTimeout() time.Duration {
	timeoutStr := os.Getenv("CONTENT_ALCHEMIST_TIMEOUT")
	if timeoutStr == "" {
		return 5 * time.Minute
	}

	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		log.Error("Invalid CONTENT_ALCHEMIST_TIMEOUT value: %s, using default 300 seconds", timeoutStr)
		return 5 * time.Minute
	}

	return time.Duration(timeoutSeconds) * time.Second
}

func CollectJob(s *gocron.Scheduler, store store.StoreInterface) {
	log.Debug("Collecting posts...")

	var status int
	var logMessage string

	defer func() {

		if r := recover(); r != nil {
			panicMessage := fmt.Sprintf("Panic occurred: %v. %s", r, logMessage)
			log.Error("Collect job panic: %v", r)
			if err := store.LogCronExecution("collect", 0, panicMessage); err != nil {
				log.Error("Failed to log panic execution: %v", err)
			}
			panic(r)
		}

		if err := store.LogCronExecution("collect", status, logMessage); err != nil {
			log.Error("Failed to log cron execution: %v", err)
		}
	}()


	settings, err := store.GetCollectSettings()
	if err != nil {
		log.Error("Error getting collect settings: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error getting collect settings: %v", err)
		return
	}

	promptSettings, err := store.GetPromptSettings()
	if err != nil {
		log.Error("Error getting prompt settings: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error getting prompt settings: %v", err)
		return
	}

	payload := generateRequest{
		MaxRepos:           settings.MaxRepos,
		Resource:           settings.Resource,
		Since:              settings.Since,
		SpokenLanguageCode: settings.SpokenLanguageCode,
		Period:             settings.Period,
		Language:           settings.Language,
		UseDirectURL:       promptSettings.UseDirectURL,
		LlmProvider:        promptSettings.LlmProvider,
		LlmOutputLanguage:  promptSettings.LlmOutputLanguage,
		LlmConfig: struct {
			Model       string  `json:"model"`
			Temperature float64 `json:"temperature"`
			Messages    []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}{
			Model:       promptSettings.Model,
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
		status = 0
		logMessage = fmt.Sprintf("Error marshaling request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("CONTENT_ALCHEMIST_URL")+"/think-root/api/auto-generate/", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("Error creating request: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	timeout := getContentAlchemistTimeout()
	log.Debugf("Using timeout of %v for content-alchemist API call", timeout)

	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error sending request: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("API returned HTTP error: %d %s", resp.StatusCode, resp.Status)
		status = 0
		logMessage = fmt.Sprintf("API returned HTTP error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error reading response: %v", err)
		return
	}

	log.Debugf("API response body: %s", string(body))

	var response generateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error("Error unmarshaling response: %v", err)
		status = 0
		logMessage = fmt.Sprintf("Error unmarshaling response: %v. Response body: %s", err, string(body))
		return
	}

	switch response.Status {
	case "ok":
		log.Debugf("Collected %d new repositories", len(response.Added))
		status = 1
		logMessage = fmt.Sprintf("Collected %d repositories.", len(response.Added))
		if len(response.DontAdded) > 0 {
			logMessage += fmt.Sprintf(" Already exists %d repositories.", len(response.DontAdded))
		}
	case "partial":
		log.Debugf("Partially collected %d new repositories, %d failed", len(response.Added), len(response.DontAdded))
		status = 2
		logMessage = fmt.Sprintf("Partially collected %d repositories. Failed: %d.", len(response.Added), len(response.DontAdded))
		if response.ErrorMessage != "" {
			logMessage += fmt.Sprintf(" Error: %s", response.ErrorMessage)
		}
	case "error":
		log.Error("API returned error status: %s", response.ErrorMessage)
		status = 0
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
		status = 0
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
