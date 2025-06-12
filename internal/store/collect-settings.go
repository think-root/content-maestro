package store

type CollectSettings struct {
	MaxRepos           int    `json:"max_repos"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
}

func (s *PostgresStore) GetCollectSettings() (*CollectSettings, error) {
	var settings CollectSettings
	err := s.db.QueryRow(`
		SELECT max_repos, since, spoken_language_code
		FROM maestro_collect_settings
		WHERE id = 1
	`).Scan(&settings.MaxRepos, &settings.Since, &settings.SpokenLanguageCode)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *PostgresStore) UpdateCollectSettings(settings *CollectSettings) error {
	_, err := s.db.Exec(`
		UPDATE maestro_collect_settings
		SET max_repos = $1, since = $2, spoken_language_code = $3
		WHERE id = 1
	`, settings.MaxRepos, settings.Since, settings.SpokenLanguageCode)
	return err
}
