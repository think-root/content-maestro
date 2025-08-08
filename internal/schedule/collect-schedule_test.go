package schedule

import (
	"os"
	"testing"
	"time"
)

func TestGetContentAlchemistTimeout(t *testing.T) {
	originalTimeout := os.Getenv("CONTENT_ALCHEMIST_TIMEOUT")
	defer func() {
		if originalTimeout == "" {
			os.Unsetenv("CONTENT_ALCHEMIST_TIMEOUT")
		} else {
			os.Setenv("CONTENT_ALCHEMIST_TIMEOUT", originalTimeout)
		}
	}()

	tests := []struct {
		name           string
		envValue       string
		expectedResult time.Duration
	}{
		{
			name:           "Default timeout when env var is empty",
			envValue:       "",
			expectedResult: 5 * time.Minute,
		},
		{
			name:           "Custom timeout from env var",
			envValue:       "120",
			expectedResult: 120 * time.Second,
		},
		{
			name:           "Invalid env var value falls back to default",
			envValue:       "invalid",
			expectedResult: 5 * time.Minute,
		},
		{
			name:           "Zero timeout from env var",
			envValue:       "0",
			expectedResult: 0 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("CONTENT_ALCHEMIST_TIMEOUT")
			} else {
				os.Setenv("CONTENT_ALCHEMIST_TIMEOUT", tt.envValue)
			}

			result := getContentAlchemistTimeout()
			if result != tt.expectedResult {
				t.Errorf("getContentAlchemistTimeout() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}
