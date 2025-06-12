package validation

import (
	"testing"
	"time"
)

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name        string
		startDate   string
		endDate     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid date range with date-only format",
			startDate:   "2024-03-01",
			endDate:     "2024-03-15",
			expectError: false,
		},
		{
			name:        "valid date range with RFC3339 format",
			startDate:   "2024-03-01T00:00:00Z",
			endDate:     "2024-03-15T23:59:59Z",
			expectError: false,
		},
		{
			name:        "mixed date formats",
			startDate:   "2024-03-01",
			endDate:     "2024-03-15T23:59:59Z",
			expectError: false,
		},
		{
			name:        "start date only",
			startDate:   "2024-03-01",
			endDate:     "",
			expectError: false,
		},
		{
			name:        "end date only",
			startDate:   "",
			endDate:     "2024-03-15",
			expectError: false,
		},
		{
			name:        "both dates empty",
			startDate:   "",
			endDate:     "",
			expectError: false,
		},
		{
			name:        "invalid start date format",
			startDate:   "2024-13-01",
			endDate:     "2024-03-15",
			expectError: true,
			errorMsg:    "invalid start_date",
		},
		{
			name:        "invalid end date format",
			startDate:   "2024-03-01",
			endDate:     "2024-03-32",
			expectError: true,
			errorMsg:    "invalid end_date",
		},
		{
			name:        "start date after end date",
			startDate:   "2024-03-15",
			endDate:     "2024-03-01",
			expectError: true,
			errorMsg:    "start_date cannot be after end_date",
		},
		{
			name:        "malformed date string",
			startDate:   "not-a-date",
			endDate:     "2024-03-15",
			expectError: true,
			errorMsg:    "invalid start_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate, err := ParseDateRange(tt.startDate, tt.endDate)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error()[:len(tt.errorMsg)] != tt.errorMsg {
					t.Errorf("expected error message to start with '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.startDate != "" && startDate == nil {
				t.Errorf("expected start date to be parsed, got nil")
			}
			if tt.endDate != "" && endDate == nil {
				t.Errorf("expected end date to be parsed, got nil")
			}
			if tt.startDate == "" && startDate != nil {
				t.Errorf("expected start date to be nil, got %v", startDate)
			}
			if tt.endDate == "" && endDate != nil {
				t.Errorf("expected end date to be nil, got %v", endDate)
			}
		})
	}
}

func TestIsWithinDateRange(t *testing.T) {
	testTime := time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC)

	startDate := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		startDate *time.Time
		endDate   *time.Time
		expected  bool
	}{
		{
			name:      "within range",
			timestamp: testTime,
			startDate: &startDate,
			endDate:   &endDate,
			expected:  true,
		},
		{
			name:      "before start date",
			timestamp: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			startDate: &startDate,
			endDate:   &endDate,
			expected:  false,
		},
		{
			name:      "after end date",
			timestamp: time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC),
			startDate: &startDate,
			endDate:   &endDate,
			expected:  false,
		},
		{
			name:      "no date constraints",
			timestamp: testTime,
			startDate: nil,
			endDate:   nil,
			expected:  true,
		},
		{
			name:      "only start date constraint - within",
			timestamp: testTime,
			startDate: &startDate,
			endDate:   nil,
			expected:  true,
		},
		{
			name:      "only start date constraint - before",
			timestamp: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			startDate: &startDate,
			endDate:   nil,
			expected:  false,
		},
		{
			name:      "only end date constraint - within",
			timestamp: testTime,
			startDate: nil,
			endDate:   &endDate,
			expected:  true,
		},
		{
			name:      "only end date constraint - after",
			timestamp: time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC),
			startDate: nil,
			endDate:   &endDate,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinDateRange(tt.timestamp, tt.startDate, tt.endDate)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
