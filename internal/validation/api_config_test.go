package validation

import (
	"content-maestro/internal/models"
	"testing"
)

func TestValidateDefaultJSONBody(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "empty string is valid",
			input:       "",
			shouldError: false,
		},
		{
			name:        "valid JSON object with string values",
			input:       `{"key1": "value1", "key2": "value2"}`,
			shouldError: false,
		},
		{
			name:        "valid empty JSON object",
			input:       `{}`,
			shouldError: false,
		},
		{
			name:        "invalid JSON - not an object",
			input:       `["array"]`,
			shouldError: true,
		},
		{
			name:        "invalid JSON - syntax error",
			input:       `{invalid}`,
			shouldError: true,
		},
		{
			name:        "invalid JSON - number values instead of strings",
			input:       `{"key": 123}`,
			shouldError: true,
		},
		{
			name:        "invalid JSON - boolean values instead of strings",
			input:       `{"key": true}`,
			shouldError: true,
		},
		{
			name:        "JSON with null values - accepted but null is omitted",
			input:       `{"key": null}`,
			shouldError: false,
		},
		{
			name:        "invalid JSON - nested object",
			input:       `{"key": {"nested": "value"}}`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefaultJSONBody(tt.input)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateAPIConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *models.CreateAPIConfigRequest
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid config with empty default_json_body",
			config: &models.CreateAPIConfigRequest{
				Name:            "test",
				URL:             "https://example.com",
				Method:          "POST",
				AuthType:        "bearer",
				ContentType:     "json",
				Timeout:         30,
				SuccessCode:     200,
				DefaultJSONBody: "",
			},
			shouldError: false,
		},
		{
			name: "valid config with valid default_json_body",
			config: &models.CreateAPIConfigRequest{
				Name:            "test",
				URL:             "https://example.com",
				Method:          "POST",
				AuthType:        "bearer",
				ContentType:     "json",
				Timeout:         30,
				SuccessCode:     200,
				DefaultJSONBody: `{"key": "value"}`,
			},
			shouldError: false,
		},
		{
			name: "invalid config with invalid default_json_body",
			config: &models.CreateAPIConfigRequest{
				Name:            "test",
				URL:             "https://example.com",
				Method:          "POST",
				AuthType:        "bearer",
				ContentType:     "json",
				Timeout:         30,
				SuccessCode:     200,
				DefaultJSONBody: `{invalid json}`,
			},
			shouldError: true,
			errorMsg:    "default_json_body must be valid JSON",
		},
		{
			name: "invalid config with non-string values in default_json_body",
			config: &models.CreateAPIConfigRequest{
				Name:            "test",
				URL:             "https://example.com",
				Method:          "POST",
				AuthType:        "bearer",
				ContentType:     "json",
				Timeout:         30,
				SuccessCode:     200,
				DefaultJSONBody: `{"key": 123}`,
			},
			shouldError: true,
			errorMsg:    "default_json_body must be valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIConfig(tt.config)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateAPIConfigUpdate(t *testing.T) {
	validJSON := `{"key": "value"}`
	invalidJSON := `{invalid}`
	nonStringJSON := `{"key": 123}`

	tests := []struct {
		name        string
		config      *models.UpdateAPIConfigRequest
		shouldError bool
	}{
		{
			name: "valid update with no default_json_body",
			config: &models.UpdateAPIConfigRequest{
				Timeout: intPtr(30),
			},
			shouldError: false,
		},
		{
			name: "valid update with valid default_json_body",
			config: &models.UpdateAPIConfigRequest{
				DefaultJSONBody: &validJSON,
			},
			shouldError: false,
		},
		{
			name: "invalid update with invalid default_json_body",
			config: &models.UpdateAPIConfigRequest{
				DefaultJSONBody: &invalidJSON,
			},
			shouldError: true,
		},
		{
			name: "invalid update with non-string values in default_json_body",
			config: &models.UpdateAPIConfigRequest{
				DefaultJSONBody: &nonStringJSON,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIConfigUpdate(tt.config)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
