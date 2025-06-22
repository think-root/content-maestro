package validation

import (
	"errors"
	"content-maestro/internal/models"
	"slices"
)

var validLlmProviders = []string{"openrouter", "openai", "mistral_api", "mistral_agent"}

func ValidatePromptSettings(settings *models.UpdatePromptSettingsRequest) error {
	if settings.Temperature != nil {
		if *settings.Temperature < 0.0 || *settings.Temperature > 2.0 {
			return errors.New("temperature must be between 0.0 and 2.0")
		}
	}
	
	if settings.LlmProvider != nil {
		valid := slices.Contains(validLlmProviders, *settings.LlmProvider)
		if !valid {
			return errors.New("invalid llm_provider")
		}
	}
	
	if settings.Content != nil && *settings.Content == "" {
		return errors.New("content cannot be empty")
	}
	
	if settings.Model != nil && *settings.Model == "" {
		return errors.New("model cannot be empty")
	}
	
	return nil
}