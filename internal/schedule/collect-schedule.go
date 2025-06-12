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

	payload := generateRequest{
		MaxRepos:           settings.MaxRepos,
		Since:              settings.Since,
		SpokenLanguageCode: settings.SpokenLanguageCode,
		UseDirectURL:       true,
		LlmProvider:        "openrouter",
		LlmConfig: struct {
			Model       string  `json:"model"`
			Temperature float64 `json:"temperature"`
			Messages    []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}{
			Model:       "openai/gpt-4o-mini-search-preview",
			Temperature: 0.2,
			Messages: []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{
					Role:    "system",
					Content: "Ти — AI асистент, що спеціалізується на створенні коротких описів GitHub-репозиторіїв українською мовою. Твоя відповідь **ПОВИННА** суворо відповідати **КОЖНІЙ** з наведених нижче вимог. Будь-яке відхилення, особливо щодо довжини тексту, є неприпустимим. Твоя основна задача — створювати описи на основі наданих URL.\n\nПід час створення опису **НЕУХИЛЬНО** дотримуйся наступних правил:\n\n1.  Включай не більше однієї ключової функції репозиторію.\n2.  **ЗАБОРОНЕНО** додавати будь-які посилання.\n3.  Пиши простою, зрозумілою мовою, без переліків. Інформацію про функції вплітай у зв’язний текст.\n4.  **ЗАБОРОНЕНО** згадувати сумісність, платформи, авторів, компанії або колаборації.\n5.  **ЗАБОРОНЕНО** використовувати будь-яку розмітку: ні HTML, ні Markdown.\n6.  Опис має бути **НАДЗВИЧАЙНО** лаконічним. **АБСОЛЮТНИЙ МАКСИМУМ — 275 символів**, враховуючи пробіли. **Це найважливіша вимога! Перевищення ліміту є КРИТИЧНОЮ помилкою.**\n7.  Технічні терміни (назви мов програмування, бібліотек, інструментів, команд тощо) залишай англійською мовою.\n8.  **ПЕРЕД НАДАННЯМ ВІДПОВІДІ:** Переконайся, що текст відповідає **ВСІМ** вимогам. **ОБОВ'ЯЗКОВО ПЕРЕВІР** довжину. Якщо вона перевищує 270 символів, **ПЕРЕПИШИ І СКОРОТИ** його, доки він не буде відповідати ліміту.\n\nТобі буде надано URL GitHub-репозиторію. Ознайомся з ним і згенеруй опис, що **ТОЧНО** відповідає цим інструкціям.",
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
