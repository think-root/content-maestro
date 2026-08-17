package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/models"
	"content-maestro/internal/store"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// retryStore is a StoreInterface stub that only serves API configs and captures
// what the retry recorded in the cron history.
type retryStore struct {
	configs []models.APIConfigModel

	loggedName    string
	loggedStatus  int
	loggedOutput  string
	loggedDetails *models.MessageRunDetails
	logCalls      int
}

func (s *retryStore) GetAllAPIConfigs() ([]models.APIConfigModel, error) {
	return s.configs, nil
}

func (s *retryStore) LogCronExecutionDetails(name string, status int, output string, details *models.MessageRunDetails) error {
	s.logCalls++
	s.loggedName = name
	s.loggedStatus = status
	s.loggedOutput = output
	s.loggedDetails = details
	return nil
}

func (s *retryStore) LogCronExecution(name string, status int, output string) error {
	return s.LogCronExecutionDetails(name, status, output, nil)
}

func (s *retryStore) Close() error                     { return nil }
func (s *retryStore) InitializeDefaultSettings() error { return nil }
func (s *retryStore) GetCronSetting(string) (*models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) GetAllCronSettings() ([]models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) UpdateCronSetting(string, string, bool) (*models.CronSetting, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) GetCronHistoryCount(string, *int, *time.Time, *time.Time) (int, error) {
	return 0, errors.New("not implemented")
}
func (s *retryStore) GetCronHistory(string, *int, int, int, string, *time.Time, *time.Time) ([]models.CronHistory, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) GetCollectSettings() (*store.CollectSettings, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) UpdateCollectSettings(*store.CollectSettings) error {
	return errors.New("not implemented")
}
func (s *retryStore) GetPromptSettings() (*models.PromptSettings, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) UpdatePromptSettings(*models.UpdatePromptSettingsRequest) error {
	return errors.New("not implemented")
}
func (s *retryStore) GetAPIConfig(string) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) CreateAPIConfig(*models.CreateAPIConfigRequest) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) UpdateAPIConfig(string, *models.UpdateAPIConfigRequest) (*models.APIConfigModel, error) {
	return nil, errors.New("not implemented")
}
func (s *retryStore) DeleteAPIConfig(string) error { return errors.New("not implemented") }

var _ store.StoreInterface = (*retryStore)(nil)

const retryTestURL = "https://github.com/resemble-ai/chatterbox"

// alchemistStub answers get-repository for a single known repository.
func alchemistStub(t *testing.T, posted bool, calls *[]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			URL          string `json:"url"`
			Posted       *bool  `json:"posted"`
			TextLanguage string `json:"text_language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode alchemist request: %v", err)
		}
		if calls != nil {
			*calls = append(*calls, body.TextLanguage)
		}

		url := body.URL
		if url == "" {
			// The "latest published" lookup, used when no url is supplied.
			url = retryTestURL
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"items": []map[string]any{{
					"id":     1327,
					"posted": posted,
					"url":    url,
					"text":   "text for " + body.TextLanguage,
				}},
			},
		})
	}))
}

// withRepositoryEndpoints points the repository package at the stub server.
func withRepositoryEndpoints(t *testing.T, alchemistURL string) {
	t.Helper()

	t.Setenv("CONTENT_ALCHEMIST_URL", alchemistURL)
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")
}

func TestRetryMessagePostSendsOnlyRequestedAPIs(t *testing.T) {
	var connectorBodies []map[string]any
	connector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		connectorBodies = append(connectorBodies, body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer connector.Close()

	var languages []string
	alchemist := alchemistStub(t, true, &languages)
	defer alchemist.Close()
	withRepositoryEndpoints(t, alchemist.URL)

	st := &retryStore{configs: []models.APIConfigModel{
		{
			Name: "threads", URL: connector.URL, Method: http.MethodPost,
			ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
			TextLanguage: "en",
		},
		{
			Name: "bluesky", URL: connector.URL, Method: http.MethodPost,
			ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
			TextLanguage: "uk",
		},
	}}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	result, err := RetryMessagePost(st, []string{"threads", "threads", " "}, retryTestURL)
	if err != nil {
		t.Fatalf("RetryMessagePost() error = %v", err)
	}

	if len(connectorBodies) != 1 {
		t.Fatalf("connector received %d requests, want 1", len(connectorBodies))
	}
	if got := connectorBodies[0]["url"]; got != retryTestURL {
		t.Errorf("published url = %v, want %v", got, retryTestURL)
	}
	if got := connectorBodies[0]["text"]; got != "text for en" {
		t.Errorf("published text = %v, want the English text", got)
	}
	if len(languages) != 1 || languages[0] != "en" {
		t.Errorf("alchemist was asked for languages %v, want [en]", languages)
	}

	if result.Status != 1 {
		t.Errorf("status = %d, want 1", result.Status)
	}
	if len(result.Succeeded) != 1 || result.Succeeded[0] != "threads" {
		t.Errorf("succeeded = %v, want [threads]", result.Succeeded)
	}
	if len(result.Failed) != 0 {
		t.Errorf("failed = %v, want none", result.Failed)
	}

	if st.logCalls != 1 {
		t.Fatalf("cron history writes = %d, want 1", st.logCalls)
	}
	if st.loggedName != "message" {
		t.Errorf("history name = %q, want %q", st.loggedName, "message")
	}
	if !strings.HasPrefix(st.loggedOutput, "Manual retry:") {
		t.Errorf("history output = %q, want a Manual retry prefix", st.loggedOutput)
	}
	if st.loggedDetails == nil || !st.loggedDetails.Manual {
		t.Fatalf("history details = %+v, want manual details", st.loggedDetails)
	}
	if st.loggedDetails.URL != retryTestURL {
		t.Errorf("history details url = %q, want %q", st.loggedDetails.URL, retryTestURL)
	}
}

func TestRetryMessagePostReportsPerAPIFailures(t *testing.T) {
	alchemist := alchemistStub(t, true, nil)
	defer alchemist.Close()
	withRepositoryEndpoints(t, alchemist.URL)

	st := &retryStore{configs: []models.APIConfigModel{
		{
			Name: "bluesky", URL: "http://127.0.0.1:1", Method: http.MethodPost,
			ContentType: "json", SuccessCode: http.StatusOK, Enabled: false,
			TextLanguage: "en",
		},
	}}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	result, err := RetryMessagePost(st, []string{"bluesky", "telegram"}, retryTestURL)
	if err != nil {
		t.Fatalf("RetryMessagePost() error = %v", err)
	}

	if result.Status != 0 {
		t.Errorf("status = %d, want 0", result.Status)
	}
	if len(result.Failed) != 2 {
		t.Fatalf("failed = %v, want both APIs", result.Failed)
	}

	reasons := map[string]string{}
	for _, outcome := range result.Outcomes {
		reasons[outcome.APIName] = outcome.Error
	}
	if !strings.Contains(reasons["bluesky"], "disabled") {
		t.Errorf("bluesky outcome = %q, want a disabled reason", reasons["bluesky"])
	}
	if !strings.Contains(reasons["telegram"], "not configured") {
		t.Errorf("telegram outcome = %q, want a not-configured reason", reasons["telegram"])
	}
}

func TestRetryMessagePostRejectsEmptyAPIList(t *testing.T) {
	st := &retryStore{}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	_, err := RetryMessagePost(st, []string{"  ", ""}, retryTestURL)
	if !errors.Is(err, ErrInvalidRetryRequest) {
		t.Fatalf("error = %v, want ErrInvalidRetryRequest", err)
	}
	if st.logCalls != 0 {
		t.Errorf("cron history writes = %d, want none for a rejected request", st.logCalls)
	}
}

// An unposted item must only leave the queue once every requested connector has
// it: marking it posted after a partial retry re-creates the failure this
// endpoint exists to repair.
func TestRetryMessagePostMarksPostedOnlyWhenComplete(t *testing.T) {
	tests := []struct {
		name       string
		apis       []string
		wantPosted bool
	}{
		{name: "every requested integration succeeds", apis: []string{"threads"}, wantPosted: true},
		{name: "one integration still fails", apis: []string{"threads", "mastodon"}, wantPosted: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			}))
			defer connector.Close()

			var postedPatches int
			alchemist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPatch {
					postedPatches++
					w.Write([]byte(`{"status":"ok","message":"updated"}`))
					return
				}

				var body struct {
					URL          string `json:"url"`
					TextLanguage string `json:"text_language"`
				}
				json.NewDecoder(r.Body).Decode(&body)

				url := body.URL
				if url == "" {
					url = retryTestURL
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"status": "ok",
					"data": map[string]any{
						"items": []map[string]any{{
							"id": 1327, "posted": false, "url": url, "text": "text",
						}},
					},
				})
			}))
			defer alchemist.Close()
			withRepositoryEndpoints(t, alchemist.URL)

			st := &retryStore{configs: []models.APIConfigModel{{
				Name: "threads", URL: connector.URL, Method: http.MethodPost,
				ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
				TextLanguage: "en",
			}}}
			if err := api.LoadAPIConfigs(st); err != nil {
				t.Fatalf("LoadAPIConfigs() error = %v", err)
			}

			if _, err := RetryMessagePost(st, tt.apis, retryTestURL); err != nil {
				t.Fatalf("RetryMessagePost() error = %v", err)
			}

			if tt.wantPosted && postedPatches != 1 {
				t.Errorf("update-posted calls = %d, want 1", postedPatches)
			}
			if !tt.wantPosted && postedPatches != 0 {
				t.Errorf("update-posted calls = %d, want none while a connector still failed", postedPatches)
			}
		})
	}
}

// A retry keeps its image in a subdirectory of the image root, so the served
// path has to carry that subdirectory or the connector fetches a 404.
func TestImageURLPath(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{name: "cron image at the root", image: imageDir + "/image_123.png", want: "image_123.png"},
		{name: "retry image in a subdirectory", image: retryImageDir + "/retry_123.png", want: "retry/retry_123.png"},
		{name: "path outside the image root falls back to the file name", image: "/var/tmp/other.png", want: "other.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageURLPath(tt.image); got != tt.want {
				t.Errorf("imageURLPath(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}

func TestNormalizeAPINames(t *testing.T) {
	got, err := normalizeAPINames([]string{" threads ", "threads", "bluesky", ""})
	if err != nil {
		t.Fatalf("normalizeAPINames() error = %v", err)
	}

	want := []string{"threads", "bluesky"}
	if len(got) != len(want) {
		t.Fatalf("normalizeAPINames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeAPINames() = %v, want %v", got, want)
		}
	}
}
