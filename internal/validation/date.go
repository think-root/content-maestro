package validation

import (
	"fmt"
	"time"
)

const (
	DateOnlyFormat = "2006-01-02"
	RFC3339Format  = time.RFC3339
)

func ParseDateRange(startDateStr, endDateStr string) (*time.Time, *time.Time, error) {
	var startDate, endDate *time.Time

	if startDateStr != "" {
		parsed, err := parseDate(startDateStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_date: %v", err)
		}
		startDate = &parsed
	}

	if endDateStr != "" {
		parsed, err := parseDate(endDateStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_date: %v", err)
		}
		endDate = &parsed
	}

	if startDate != nil && endDate != nil {
		if startDate.After(*endDate) {
			return nil, nil, fmt.Errorf("start_date cannot be after end_date")
		}
	}

	return startDate, endDate, nil
}

func parseDate(dateStr string) (time.Time, error) {
	if parsed, err := time.Parse(DateOnlyFormat, dateStr); err == nil {
		return parsed, nil
	}

	if parsed, err := time.Parse(RFC3339Format, dateStr); err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("date must be in format YYYY-MM-DD or RFC3339 (e.g., 2006-01-02 or 2006-01-02T15:04:05Z)")
}

func IsWithinDateRange(timestamp time.Time, startDate, endDate *time.Time) bool {
	if startDate != nil && timestamp.Before(*startDate) {
		return false
	}
	if endDate != nil && timestamp.After(endDate.Add(24*time.Hour-time.Nanosecond)) {
		return false
	}
	return true
}
