package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	APIs map[string]APIEndpoint `yaml:"apis"`
}

type APIEndpoint struct {
	URL          string            `yaml:"url"`
	Method       string            `yaml:"method"`
	Headers      map[string]string `yaml:"headers"`
	AuthType     string            `yaml:"auth_type"`
	TokenEnvVar  string            `yaml:"token_env_var"`
	TokenHeader  string            `yaml:"token_header"`
	ContentType  string            `yaml:"content_type"`
	Timeout      int               `yaml:"timeout"`
	SuccessCode  int               `yaml:"success_code"`
	Enabled      bool              `yaml:"enabled"`
	ResponseType string            `yaml:"response_type"`
}

type RequestConfig struct {
	APIName     string                 `json:"api_name" yaml:"api_name"`
	URLParams   map[string]string      `json:"url_params" yaml:"url_params"`
	JSONBody    map[string]interface{} `json:"json_body" yaml:"json_body"`
	FormFields  map[string]string      `json:"form_fields" yaml:"form_fields"`
	FileFields  map[string]string      `json:"file_fields" yaml:"file_fields"`
	RawBody     []byte                 `json:"-" yaml:"-"`
	ExtraParams map[string]interface{} `json:"extra_params" yaml:"extra_params"`
}

type APIResponse struct {
	Success      bool          `json:"success"`
	StatusCode   int           `json:"status_code"`
	Body         []byte        `json:"-"`
	JSONResponse interface{}   `json:"response,omitempty"`
	Error        string        `json:"error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	APIName      string        `json:"api_name"`
	Timestamp    time.Time     `json:"timestamp"`
}

var apiConfig *APIConfig

func GetAPIConfigs() *APIConfig {
	return apiConfig
}

func LoadAPIConfigs(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	config := &APIConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	apiConfig = config
	return nil
}

func ExecuteRequest(reqConfig RequestConfig) (*APIResponse, error) {
	if apiConfig == nil {
		return nil, fmt.Errorf("API configuration not loaded, call LoadAPIConfigs first")
	}

	apiEndpoint, exists := apiConfig.APIs[reqConfig.APIName]
	if !exists {
		return nil, fmt.Errorf("API endpoint '%s' not found in configuration", reqConfig.APIName)
	}

	if !apiEndpoint.Enabled {
		return nil, fmt.Errorf("API endpoint '%s' is disabled", reqConfig.APIName)
	}

	startTime := time.Now()

	url := apiEndpoint.URL
	if reqConfig.URLParams != nil {
		for key, value := range reqConfig.URLParams {
			url = strings.Replace(url, fmt.Sprintf("{%s}", key), value, -1)
		}
	}

	if strings.Contains(url, "{env.") {
		for _, envVar := range extractEnvVarsFromString(url) {
			url = strings.Replace(url, fmt.Sprintf("{env.%s}", envVar), os.Getenv(envVar), -1)
		}
	}

	var body io.Reader
	var contentType string

	switch apiEndpoint.ContentType {
	case "json":
		if reqConfig.JSONBody != nil {
			jsonData, err := json.Marshal(reqConfig.JSONBody)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
			}
			body = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		} else if reqConfig.RawBody != nil {
			body = bytes.NewBuffer(reqConfig.RawBody)
			contentType = "application/json"
		}
	case "multipart":
		bodyBuf := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuf)

		for key, value := range reqConfig.FormFields {
			if err := writer.WriteField(key, value); err != nil {
				return nil, fmt.Errorf("failed to write form field '%s': %w", key, err)
			}
		}

		for fieldName, filePath := range reqConfig.FileFields {
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file '%s': %w", filePath, err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
			if err != nil {
				return nil, fmt.Errorf("failed to create form file '%s': %w", fieldName, err)
			}

			if _, err = io.Copy(part, file); err != nil {
				return nil, fmt.Errorf("failed to copy file contents: %w", err)
			}
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		body = bodyBuf
		contentType = writer.FormDataContentType()
	}

	req, err := http.NewRequest(apiEndpoint.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	for key, value := range apiEndpoint.Headers {
		if strings.Contains(value, "{env.") {
			for _, envVar := range extractEnvVarsFromString(value) {
				value = strings.Replace(value, fmt.Sprintf("{env.%s}", envVar), os.Getenv(envVar), -1)
			}
		}
		req.Header.Set(key, value)
	}

	switch apiEndpoint.AuthType {
	case "bearer":
		token := os.Getenv(apiEndpoint.TokenEnvVar)
		if token == "" {
			return nil, fmt.Errorf("bearer token not found in environment variable '%s'", apiEndpoint.TokenEnvVar)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case "api_key":
		token := os.Getenv(apiEndpoint.TokenEnvVar)
		if token == "" {
			return nil, fmt.Errorf("API key not found in environment variable '%s'", apiEndpoint.TokenEnvVar)
		}
		req.Header.Set(apiEndpoint.TokenHeader, token)
	}

	timeout := time.Duration(apiEndpoint.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	responseTime := time.Since(startTime)

	success := resp.StatusCode == apiEndpoint.SuccessCode

	apiResp := &APIResponse{
		Success:      success,
		StatusCode:   resp.StatusCode,
		Body:         respBody,
		ResponseTime: responseTime,
		APIName:      reqConfig.APIName,
		Timestamp:    time.Now(),
	}

	if success && apiEndpoint.ResponseType == "json" {
		var jsonResponse interface{}
		if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
			log.Printf("Warning: Failed to parse JSON response: %v", err)
		} else {
			apiResp.JSONResponse = jsonResponse
		}
	}

	return apiResp, nil
}

func extractEnvVarsFromString(input string) []string {
	var envVars []string
	start := 0

	for start < len(input) {
		envStart := strings.Index(input[start:], "{env.")
		if envStart == -1 {
			break
		}
		envStart += start + len("{env.")

		envEnd := strings.Index(input[envStart:], "}")
		if envEnd == -1 {
			break
		}
		envEnd += envStart

		envVar := input[envStart:envEnd]
		envVars = append(envVars, envVar)
		start = envEnd + 1
	}

	return envVars
}
