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
