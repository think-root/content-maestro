package server

import (
	apiExecutor "content-maestro/internal/api"
	"content-maestro/internal/models"
	"content-maestro/internal/schedule"
	"content-maestro/internal/store"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// publishStore is a StoreInterface stub that only serves API configs and records
// what a run wrote to the cron history. Every other method is a hard error, so a
// handler reaching for one fails loudly instead of silently reading a zero value.
type publishStore struct {
	configs  []models.APIConfigModel
	logCalls int
}

func (s *publishStore) GetAllAPIConfigs() ([]models.APIConfigModel, error) {
	return s.configs, nil
}

func (s *publishStore) LogCronExecutionDetails(string, int, string, *models.MessageRunDetails) error {
	s.logCalls++
	return nil
}

func (s *publishStore) LogCronExecution(name string, status int, output string) error {
	return s.LogCronExecutionDetails(name, status, output, nil)
}

func (s *publishStore) Close() error                     { return nil }
func (s *publishStore) InitializeDefaultSettings() error { return nil }
func (s *publishStore) GetCronSetting(string) (*models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) GetAllCronSettings() ([]models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) UpdateCronSetting(string, string, bool) (*models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) GetCronHistoryCount(string, *int, *time.Time, *time.Time) (int, error) {
	return 0, errors.New("not implemented")
}
func (s *publishStore) GetCronHistory(string, *int, int, int, string, *time.Time, *time.Time) ([]models.CronHistory, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) GetCollectSettings() (*store.CollectSettings, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) UpdateCollectSettings(*store.CollectSettings) error {
	return errors.New("not implemented")
}
func (s *publishStore) GetPromptSettings() (*models.PromptSettings, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) UpdatePromptSettings(*models.UpdatePromptSettingsRequest) error {
	return errors.New("not implemented")
}
func (s *publishStore) GetAPIConfig(string) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) CreateAPIConfig(*models.CreateAPIConfigRequest) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) UpdateAPIConfig(string, *models.UpdateAPIConfigRequest) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *publishStore) DeleteAPIConfig(string) error { return errors.New("not implemented") }

var _ store.StoreInterface = (*publishStore)(nil)

const publishTestURL = "https://github.com/resemble-ai/chatterbox"

// The handler is called directly rather than through the middleware chain, so
// these tests cover the request validation and the error-to-status mapping only.
func newPublishAPI(t *testing.T, configs []models.APIConfigModel) (*CronAPI, *publishStore) {
	t.Helper()

	st := &publishStore{configs: configs}
	if err := apiExecutor.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	return NewCronAPI(st, nil, nil), st
}

func publishRequest(method, body string) *http.Request {
	return httptest.NewRequest(method, "/api/message/publish", strings.NewReader(body))
}

func TestPublishMessageNowRejectsBadRequests(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{name: "wrong method", method: http.MethodGet, body: "", wantStatus: http.StatusMethodNotAllowed},
		{name: "malformed body", method: http.MethodPost, body: "{", wantStatus: http.StatusBadRequest},
		{name: "missing url", method: http.MethodPost, body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "blank url", method: http.MethodPost, body: `{"url":"   "}`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api, st := newPublishAPI(t, nil)

			recorder := httptest.NewRecorder()
			api.PublishMessageNow(recorder, publishRequest(tt.method, tt.body))

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body %q)", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if st.logCalls != 0 {
				t.Errorf("cron history writes = %d, want none for a rejected request", st.logCalls)
			}
		})
	}
}

// Nothing about the request is wrong when no integration is enabled - the server
// simply cannot honour it - so it answers 409 rather than 400.
func TestPublishMessageNowConflictsWhenNoIntegrationEnabled(t *testing.T) {
	api, _ := newPublishAPI(t, []models.APIConfigModel{{
		Name: "threads", URL: "http://127.0.0.1:1", Method: http.MethodPost,
		ContentType: "json", SuccessCode: http.StatusOK, Enabled: false,
	}})

	recorder := httptest.NewRecorder()
	api.PublishMessageNow(recorder, publishRequest(http.MethodPost, `{"url":"`+publishTestURL+`"}`))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d (body %q)", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), schedule.ErrNoEnabledIntegrations.Error()) {
		t.Errorf("body = %q, want it to name the missing integrations", recorder.Body.String())
	}
}

// A partially failed run is still a run that happened: the per-integration
// outcomes carry the failures, so the response stays 200.
func TestPublishMessageNowAnswersOKWithOutcomes(t *testing.T) {
	connector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer connector.Close()

	alchemist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			w.Write([]byte(`{"status":"ok","message":"updated"}`))
			return
		}

		var body struct {
			URL          string `json:"url"`
			TextLanguage string `json:"text_language"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"items": []map[string]any{{
					"id": 1327, "posted": false, "url": body.URL, "text": "text",
				}},
			},
		})
	}))
	defer alchemist.Close()
	t.Setenv("CONTENT_ALCHEMIST_URL", alchemist.URL)
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	api, st := newPublishAPI(t, []models.APIConfigModel{
		{
			Name: "threads", URL: connector.URL, Method: http.MethodPost,
			ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
			TextLanguage: "en",
		},
		{
			Name: "twitter", URL: "http://127.0.0.1:1", Method: http.MethodPost,
			ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
			TextLanguage: "en",
		},
	})

	recorder := httptest.NewRecorder()
	api.PublishMessageNow(recorder, publishRequest(http.MethodPost, `{"url":"`+publishTestURL+`"}`))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content type = %q, want application/json", got)
	}

	var result schedule.PublishResult
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode the response: %v", err)
	}
	if result.Status != 2 {
		t.Errorf("status = %d, want 2 (partial)", result.Status)
	}
	if len(result.Outcomes) != 2 {
		t.Errorf("outcomes = %+v, want one per integration", result.Outcomes)
	}
	if !result.Posted {
		t.Errorf("posted = false, want the item marked as published")
	}
	if st.logCalls != 1 {
		t.Errorf("cron history writes = %d, want 1", st.logCalls)
	}
}
