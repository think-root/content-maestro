package repository

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateRepositoryURL(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		wantStatusCode int
		wantErr        bool
	}{
		{
			name:           "valid repository returns 200",
			statusCode:     http.StatusOK,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "not found repository returns 404",
			statusCode:     http.StatusNotFound,
			wantStatusCode: http.StatusNotFound,
			wantErr:        false,
		},
		{
			name:           "DMCA blocked repository returns 451",
			statusCode:     451,
			wantStatusCode: 451,
			wantErr:        false,
		},
		{
			name:           "redirect returns 301",
			statusCode:     http.StatusMovedPermanently,
			wantStatusCode: http.StatusMovedPermanently,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("Expected HEAD request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			statusCode, err := ValidateRepositoryURL(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepositoryURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if statusCode != tt.wantStatusCode {
				t.Errorf("ValidateRepositoryURL() statusCode = %v, want %v", statusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestValidateRepositoryURL_InvalidURL(t *testing.T) {
	_, err := ValidateRepositoryURL("http://invalid-url-that-does-not-exist.local")
	if err == nil {
		t.Error("ValidateRepositoryURL() expected error for invalid URL, got nil")
	}
}

func TestDeleteRepository(t *testing.T) {
	originalClient := client
	originalURL := deleteRepositoryUrl
	originalBearer := bearerToken

	bearerToken = "Bearer test-token"

	defer func() {
		client = originalClient
		deleteRepositoryUrl = originalURL
		bearerToken = originalBearer
	}()

	tests := []struct {
		name           string
		url            string
		serverResponse string
		statusCode     int
		checkRequest   bool
		wantErr        bool
		wantResult     bool
	}{
		{
			name:           "successful delete",
			url:            "https://github.com/example/repo",
			serverResponse: `{"status": "ok", "message": "Repository deleted successfully"}`,
			statusCode:     http.StatusOK,
			checkRequest:   true,
			wantErr:        false,
			wantResult:     true,
		},
		{
			name:           "repository not found",
			url:            "https://github.com/example/not-found",
			serverResponse: `{"status": "error", "message": "repository with URL https://github.com/example/not-found not found"}`,
			statusCode:     http.StatusOK,
			checkRequest:   false,
			wantErr:        false,
			wantResult:     false,
		},
		{
			name:           "invalid json response",
			url:            "https://github.com/example/invalid",
			serverResponse: `invalid json`,
			statusCode:     http.StatusOK,
			checkRequest:   false,
			wantErr:        true,
			wantResult:     false,
		},
		{
			name:           "server error",
			url:            "https://github.com/example/error",
			serverResponse: `{"status": "error", "message": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			checkRequest:   false,
			wantErr:        false,
			wantResult:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be application/json")
				}

				if r.Header.Get("Authorization") != bearerToken {
					t.Errorf("Expected Authorization header to be %s, got %s", bearerToken, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			deleteRepositoryUrl = server.URL
			client = server.Client()

			result, err := DeleteRepository(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.wantResult {
				t.Errorf("DeleteRepository() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}
