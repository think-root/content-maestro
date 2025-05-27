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
	UseDirectURL       bool   `json:"use_direct_url"`
	LlmProvider        string `json:"llm_provider"`
	LlmConfig          struct {
		Model     string `json:"model"`
		Temperature float64 `json:"temperature"`
		Messages []struct {
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

func CollectJob(s *gocron.Scheduler, store *store.Store) {
	log.Debug("Collecting posts...")

	settings, err := store.GetCollectSettings()
	if err != nil {
		log.Error("Error getting collect settings: %v", err)
		return
	}

	payload := generateRequest{
		MaxRepos:           settings.MaxRepos,
		Since:              settings.Since,
		SpokenLanguageCode: settings.SpokenLanguageCode,
		UseDirectURL:       true,
		LlmProvider:        "openrouter",
		LlmConfig: struct {
			Model     string `json:"model"`
			Temperature float64 `json:"temperature"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}{
			Model:     "openai/gpt-4.1-mini:online",
			Temperature: 0.2,
			Messages: []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{
					Role:    "system",
					Content: "Ти слухняний і корисний помічник, який суворо дотримується всіх наведених нижче вимог. Твоя основна задача — створювати короткі описи GitHub репозиторіїв українською мовою на основі наданих URL.\n\nЦі URL ведуть на GitHub-репозиторії. Під час створення опису обов’язково дотримуйся наступних правил:\n\n1. Включай не більше трьох ключових функцій репозиторію.\n2. Не додавай жодних посилань у тексті.\n3. Пиши простою, зрозумілою мовою, без переліків. Інформацію про функції вплітай у зв’язний текст.\n4. Не згадуй сумісність, платформи, авторів, компанії або колаборації.\n5. Не використовуй жодної розмітки: ні HTML, ні Markdown.\n6. Опис має бути лаконічним і точним, як один твіт — не більше 280 символів з урахуванням пробілів.\n7. Технічні терміни (назви мов програмування, бібліотек, інструментів, команд тощо) залишай англійською мовою.\n8. Перед генерацією переконайся, що текст відповідає всім цим вимогам.\n\nПісля цього тобі буде надано URL GitHub-репозиторію. Твоє завдання — ознайомитися з його вмістом і створити короткий, чіткий і зрозумілий опис, що повністю відповідає цим правилам.",
				},
			},
		},
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

	s.Cron(setting.Schedule).Do(CollectJob, s, store)
	s.StartAsync()
	log.Debug("scheduler started successfully")
	return s
}
