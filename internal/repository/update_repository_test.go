package repository

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateRepositoryPosted(t *testing.T) {
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

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

				if r.Header.Get("Authorization") != testBearerToken {
					t.Errorf("Expected Authorization header to be %s, got %s", testBearerToken, r.Header.Get("Authorization"))
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

			t.Setenv("CONTENT_ALCHEMIST_URL", server.URL)

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

func TestEndpointsFollowEnvironment(t *testing.T) {
	t.Setenv("CONTENT_ALCHEMIST_URL", "https://test.example.com")
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "update posted", got: updatePostedURL(), want: "https://test.example.com/think-root/api/update-posted/"},
		{name: "get repository", got: getRepositoryURL(), want: "https://test.example.com/think-root/api/get-repository/"},
		{name: "delete repository", got: deleteRepositoryURL(), want: "https://test.example.com/think-root/api/delete-repository/"},
		{name: "authorization", got: authorizationHeader(), want: "Bearer test-token"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %s, want %s", tt.name, tt.got, tt.want)
		}
	}

	// A trailing slash in the configured base must not double up in the path.
	t.Setenv("CONTENT_ALCHEMIST_URL", "https://test.example.com/")
	if got, want := getRepositoryURL(), "https://test.example.com/think-root/api/get-repository/"; got != want {
		t.Errorf("get repository with trailing slash = %s, want %s", got, want)
	}
}
