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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS think_prompt (
			id SERIAL PRIMARY KEY,
			use_direct_url BOOLEAN NOT NULL DEFAULT TRUE,
			llm_provider VARCHAR(255) NOT NULL DEFAULT 'openrouter',
			temperature DECIMAL(3,2) NOT NULL DEFAULT 0.2,
			content TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create think_prompt table: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO think_prompt (use_direct_url, llm_provider, temperature, content, updated_at)
		SELECT TRUE, 'openrouter', 0.2, 'Ти — AI асистент, що спеціалізується на створенні коротких описів GitHub-репозиторіїв українською мовою. Твоя відповідь **ПОВИННА** суворо відповідати **КОЖНІЙ** з наведених нижче вимог. Будь-яке відхилення, особливо щодо довжини тексту, є неприпустимим. Твоя основна задача — створювати описи на основі наданих URL.

Під час створення опису **НЕУХИЛЬНО** дотримуйся наступних правил:

1.  Включай не більше однієї ключової функції репозиторію.
2.  **ЗАБОРОНЕНО** додавати будь-які посилання.
3.  Пиши простою, зрозумілою мовою, без переліків. Інформацію про функції вплітай у зв''язний текст.
4.  **ЗАБОРОНЕНО** згадувати сумісність, платформи, авторів, компанії або колаборації.
5.  **ЗАБОРОНЕНО** використовувати будь-яку розмітку: ні HTML, ні Markdown.
6.  Опис має бути **НАДЗВИЧАЙНО** лаконічним. **АБСОЛЮТНИЙ МАКСИМУМ — 275 символів**, враховуючи пробіли. **Це найважливіша вимога! Перевищення ліміту є КРИТИЧНОЮ помилкою.**
7.  Технічні терміни (назви мов програмування, бібліотек, інструментів, команд тощо) залишай англійською мовою.
8.  **ПЕРЕД НАДАННЯМ ВІДПОВІДІ:** Переконайся, що текст відповідає **ВСІМ** вимогам. **ОБОВ''ЯЗКОВО ПЕРЕВІР** довжину. Якщо вона перевищує 270 символів, **ПЕРЕПИШИ І СКОРОТИ** його, доки він не буде відповідати ліміту.

Тобі буде надано URL GitHub-репозиторію. Ознайомся з ним і згенеруй опис, що **ТОЧНО** відповідає цим інструкціям.', CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM think_prompt)`)
	if err != nil {
		return fmt.Errorf("failed to insert default prompt settings: %v", err)
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
	query := "SELECT name, schedule, is_active, updated_at FROM maestro_cron_settings ORDER BY id ASC"
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
	if name == "" {
		return fmt.Errorf("cron job name cannot be empty")
	}
	
	const maxOutputLength = 10000
	if len(output) > maxOutputLength {
		output = output[:maxOutputLength-50] + "... [truncated due to length]"
	}
	
	query := "INSERT INTO maestro_cron_history (name, timestamp, success, output) VALUES ($1, $2, $3, $4)"
	timestamp := time.Now()
	
	_, err := s.db.Exec(query, name, timestamp, success, output)
	if err != nil {
		fmt.Printf("Failed to log cron execution to database: %v\n", err)
		fmt.Printf("Attempted to log: name=%s, success=%t, timestamp=%v, output_length=%d\n",
			name, success, timestamp, len(output))
		return fmt.Errorf("failed to log cron execution: %v", err)
	}
	
	fmt.Printf("Successfully logged cron execution: name=%s, success=%t, timestamp=%v\n",
		name, success, timestamp)
	
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
