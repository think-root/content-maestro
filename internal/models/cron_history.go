package models

import "time"

type CronHistory struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool     `json:"success"`
	Error     string    `json:"error,omitempty"`
}
