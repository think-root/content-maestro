package models

import "time"

type CronHistory struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type PaginationMetadata struct {
	TotalCount  int  `json:"total_count"`
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

type PaginatedCronHistoryResponse struct {
	Data       []CronHistory      `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}
