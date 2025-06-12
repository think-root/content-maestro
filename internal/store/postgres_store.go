package store

import (
	"database/sql"
	"fmt"
	"time"

	"content-maestro/internal/models"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(host, port, user, password, dbname string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %v", err)
	}

	err = createTablesIfNotExist(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &PostgresStore{db: db}, nil
}

func createTablesIfNotExist(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS maestro_cron_settings (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			schedule VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create maestro_cron_settings table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS maestro_cron_history (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			success BOOLEAN NOT NULL,
			output TEXT
		)`)
	if err != nil {
		return fmt.Errorf("failed to create maestro_cron_history table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS maestro_collect_settings (
			id SERIAL PRIMARY KEY,
			max_repos INTEGER NOT NULL DEFAULT 5,
			since VARCHAR(50) NOT NULL DEFAULT 'daily',
			spoken_language_code VARCHAR(10) NOT NULL DEFAULT 'en'
		)`)
	if err != nil {
		return fmt.Errorf("failed to create maestro_collect_settings table: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO maestro_cron_settings (name, schedule, is_active, updated_at)
		SELECT 'collect', '13 13 * * 6', TRUE, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM maestro_cron_settings WHERE name = 'collect')`)
	if err != nil {
		return fmt.Errorf("failed to insert default collect setting: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO maestro_cron_settings (name, schedule, is_active, updated_at)
		SELECT 'message', '12 12 * * *', TRUE, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM maestro_cron_settings WHERE name = 'message')`)
	if err != nil {
		return fmt.Errorf("failed to insert default message setting: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO maestro_collect_settings (max_repos, since, spoken_language_code)
		SELECT 5, 'daily', 'en'
		WHERE NOT EXISTS (SELECT 1 FROM maestro_collect_settings)`)
	if err != nil {
		return fmt.Errorf("failed to insert default collect settings: %v", err)
	}

	return nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) GetCronSetting(name string) (*models.CronSetting, error) {
	var setting models.CronSetting
	query := "SELECT name, schedule, is_active, updated_at FROM maestro_cron_settings WHERE name = $1"
	err := s.db.QueryRow(query, name).Scan(&setting.Name, &setting.Schedule, &setting.IsActive, &setting.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cron setting: %v", err)
	}
	return &setting, nil
}

func (s *PostgresStore) GetAllCronSettings() ([]models.CronSetting, error) {
	var settings []models.CronSetting
	query := "SELECT name, schedule, is_active, updated_at FROM maestro_cron_settings"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cron settings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var setting models.CronSetting
		if err := rows.Scan(&setting.Name, &setting.Schedule, &setting.IsActive, &setting.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cron setting: %v", err)
		}
		settings = append(settings, setting)
	}
	return settings, nil
}

func (s *PostgresStore) UpdateCronSetting(name string, schedule string, isActive bool) (*models.CronSetting, error) {
	setting := models.CronSetting{
		Name:      name,
		Schedule:  schedule,
		IsActive:  isActive,
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO maestro_cron_settings (name, schedule, is_active, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		SET schedule = $2, is_active = $3, updated_at = $4
		RETURNING name, schedule, is_active, updated_at`
	err := s.db.QueryRow(query, name, schedule, isActive, setting.UpdatedAt).
		Scan(&setting.Name, &setting.Schedule, &setting.IsActive, &setting.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update cron setting: %v", err)
	}
	return &setting, nil
}

func (s *PostgresStore) InitializeDefaultSettings() error {
	defaults := []models.CronSetting{
		{
			Name:      "collect",
			Schedule:  "13 13 * * 6",
			IsActive:  false,
			UpdatedAt: time.Now(),
		},
		{
			Name:      "message",
			Schedule:  "12 12 * * *",
			IsActive:  false,
			UpdatedAt: time.Now(),
		},
	}

	for _, setting := range defaults {
		exists, err := s.GetCronSetting(setting.Name)
		if err != nil {
			return fmt.Errorf("failed to check existing setting: %v", err)
		}
		if exists == nil {
			if _, err := s.UpdateCronSetting(setting.Name, setting.Schedule, setting.IsActive); err != nil {
				return fmt.Errorf("failed to initialize default setting: %v", err)
			}
		}
	}

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM maestro_collect_settings").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check collect settings: %v", err)
	}
	if count == 0 {
		query := "INSERT INTO maestro_collect_settings (max_repos, since, spoken_language_code) VALUES ($1, $2, $3)"
		_, err := s.db.Exec(query, 2, "daily", "en")
		if err != nil {
			return fmt.Errorf("failed to initialize collect settings: %v", err)
		}
	}
	return nil
}

func (s *PostgresStore) LogCronExecution(name string, success bool, output string) error {
	query := "INSERT INTO maestro_cron_history (name, timestamp, success, output) VALUES ($1, $2, $3, $4)"
	_, err := s.db.Exec(query, name, time.Now(), success, output)
	if err != nil {
		return fmt.Errorf("failed to log cron execution: %v", err)
	}
	return nil
}

func (s *PostgresStore) GetCronHistoryCount(name string, success *bool, startDate, endDate *time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM maestro_cron_history WHERE 1=1"
	args := []any{}
	argIndex := 1

	if name != "" {
		query += fmt.Sprintf(" AND name = $%d", argIndex)
		args = append(args, name)
		argIndex++
	}
	if success != nil {
		query += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, *success)
		argIndex++
	}
	if startDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, endDate.Add(24*time.Hour-time.Nanosecond))
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cron history: %v", err)
	}
	return count, nil
}

func (s *PostgresStore) GetCronHistory(name string, success *bool, offset, limit int, sortOrder string, startDate, endDate *time.Time) ([]models.CronHistory, error) {
	query := "SELECT name, timestamp, success, output FROM maestro_cron_history WHERE 1=1"
	args := []any{}
	argIndex := 1

	if name != "" {
		query += fmt.Sprintf(" AND name = $%d", argIndex)
		args = append(args, name)
		argIndex++
	}
	if success != nil {
		query += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, *success)
		argIndex++
	}
	if startDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, endDate.Add(24*time.Hour-time.Nanosecond))
		argIndex++
	}

	if sortOrder == "asc" {
		query += " ORDER BY timestamp ASC"
	} else {
		query += " ORDER BY timestamp DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron history: %v", err)
	}
	defer rows.Close()

	var history []models.CronHistory
	for rows.Next() {
		var h models.CronHistory
		if err := rows.Scan(&h.Name, &h.Timestamp, &h.Success, &h.Output); err != nil {
			return nil, fmt.Errorf("failed to scan cron history: %v", err)
		}
		history = append(history, h)
	}
	return history, nil
}

func (s *PostgresStore) HasMigrationFlag() (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'migration_flag')"
	var exists bool
	err := s.db.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check migration flag: %v", err)
	}
	return exists, nil
}

func (s *PostgresStore) SetMigrationFlag() error {
	query := "CREATE TABLE IF NOT EXISTS migration_flag (id SERIAL PRIMARY KEY, migrated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to set migration flag: %v", err)
	}
	return nil
}
