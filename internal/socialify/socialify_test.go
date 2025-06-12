package socialify

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func setupTestEnvironment(t *testing.T) func() {
	SetRetryConfig(RetryConfig{
		MaxRetries:    2,
		RetryInterval: 100 * time.Millisecond,
	})

	err := os.MkdirAll("./tmp/gh_project_img", 0755)
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		ResetRetryConfig()
		os.RemoveAll("./tmp")
	}
}

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestSocialify(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name           string
		usernameRepo   string
		responseStatus int
		responseBody   []byte
		expectedError  bool
	}{
		{
			name:           "successful image download",
			usernameRepo:   "test/repo",
			responseStatus: http.StatusOK,
			responseBody:   []byte("fake image data"),
			expectedError:  false,
		},
		{
			name:           "server error",
			usernameRepo:   "test/repo",
			responseStatus: http.StatusInternalServerError,
			responseBody:   []byte("error"),
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &http.Response{
				StatusCode: tt.responseStatus,
				Body:       io.NopCloser(bytes.NewReader(tt.responseBody)),
			}

			client := &http.Client{
				Transport: &mockTransport{
					response: response,
				},
			}

			oldClient := SocialifyHTTPClient
			SocialifyHTTPClient = client
			defer func() { SocialifyHTTPClient = oldClient }()

			err := Socialify(tt.usernameRepo)

			if (err != nil) != tt.expectedError {
				t.Errorf("Socialify() error = %v, expectedError %v", err, tt.expectedError)
			} else if tt.expectedError {
				t.Logf("Expected error received: %v", err)
			} else {
				t.Logf("Successfully executed Socialify for %s", tt.usernameRepo)
			}

			if !tt.expectedError {
				data, err := os.ReadFile("./tmp/gh_project_img/image.png")
				if err != nil {
					t.Fatalf("Failed to read created image: %v", err)
				}
				if !bytes.Equal(data, tt.responseBody) {
					t.Error("Created image content does not match expected data")
				} else {
					t.Logf("Image content verified successfully (%d bytes)", len(data))
				}
			}
		})
	}
}

func TestSocialifyInvalidPath(t *testing.T) {
	t.Log("Starting TestSocialifyInvalidPath test")

	SetRetryConfig(RetryConfig{
		MaxRetries:    1,
		RetryInterval: 100 * time.Millisecond,
	})
	defer ResetRetryConfig()

	originalDir := "./tmp/gh_project_img"
	if err := os.RemoveAll(originalDir); err != nil {
		t.Logf("Error removing directory (if exists): %v", err)
	}

	err := Socialify("test/repo")

	if err == nil {
		t.Error("Expected error when directory doesn't exist, got nil")
	} else {
		t.Logf("Received expected error: %v", err)

		if os.IsNotExist(err) {
			t.Log("Confirmed error is 'file not exists' error as expected")
		} else {
			t.Logf("Error type: %T", err)
		}
	}

	if _, err := os.Stat(originalDir); !os.IsNotExist(err) {
		t.Error("Directory should not exist after failed operation")
	} else {
		t.Log("Verified directory still doesn't exist after test")
	}
}
