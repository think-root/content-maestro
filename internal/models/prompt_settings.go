package models

import "time"

type PromptSettings struct {
	UseDirectURL bool      `json:"use_direct_url"`
	LlmProvider  string    `json:"llm_provider"`
	Temperature  float64   `json:"temperature"`
	Content      string    `json:"content"`
	Model        string    `json:"model"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdatePromptSettingsRequest struct {
	UseDirectURL *bool    `json:"use_direct_url,omitempty"`
	LlmProvider  *string  `json:"llm_provider,omitempty"`
	Temperature  *float64 `json:"temperature,omitempty"`
	Content      *string  `json:"content,omitempty"`
	Model        *string  `json:"model,omitempty"`
}