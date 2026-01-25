package validation

import (
	"content-maestro/internal/models"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

var validMethods = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"PATCH":  true,
}

var validAuthTypes = map[string]bool{
	"":        true,
	"bearer":  true,
	"api_key": true,
}

var validContentTypes = map[string]bool{
	"json":      true,
	"multipart": true,
}

func validateDefaultJSONBody(defaultJSONBody string) error {
	if defaultJSONBody == "" {
		return nil
	}

	var jsonObj map[string]string
	if err := json.Unmarshal([]byte(defaultJSONBody), &jsonObj); err != nil {
		return fmt.Errorf("default_json_body must be valid JSON representing an object with string values: %w", err)
	}

	return nil
}

func ValidateAPIConfig(config *models.CreateAPIConfigRequest) error {
	if config.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	namePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !namePattern.MatchString(config.Name) {
		return fmt.Errorf("name must contain only alphanumeric characters, hyphens, and underscores")
	}

	if config.URL == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if _, err := url.Parse(config.URL); err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if !validMethods[config.Method] {
		return fmt.Errorf("invalid method: must be one of GET, POST, PUT, DELETE, PATCH")
	}

	if !validAuthTypes[config.AuthType] {
		return fmt.Errorf("invalid auth_type: must be one of bearer, api_key, or empty")
	}

	if !validContentTypes[config.ContentType] {
		return fmt.Errorf("invalid content_type: must be json or multipart")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if config.SuccessCode < 100 || config.SuccessCode > 599 {
		return fmt.Errorf("success_code must be a valid HTTP status code (100-599)")
	}

	if err := validateDefaultJSONBody(config.DefaultJSONBody); err != nil {
		return err
	}

	return nil
}

func ValidateAPIConfigUpdate(config *models.UpdateAPIConfigRequest) error {
	if config.URL != nil {
		if *config.URL == "" {
			return fmt.Errorf("url cannot be empty")
		}
		if _, err := url.Parse(*config.URL); err != nil {
			return fmt.Errorf("invalid url format: %w", err)
		}
	}

	if config.Method != nil && !validMethods[*config.Method] {
		return fmt.Errorf("invalid method: must be one of GET, POST, PUT, DELETE, PATCH")
	}

	if config.AuthType != nil && !validAuthTypes[*config.AuthType] {
		return fmt.Errorf("invalid auth_type: must be one of bearer, api_key, or empty")
	}

	if config.ContentType != nil && !validContentTypes[*config.ContentType] {
		return fmt.Errorf("invalid content_type: must be json or multipart")
	}

	if config.Timeout != nil && *config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if config.SuccessCode != nil && (*config.SuccessCode < 100 || *config.SuccessCode > 599) {
		return fmt.Errorf("success_code must be a valid HTTP status code (100-599)")
	}

	if config.DefaultJSONBody != nil {
		if err := validateDefaultJSONBody(*config.DefaultJSONBody); err != nil {
			return err
		}
	}

	return nil
}
