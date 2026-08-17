package models

import "time"

type CronHistory struct {
	Name      string             `json:"name"`
	Timestamp time.Time          `json:"timestamp"`
	Success   int                `json:"status"`
	Output    string             `json:"output,omitempty"`
	Details   *MessageRunDetails `json:"details,omitempty"`
}

// MessageRunDetails records which item a message run published and where it
// landed, so a partial run can be re-sent to the connectors that missed it
// without re-parsing the human-readable output.
type MessageRunDetails struct {
	URL    string   `json:"url,omitempty"`
	Sent   []string `json:"sent,omitempty"`
	Failed []string `json:"failed,omitempty"`
	Manual bool     `json:"manual,omitempty"`
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
