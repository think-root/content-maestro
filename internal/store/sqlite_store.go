package store

import (
	"content-maestro/internal/models"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DefaultMaxRepos           = 5
	DefaultResource           = "github"
	DefaultSince              = "daily"
	DefaultSpokenLanguageCode = "en"
	DefaultPeriod             = "past_24_hours"
	DefaultLanguage           = "All"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	connStr := fmt.Sprintf("file:%s?_journal_mode=WAL&_timeout=5000&_cache_size=-64000&_synchronous=NORMAL", dbPath)
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := createTablesIfNotExist(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &SQLiteStore{db: db}, nil
}

func createTablesIfNotExist(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cron_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			schedule TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create cron_settings table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cron_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			success INTEGER NOT NULL,
			output TEXT
		)`)
	if err != nil {
		return fmt.Errorf("failed to create cron_history table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS collect_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			max_repos INTEGER NOT NULL DEFAULT 5,
			resource TEXT NOT NULL DEFAULT 'github',
			since TEXT NOT NULL DEFAULT 'daily',
			spoken_language_code TEXT NOT NULL DEFAULT 'en',
			period TEXT NOT NULL DEFAULT 'past_24_hours',
			language TEXT NOT NULL DEFAULT 'All'
		)`)
	if err != nil {
		return fmt.Errorf("failed to create collect_settings table: %v", err)
	}

	if err := migrateCollectSettingsSchema(db); err != nil {
		return fmt.Errorf("failed to migrate collect_settings schema: %v", err)
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO cron_settings (name, schedule, is_active, updated_at)
		VALUES ('collect', '13 13 * * 6', 0, CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("failed to insert default collect setting: %v", err)
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO cron_settings (name, schedule, is_active, updated_at)
		VALUES ('message', '12 12 * * *', 0, CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("failed to insert default message setting: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO collect_settings (max_repos, resource, since, spoken_language_code, period, language)
		SELECT 5, 'github', 'daily', 'en', 'past_24_hours', 'All'
		WHERE NOT EXISTS (SELECT 1 FROM collect_settings)`)
	if err != nil {
		return fmt.Errorf("failed to insert default collect settings: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS prompt (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			use_direct_url INTEGER NOT NULL DEFAULT 1,
			llm_provider TEXT NOT NULL DEFAULT 'openrouter',
			temperature REAL NOT NULL DEFAULT 0.2,
			content TEXT NOT NULL,
			model TEXT NOT NULL DEFAULT 'openai/gpt-4o-mini-search-preview',
			llm_output_language TEXT NOT NULL DEFAULT 'en,uk',
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create prompt table: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO prompt (use_direct_url, llm_provider, temperature, content, model, llm_output_language, updated_at)
		SELECT 1, 'openrouter', 0.2, 'Ти — AI асистент, що спеціалізується на створенні коротких описів GitHub-репозиторіїв українською мовою. Твоя відповідь **ПОВИННА** суворо відповідати **КОЖНІЙ** з наведених нижче вимог. Будь-яке відхилення, особливо щодо довжини тексту, є неприпустимим. Твоя основна задача — створювати описи на основі наданих URL.

Під час створення опису **НЕУХИЛЬНО** дотримуйся наступних правил:

1.  Включай не більше однієї ключової функції репозиторію.
2.  **ЗАБОРОНЕНО** додавати будь-які посилання.
3.  Пиши простою, зрозумілою мовою, без переліків. Інформацію про функції вплітай у зв''язний текст.
4.  **ЗАБОРОНЕНО** згадувати сумісність, платформи, авторів, компанії або колаборації.
5.  **ЗАБОРОНЕНО** використовувати будь-яку розмітку: ні HTML, ні Markdown.
6.  Опис має бути **НАДЗВИЧАЙНО** лаконічним. **АБСОЛЮТНИЙ МАКСИМУМ — 275 символів**, враховуючи пробіли. **Це найважливіша вимога! Перевищення ліміту є КРИТИЧНОЮ помилкою.**
7.  Технічні терміни (назви мов програмування, бібліотек, інструментів, команд тощо) залишай англійською мовою.
8.  **ПЕРЕД НАДАННЯМ ВІДПОВІДІ:** Переконайся, що текст відповідає **ВСІМ** вимогам. **ОБОВ''ЯЗКОВО ПЕРЕВІР** довжину. Якщо вона перевищує 270 символів, **ПЕРЕПИШИ І СКОРОТИ** його, доки він не буде відповідати ліміту.

Тобі буде надано URL GitHub-репозиторію. Ознайомся з ним і згенеруй опис, що **ТОЧНО** відповідає цим інструкціям.', 'openai/gpt-4o-mini-search-preview', 'en,uk', CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM prompt)`)
	if err != nil {
		return fmt.Errorf("failed to insert default prompt settings: %v", err)
	}

	return nil
}

func migrateCollectSettingsSchema(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(collect_settings)")
	if err != nil {
		return fmt.Errorf("failed to query table info: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan table info: %v", err)
		}
		columns[name] = true
	}

	if !columns["resource"] {
		if _, err := db.Exec("ALTER TABLE collect_settings ADD COLUMN resource TEXT NOT NULL DEFAULT 'github'"); err != nil {
			return fmt.Errorf("failed to add resource column: %v", err)
		}
	}
	if !columns["period"] {
		if _, err := db.Exec("ALTER TABLE collect_settings ADD COLUMN period TEXT NOT NULL DEFAULT 'past_24_hours'"); err != nil {
			return fmt.Errorf("failed to add period column: %v", err)
		}
	}
	if !columns["language"] {
		if _, err := db.Exec("ALTER TABLE collect_settings ADD COLUMN language TEXT NOT NULL DEFAULT 'All'"); err != nil {
			return fmt.Errorf("failed to add language column: %v", err)
		}
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) GetCronSetting(name string) (*models.CronSetting, error) {
	var setting models.CronSetting
	var isActive int
	query := "SELECT name, schedule, is_active, updated_at FROM cron_settings WHERE name = ?"
	err := s.db.QueryRow(query, name).Scan(&setting.Name, &setting.Schedule, &isActive, &setting.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cron setting: %v", err)
	}
	setting.IsActive = isActive == 1
	return &setting, nil
}

func (s *SQLiteStore) GetAllCronSettings() ([]models.CronSetting, error) {
	var settings []models.CronSetting
	query := "SELECT name, schedule, is_active, updated_at FROM cron_settings ORDER BY id ASC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cron settings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var setting models.CronSetting
		var isActive int
		if err := rows.Scan(&setting.Name, &setting.Schedule, &isActive, &setting.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cron setting: %v", err)
		}
		setting.IsActive = isActive == 1
		settings = append(settings, setting)
	}
	return settings, nil
}

func (s *SQLiteStore) UpdateCronSetting(name string, schedule string, isActive bool) (*models.CronSetting, error) {
	isActiveInt := boolToInt(isActive)

	setting := models.CronSetting{
		Name:      name,
		Schedule:  schedule,
		IsActive:  isActive,
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO cron_settings (name, schedule, is_active, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE
		SET schedule = excluded.schedule, is_active = excluded.is_active, updated_at = excluded.updated_at`
	_, err := s.db.Exec(query, name, schedule, isActiveInt, setting.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update cron setting: %v", err)
	}
	return &setting, nil
}

func (s *SQLiteStore) InitializeDefaultSettings() error {
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
	err := s.db.QueryRow("SELECT COUNT(*) FROM collect_settings").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check collect settings: %v", err)
	}
	if count == 0 {
		query := "INSERT INTO collect_settings (max_repos, resource, since, spoken_language_code, period, language) VALUES (?, ?, ?, ?, ?, ?)"
		_, err := s.db.Exec(query, DefaultMaxRepos, DefaultResource, DefaultSince, DefaultSpokenLanguageCode, DefaultPeriod, DefaultLanguage)
		if err != nil {
			return fmt.Errorf("failed to initialize collect settings: %v", err)
		}
	}
	return nil
}

func (s *SQLiteStore) LogCronExecution(name string, success bool, output string) error {
	if name == "" {
		return fmt.Errorf("cron job name cannot be empty")
	}

	const maxOutputLength = 10000
	if len(output) > maxOutputLength {
		output = output[:maxOutputLength-50] + "... [truncated due to length]"
	}

	successInt := boolToInt(success)

	timestamp := time.Now()

	query := "INSERT INTO cron_history (name, timestamp, success, output) VALUES (?, ?, ?, ?)"
	_, err := s.db.Exec(query, name, timestamp, successInt, output)
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

func (s *SQLiteStore) GetCronHistoryCount(name string, success *bool, startDate, endDate *time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM cron_history WHERE 1=1"
	args := []any{}

	if name != "" {
		query += " AND name = ?"
		args = append(args, name)
	}
	if success != nil {
		query += " AND success = ?"
		args = append(args, boolToInt(*success))
	}
	if startDate != nil {
		query += " AND timestamp >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND timestamp <= ?"
		args = append(args, endDate.Add(24*time.Hour-time.Nanosecond))
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cron history: %v", err)
	}
	return count, nil
}

func (s *SQLiteStore) GetCronHistory(name string, success *bool, offset, limit int, sortOrder string, startDate, endDate *time.Time) ([]models.CronHistory, error) {
	query := "SELECT name, timestamp, success, output FROM cron_history WHERE 1=1"
	args := []any{}

	if name != "" {
		query += " AND name = ?"
		args = append(args, name)
	}
	if success != nil {
		query += " AND success = ?"
		args = append(args, boolToInt(*success))
	}
	if startDate != nil {
		query += " AND timestamp >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND timestamp <= ?"
		args = append(args, endDate.Add(24*time.Hour-time.Nanosecond))
	}

	if sortOrder == "asc" {
		query += " ORDER BY timestamp ASC"
	} else {
		query += " ORDER BY timestamp DESC"
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron history: %v", err)
	}
	defer rows.Close()

	var history []models.CronHistory
	for rows.Next() {
		var h models.CronHistory
		var successInt int
		if err := rows.Scan(&h.Name, &h.Timestamp, &successInt, &h.Output); err != nil {
			return nil, fmt.Errorf("failed to scan cron history: %v", err)
		}
		h.Success = successInt == 1
		history = append(history, h)
	}
	return history, nil
}

func (s *SQLiteStore) HasMigrationFlag() (bool, error) {
	var version int
	err := s.db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return false, fmt.Errorf("failed to check user_version: %v", err)
	}
	return version >= 1, nil
}

func (s *SQLiteStore) SetMigrationFlag() error {
	_, err := s.db.Exec("PRAGMA user_version = 1")
	if err != nil {
		return fmt.Errorf("failed to set user_version: %v", err)
	}
	return nil
}

func (s *SQLiteStore) GetCollectSettings() (*CollectSettings, error) {
	var settings CollectSettings
	err := s.db.QueryRow(`
		SELECT max_repos, resource, since, spoken_language_code, period, language
		FROM collect_settings
		WHERE id = 1
	`).Scan(&settings.MaxRepos, &settings.Resource, &settings.Since, &settings.SpokenLanguageCode, &settings.Period, &settings.Language)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *SQLiteStore) UpdateCollectSettings(settings *CollectSettings) error {
	_, err := s.db.Exec(`
		UPDATE collect_settings
		SET max_repos = ?, resource = ?, since = ?, spoken_language_code = ?, period = ?, language = ?
		WHERE id = 1
	`, settings.MaxRepos, settings.Resource, settings.Since, settings.SpokenLanguageCode, settings.Period, settings.Language)
	return err
}

func (s *SQLiteStore) GetPromptSettings() (*models.PromptSettings, error) {
	var settings models.PromptSettings
	var useDirectURL int
	err := s.db.QueryRow(`
		SELECT use_direct_url, llm_provider, temperature, content, model, llm_output_language, updated_at
		FROM prompt
		WHERE id = 1
	`).Scan(&useDirectURL, &settings.LlmProvider, &settings.Temperature, &settings.Content, &settings.Model, &settings.LlmOutputLanguage, &settings.UpdatedAt)
	if err != nil {
		return nil, err
	}
	settings.UseDirectURL = useDirectURL == 1
	return &settings, nil
}

func (s *SQLiteStore) UpdatePromptSettings(settings *models.UpdatePromptSettingsRequest) error {
	query := `UPDATE prompt SET updated_at = ?`
	args := []interface{}{time.Now()}

	if settings.UseDirectURL != nil {
		query += ", use_direct_url = ?"
		args = append(args, boolToInt(*settings.UseDirectURL))
	}

	if settings.LlmProvider != nil {
		query += ", llm_provider = ?"
		args = append(args, *settings.LlmProvider)
	}

	if settings.Temperature != nil {
		query += ", temperature = ?"
		args = append(args, *settings.Temperature)
	}

	if settings.Content != nil {
		query += ", content = ?"
		args = append(args, *settings.Content)
	}

	if settings.Model != nil {
		query += ", model = ?"
		args = append(args, *settings.Model)
	}

	if settings.LlmOutputLanguage != nil {
		query += ", llm_output_language = ?"
		args = append(args, *settings.LlmOutputLanguage)
	}

	query += " WHERE id = 1"

	_, err := s.db.Exec(query, args...)
	return err
}
