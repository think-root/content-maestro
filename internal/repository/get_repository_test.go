package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testBearerToken = "Bearer test-token"

func TestGetRepository(t *testing.T) {
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

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
			wantErr:        true,
			expectedAll:    0,
		},
		{
			// A rejected request must not be mistaken for an empty queue.
			name:           "rejected request",
			limit:          1,
			sort_by:        "publication_queue",
			sort_order:     "ASC",
			serverResponse: `{"message":"Invalid language code: xx","status":"error"}`,
			statusCode:     http.StatusBadRequest,
			wantErr:        true,
			expectedAll:    0,
		},
		{
			// 401 and 429 come back as plain text from the middleware.
			name:           "plain text rejection",
			limit:          1,
			serverResponse: "Too Many Requests",
			statusCode:     http.StatusTooManyRequests,
			wantErr:        true,
			expectedAll:    0,
		},
		{
			// 200 with an error envelope must not look like a successful fetch.
			name:           "error envelope with 200",
			limit:          1,
			serverResponse: `{"message":"Failed to fetch repositories","status":"error"}`,
			statusCode:     http.StatusOK,
			wantErr:        true,
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

				if r.Header.Get("Authorization") != testBearerToken {
					t.Errorf("Expected Authorization header to be %s, got %s", testBearerToken, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			t.Setenv("CONTENT_ALCHEMIST_URL", server.URL)

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

func TestGetRepositoryByURL(t *testing.T) {
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	const wantURL = "https://github.com/resemble-ai/chatterbox"

	tests := []struct {
		name           string
		requestURL     string
		serverResponse string
		statusCode     int
		wantErr        bool
	}{
		{
			name:       "returns the requested repository",
			requestURL: wantURL,
			serverResponse: `{"data":{"items":[{"id":1327,"posted":true,` +
				`"url":"https://github.com/resemble-ai/chatterbox","text":"English text"}]},"status":"ok"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			// An alchemist that does not know the url filter answers with the head
			// of the queue; publishing that would post the wrong repository.
			name:       "rejects a different repository",
			requestURL: wantURL,
			serverResponse: `{"data":{"items":[{"id":1,"posted":false,` +
				`"url":"https://github.com/open-webui/open-webui","text":"Other text"}]},"status":"ok"}`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:           "reports an empty result",
			requestURL:     wantURL,
			serverResponse: `{"data":{"items":[]},"status":"ok"}`,
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
		{
			name:       "requires a url",
			requestURL: "   ",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body getRepositoryRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if body.URL != tt.requestURL {
					t.Errorf("expected url %q in request, got %q", tt.requestURL, body.URL)
				}
				if body.TextLanguage != "en" {
					t.Errorf("expected text_language en, got %q", body.TextLanguage)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			t.Setenv("CONTENT_ALCHEMIST_URL", server.URL)

			item, err := GetRepositoryByURL(tt.requestURL, "en")
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRepositoryByURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && item.URL != wantURL {
				t.Errorf("GetRepositoryByURL() returned %q, want %q", item.URL, wantURL)
			}
		})
	}
}
