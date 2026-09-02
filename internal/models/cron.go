package models

import "time"

type CronSetting struct {
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CronResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UpdateScheduleRequest struct {
	Schedule string `json:"schedule"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// RetryMessageRequest asks for a repository to be re-sent to the named APIs.
// URL is optional: when empty the most recently published repository is used.
type RetryMessageRequest struct {
	APIs []string `json:"apis"`
	URL  string   `json:"url"`
}

// PublishMessageRequest asks for one repository to be published immediately to
// every enabled integration. URL is required: there is no sensible fallback for
// "publish something now".
type PublishMessageRequest struct {
	URL string `json:"url"`
}
