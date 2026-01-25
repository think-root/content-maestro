package models

import "time"

type APIConfigModel struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	URL              string    `json:"url"`
	Method           string    `json:"method"`
	AuthType         string    `json:"auth_type"`
	TokenEnvVar      string    `json:"token_env_var"`
	TokenHeader      string    `json:"token_header"`
	ContentType      string    `json:"content_type"`
	Timeout          int       `json:"timeout"`
	SuccessCode      int       `json:"success_code"`
	Enabled          bool      `json:"enabled"`
	ResponseType     string    `json:"response_type"`
	TextLanguage     string    `json:"text_language"`
	SocialifyImage   bool      `json:"socialify_image"`
	DefaultJSONBody  string    `json:"default_json_body"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateAPIConfigRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Method          string `json:"method"`
	AuthType        string `json:"auth_type"`
	TokenEnvVar     string `json:"token_env_var"`
	TokenHeader     string `json:"token_header"`
	ContentType     string `json:"content_type"`
	Timeout         int    `json:"timeout"`
	SuccessCode     int    `json:"success_code"`
	Enabled         bool   `json:"enabled"`
	ResponseType    string `json:"response_type"`
	TextLanguage    string `json:"text_language"`
	SocialifyImage  bool   `json:"socialify_image"`
	DefaultJSONBody string `json:"default_json_body"`
}

type UpdateAPIConfigRequest struct {
	URL             *string `json:"url,omitempty"`
	Method          *string `json:"method,omitempty"`
	AuthType        *string `json:"auth_type,omitempty"`
	TokenEnvVar     *string `json:"token_env_var,omitempty"`
	TokenHeader     *string `json:"token_header,omitempty"`
	ContentType     *string `json:"content_type,omitempty"`
	Timeout         *int    `json:"timeout,omitempty"`
	SuccessCode     *int    `json:"success_code,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
	ResponseType    *string `json:"response_type,omitempty"`
	TextLanguage    *string `json:"text_language,omitempty"`
	SocialifyImage  *bool   `json:"socialify_image,omitempty"`
	DefaultJSONBody *string `json:"default_json_body,omitempty"`
}
