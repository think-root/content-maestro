package validation

import (
	"fmt"
	"time"
)

// Supported date formats for parsing
const (
	DateOnlyFormat = "2006-01-02"
	RFC3339Format  = time.RFC3339
)

// ParseDateRange parses start and end date strings and validates them
func ParseDateRange(startDateStr, endDateStr string) (*time.Time, *time.Time, error) {
	var startDate, endDate *time.Time

	// Parse start date if provided
	if startDateStr != "" {
		parsed, err := parseDate(startDateStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_date: %v", err)
		}
		startDate = &parsed
	}

	// Parse end date if provided
	if endDateStr != "" {
		parsed, err := parseDate(endDateStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_date: %v", err)
		}
		endDate = &parsed
	}

	// Validate date range
	if startDate != nil && endDate != nil {
		if startDate.After(*endDate) {
			return nil, nil, fmt.Errorf("start_date cannot be after end_date")
		}
	}

	return startDate, endDate, nil
}

// parseDate attempts to parse a date string using multiple formats
func parseDate(dateStr string) (time.Time, error) {
	// Try date-only format first (2006-01-02)
	if parsed, err := time.Parse(DateOnlyFormat, dateStr); err == nil {
		return parsed, nil
	}

	// Try RFC3339 format (2006-01-02T15:04:05Z07:00)
	if parsed, err := time.Parse(RFC3339Format, dateStr); err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("date must be in format YYYY-MM-DD or RFC3339 (e.g., 2006-01-02 or 2006-01-02T15:04:05Z)")
}

// IsWithinDateRange checks if a timestamp falls within the specified date range
func IsWithinDateRange(timestamp time.Time, startDate, endDate *time.Time) bool {
	if startDate != nil && timestamp.Before(*startDate) {
		return false
	}
	if endDate != nil && timestamp.After(endDate.Add(24*time.Hour-time.Nanosecond)) {
		return false
	}
	return true
}
