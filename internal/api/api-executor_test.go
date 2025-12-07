package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadAPIConfigs(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	testConfig := `
apis:
  test_api:
    url: "http://example.com/api"
    method: "POST"
    headers:
      Content-Type: "application/json"
    auth_type: "bearer"
    token_env_var: "API_TOKEN"
    content_type: "json"
    timeout: 30
    success_code: 200
    enabled: true
    response_type: "json"
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	if err := LoadAPIConfigs(configPath); err != nil {
		t.Errorf("LoadAPIConfigs failed: %v", err)
	}

	cfg := GetAPIConfigs()
	if cfg == nil {
		t.Fatal("GetAPIConfigs returned nil")
	}

	if len(cfg.APIs) != 1 {
		t.Errorf("Expected 1 API config, got %d", len(cfg.APIs))
	}

	api := cfg.APIs["test_api"]
	if api.URL != "http://example.com/api" {
		t.Errorf("Expected URL 'http://example.com/api', got '%s'", api.URL)
	}
}

func TestExecuteRequest(t *testing.T) {
	var lastBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		if len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &lastBody)
		} else {
			lastBody = nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	apiConfig = &APIConfig{
		APIs: map[string]APIEndpoint{
			"test_api": {
				URL:          server.URL,
				Method:       "POST",
				AuthType:     "bearer",
				TokenEnvVar:  "API_TOKEN",
				ContentType:  "json",
				Timeout:      30,
				SuccessCode:  200,
				Enabled:      true,
				ResponseType: "json",
				DefaultJSONBody: map[string]string{
					"default_field": "{env.EXTRA_VALUE}",
				},
			},
		},
	}

	os.Setenv("API_TOKEN", "test-token")
	defer os.Unsetenv("API_TOKEN")
	os.Setenv("EXTRA_VALUE", "from-env")
	defer os.Unsetenv("EXTRA_VALUE")

	tests := []struct {
		name        string
		reqConfig   RequestConfig
		wantStatus  int
		wantSuccess bool
		wantBody    map[string]any
	}{
		{
			name: "successful request",
			reqConfig: RequestConfig{
				APIName:  "test_api",
				JSONBody: map[string]any{"test": "data"},
			},
			wantStatus:  200,
			wantSuccess: true,
			wantBody: map[string]any{
				"test":          "data",
				"default_field": "from-env",
			},
		},
		{
			name: "disabled api",
			reqConfig: RequestConfig{
				APIName: "nonexistent_api",
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastBody = nil

			resp, err := ExecuteRequest(tt.reqConfig)
			if tt.wantSuccess {
				if err != nil {
					t.Errorf("ExecuteRequest() error = %v", err)
					return
				}
				if resp.StatusCode != tt.wantStatus {
					t.Errorf("ExecuteRequest() status = %v, want %v", resp.StatusCode, tt.wantStatus)
				}
				if tt.wantBody != nil && !reflect.DeepEqual(lastBody, tt.wantBody) {
					t.Errorf("ExecuteRequest() body = %+v, want %+v", lastBody, tt.wantBody)
				}
			} else if err == nil {
				t.Error("ExecuteRequest() expected error, got nil")
			}
		})
	}
}

func TestExtractEnvVarsFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "https://api.example.com/{env.API_KEY}/data",
			expected: []string{"API_KEY"},
		},
		{
			input:    "no env vars here",
			expected: nil,
		},
		{
			input:    "{env.VAR1}/path/{env.VAR2}",
			expected: []string{"VAR1", "VAR2"},
		},
	}

	for _, tt := range tests {
		result := extractEnvVarsFromString(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("For input %s, expected %v but got %v", tt.input, tt.expected, result)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("For input %s, expected %v but got %v", tt.input, tt.expected, result)
			}
		}
	}
}
