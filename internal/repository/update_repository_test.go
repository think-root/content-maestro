package repository

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUpdateRepositoryPosted(t *testing.T) {
	originalClient := client
	originalURL := updatePostedUrl
	originalBearer := bearerToken

	bearerToken = "Bearer test-token"

	defer func() {
		client = originalClient
		updatePostedUrl = originalURL
		bearerToken = originalBearer
	}()

	tests := []struct {
		name           string
		url            string
		posted         bool
		serverResponse string
		statusCode     int
		checkRequest   bool
		wantErr        bool
		wantResult     bool
	}{
		{
			name:           "successful update",
			url:            "https://example.com/post",
			posted:         true,
			serverResponse: `{"message": "Updated successfully", "status": "ok"}`,
			statusCode:     http.StatusOK,
			checkRequest:   true,
			wantErr:        false,
			wantResult:     true,
		},
		{
			name:           "failed update",
			url:            "https://example.com/error",
			posted:         false,
			serverResponse: `{"message": "Update failed", "status": "error"}`,
			statusCode:     http.StatusOK,
			checkRequest:   false,
			wantErr:        false,
			wantResult:     false,
		},
		{
			name:           "invalid json response",
			url:            "https://example.com/invalid",
			posted:         true,
			serverResponse: `invalid json`,
			statusCode:     http.StatusOK,
			checkRequest:   false,
			wantErr:        true,
			wantResult:     false,
		},
		{
			name:           "server error",
			url:            "https://example.com/server-error",
			posted:         true,
			serverResponse: `{"error": "Internal Server Error"}`,
			statusCode:     http.StatusInternalServerError,
			checkRequest:   false,
			wantErr:        false,
			wantResult:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("Expected PATCH request, got %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type header to be application/json")
				}

				if r.Header.Get("Authorization") != bearerToken {
					t.Errorf("Expected Authorization header to be %s, got %s", bearerToken, r.Header.Get("Authorization"))
				}

				if tt.checkRequest {
					body, _ := io.ReadAll(r.Body)
					expected := fmt.Sprintf(`{"url":"%s","posted":%t}`, tt.url, tt.posted)

					if string(body) != expected {
						t.Errorf("Expected payload %s, got %s", expected, string(body))
					}
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			updatePostedUrl = server.URL
			client = server.Client()

			result, err := UpdateRepositoryPosted(tt.url, tt.posted)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRepositoryPosted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.wantResult {
				t.Errorf("UpdateRepositoryPosted() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	originalURL := os.Getenv("CONTENT_ALCHEMIST_URL")
	originalBearer := os.Getenv("CONTENT_ALCHEMIST_BEARER")

	defer func() {
		os.Setenv("CONTENT_ALCHEMIST_URL", originalURL)
		os.Setenv("CONTENT_ALCHEMIST_BEARER", originalBearer)
	}()

	originalUpdateURL := updatePostedUrl
	originalGetURL := getRepositoryUrl
	originalToken := bearerToken

	updatePostedUrl = ""
	getRepositoryUrl = ""
	bearerToken = ""

	os.Setenv("CONTENT_ALCHEMIST_URL", "https://test.example.com")
	os.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	var _ = &http.Client{}

	updatePostedUrl = os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/update-posted/"
	getRepositoryUrl = os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/get-repository/"
	bearerToken = "Bearer " + os.Getenv("CONTENT_ALCHEMIST_BEARER")

	expectedUpdateURL := "https://test.example.com/think-root/api/update-posted/"
	if updatePostedUrl != expectedUpdateURL {
		t.Errorf("Expected updatePostedUrl to be %s, got %s", expectedUpdateURL, updatePostedUrl)
	}

	expectedGetURL := "https://test.example.com/think-root/api/get-repository/"
	if getRepositoryUrl != expectedGetURL {
		t.Errorf("Expected getRepositoryUrl to be %s, got %s", expectedGetURL, getRepositoryUrl)
	}

	expectedToken := "Bearer test-token"
	if bearerToken != expectedToken {
		t.Errorf("Expected bearerToken to be %s, got %s", expectedToken, bearerToken)
	}

	updatePostedUrl = originalUpdateURL
	getRepositoryUrl = originalGetURL
	bearerToken = originalToken
}
