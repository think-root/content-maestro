package validation

import (
	"content-maestro/internal/models"
	"errors"
	"regexp"
	"slices"
	"strings"
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
	
	if settings.LlmOutputLanguage != nil && *settings.LlmOutputLanguage != "" {
		if err := validateLanguageCodes(*settings.LlmOutputLanguage); err != nil {
			return err
		}
	}
	
	return nil
}

func validateLanguageCodes(languageCodes string) error {
	pattern := `^[a-z]{2,3}(,[a-z]{2,3})*$`
	matched, err := regexp.MatchString(pattern, languageCodes)
	if err != nil {
		return errors.New("error validating language codes format")
	}
	
	if !matched {
		return errors.New("llm_output_language must be comma-separated language codes (e.g., 'en,uk,fr')")
	}
	
	codes := strings.Split(languageCodes, ",")
	
	seen := make(map[string]bool)
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if seen[code] {
			return errors.New("duplicate language codes are not allowed")
		}
		seen[code] = true
	}
	
	return nil
}
