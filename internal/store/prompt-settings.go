package store

import (
	"content-maestro/internal/models"
	"fmt"
	"time"
)

func (s *PostgresStore) GetPromptSettings() (*models.PromptSettings, error) {
	var settings models.PromptSettings
	err := s.db.QueryRow(`
		SELECT use_direct_url, llm_provider, temperature, content, model, llm_output_language, updated_at
		FROM think_prompt
		WHERE id = 1
	`).Scan(&settings.UseDirectURL, &settings.LlmProvider, &settings.Temperature, &settings.Content, &settings.Model, &settings.LlmOutputLanguage, &settings.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *PostgresStore) UpdatePromptSettings(settings *models.UpdatePromptSettingsRequest) error {
	query := `
		UPDATE think_prompt 
		SET updated_at = $1`
	args := []interface{}{time.Now()}
	argIndex := 2
	
	if settings.UseDirectURL != nil {
		query += fmt.Sprintf(", use_direct_url = $%d", argIndex)
		args = append(args, *settings.UseDirectURL)
		argIndex++
	}
	
	if settings.LlmProvider != nil {
		query += fmt.Sprintf(", llm_provider = $%d", argIndex)
		args = append(args, *settings.LlmProvider)
		argIndex++
	}
	
	if settings.Temperature != nil {
		query += fmt.Sprintf(", temperature = $%d", argIndex)
		args = append(args, *settings.Temperature)
		argIndex++
	}
	
	if settings.Content != nil {
		query += fmt.Sprintf(", content = $%d", argIndex)
		args = append(args, *settings.Content)
		argIndex++
	}
	
	if settings.Model != nil {
		query += fmt.Sprintf(", model = $%d", argIndex)
		args = append(args, *settings.Model)
		argIndex++
	}
	
	if settings.LlmOutputLanguage != nil {
		query += fmt.Sprintf(", llm_output_language = $%d", argIndex)
		args = append(args, *settings.LlmOutputLanguage)
		argIndex++
	}
	
	query += " WHERE id = 1"
	
	_, err := s.db.Exec(query, args...)
	return err
}