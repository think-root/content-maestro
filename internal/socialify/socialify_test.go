package socialify

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

func setupTestEnvironment(t *testing.T) func() {
	err := os.MkdirAll("./tmp/gh_project_img", 0755)
	if err != nil {
		t.Fatal(err)
	}

	return func() {
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

			oldClient := http.DefaultClient
			http.DefaultClient = client
			defer func() { http.DefaultClient = oldClient }()

			err := Socialify(tt.usernameRepo)

			if (err != nil) != tt.expectedError {
				t.Errorf("Socialify() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError {
				data, err := os.ReadFile("./tmp/gh_project_img/image.png")
				if err != nil {
					t.Fatalf("Failed to read created image: %v", err)
				}
				if !bytes.Equal(data, tt.responseBody) {
					t.Error("Created image content does not match expected data")
				}
			}
		})
	}
}

func TestSocialifyInvalidPath(t *testing.T) {
	originalDir := "./tmp/gh_project_img"
	os.RemoveAll(originalDir)

	err := Socialify("test/repo")
	if err == nil {
		t.Error("Expected error when directory doesn't exist, got nil")
	}
}
