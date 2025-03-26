package repository

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRepository(t *testing.T) {
	originalClient := client
	originalURL := getRepositoryUrl
	originalBearer := bearerToken

	bearerToken = "Bearer test-token"

	defer func() {
		client = originalClient
		getRepositoryUrl = originalURL
		bearerToken = originalBearer
	}()

	tests := []struct {
		name           string
		limit          int
		posted         bool
		sort_by        string
		sort_order     string
		serverResponse string
		statusCode     int
		wantErr        bool
		expectedAll    int
	}{
		{
			name:   "successful response",
			limit:  10,
			posted: false,
			serverResponse: `{
				"data": {
					"all": 42,
					"posted": 20,
					"unposted": 22,
					"items": [
						{
							"id": 1,
							"posted": false,
							"url": "https://example.com/1",
							"text": "Example 1"
						}
					]
				},
				"message": "Success",
				"status": "ok"
			}`,
			statusCode:  http.StatusOK,
			wantErr:     false,
			expectedAll: 42,
		},
		{
			name:           "invalid json response",
			limit:          5,
			posted:         true,
			sort_by:        "date_added",
			sort_order:     "ASC",
			serverResponse: `invalid json`,
			statusCode:     http.StatusOK,
			wantErr:        true,
			expectedAll:    0,
		},
		{
			name:           "server error",
			limit:          5,
			sort_by:        "date_added",
			sort_order:     "ASC",
			posted:         true,
			serverResponse: `{"error": "Internal Server Error"}`,
			statusCode:     http.StatusInternalServerError,
			wantErr:        false,
			expectedAll:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
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

			getRepositoryUrl = server.URL
			client = server.Client()

			resp, err := GetRepository(tt.limit, tt.posted, tt.sort_order, tt.sort_by)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && resp.Data.All != tt.expectedAll {
				t.Errorf("GetRepository() expected All = %v, got = %v", tt.expectedAll, resp.Data.All)
			}
		})
	}
}
