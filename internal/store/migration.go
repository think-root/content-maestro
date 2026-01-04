package store

import (
	"content-maestro/internal/logger"
	"content-maestro/internal/models"
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var log = logger.NewLogger()

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func GetPostgresConfigFromEnv() *PostgresConfig {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	if host == "" || port == "" || user == "" || password == "" || dbName == "" {
		return nil
	}

	return &PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
	}
}

func IsPostgresAvailable(config *PostgresConfig) bool {
	if config == nil {
		return false
	}

	address := net.JoinHostPort(config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Debugf("PostgreSQL server not reachable at %s: %v", address, err)
		return false
	}
	conn.Close()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Debugf("Failed to open PostgreSQL connection: %v", err)
		return false
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Debugf("Failed to ping PostgreSQL: %v", err)
		return false
	}

	return true
}

func MigrateFromPostgres(sqliteStore *SQLiteStore, config *PostgresConfig) error {
	log.Debug("Starting migration from PostgreSQL to SQLite...")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	pgDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	if err := pgDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	log.Debug("Connected to PostgreSQL successfully")

	tx, err := sqliteStore.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start SQLite transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = migrateCronSettings(pgDB, tx); err != nil {
		return fmt.Errorf("failed to migrate cron_settings: %v", err)
	}

	if err = migrateCronHistory(pgDB, tx); err != nil {
		return fmt.Errorf("failed to migrate cron_history: %v", err)
	}

	if err = migrateCollectSettings(pgDB, tx); err != nil {
		return fmt.Errorf("failed to migrate collect_settings: %v", err)
	}

	if err = migratePromptSettings(pgDB, tx); err != nil {
		return fmt.Errorf("failed to migrate prompt_settings: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %v", err)
	}

	if err = sqliteStore.SetMigrationFlag(); err != nil {
		return fmt.Errorf("failed to set migration flag: %v", err)
	}

	log.Debug("Migration from PostgreSQL to SQLite completed successfully")
	return nil
}

func migrateCronSettings(pgDB *sql.DB, tx *sql.Tx) error {
	log.Debug("Migrating cron_settings...")

	var tableExists bool
	err := pgDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'maestro_cron_settings'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if maestro_cron_settings table exists: %v", err)
	}

	if !tableExists {
		log.Debug("maestro_cron_settings table does not exist in PostgreSQL, skipping...")
		return nil
	}

	rows, err := pgDB.Query("SELECT name, schedule, is_active, updated_at FROM maestro_cron_settings")
	if err != nil {
		return fmt.Errorf("failed to query maestro_cron_settings: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var setting models.CronSetting
		if err := rows.Scan(&setting.Name, &setting.Schedule, &setting.IsActive, &setting.UpdatedAt); err != nil {
			return fmt.Errorf("failed to scan cron_setting: %v", err)
		}

		isActiveInt := 0
		if setting.IsActive {
			isActiveInt = 1
		}

		_, err = tx.Exec(`
			INSERT INTO cron_settings (name, schedule, is_active, updated_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE
			SET schedule = excluded.schedule, is_active = excluded.is_active, updated_at = excluded.updated_at`,
			setting.Name, setting.Schedule, isActiveInt, setting.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to upsert cron_setting into SQLite: %v", err)
		}

		count++
	}

	log.Debugf("Migrated %d cron_settings records", count)
	return nil
}

func migrateCronHistory(pgDB *sql.DB, tx *sql.Tx) error {
	log.Debug("Migrating cron_history...")

	var tableExists bool
	err := pgDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'maestro_cron_history'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if maestro_cron_history table exists: %v", err)
	}

	if !tableExists {
		log.Debug("maestro_cron_history table does not exist in PostgreSQL, skipping...")
		return nil
	}

	_, err = tx.Exec("DELETE FROM cron_history")
	if err != nil {
		return fmt.Errorf("failed to truncate cron_history: %v", err)
	}
	log.Debug("Truncated cron_history table before migration")

	rows, err := pgDB.Query("SELECT name, timestamp, success, COALESCE(output, '') FROM maestro_cron_history ORDER BY id ASC")
	if err != nil {
		return fmt.Errorf("failed to query maestro_cron_history: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var h models.CronHistory
		if err := rows.Scan(&h.Name, &h.Timestamp, &h.Success, &h.Output); err != nil {
			return fmt.Errorf("failed to scan maestro_cron_history: %v", err)
		}

		successInt := 0
		if h.Success {
			successInt = 1
		}

		_, err := tx.Exec(
			"INSERT INTO cron_history (name, timestamp, success, output) VALUES (?, ?, ?, ?)",
			h.Name, h.Timestamp, successInt, h.Output,
		)
		if err != nil {
			return fmt.Errorf("failed to insert cron_history into SQLite: %v", err)
		}

		count++
	}

	log.Debugf("Migrated %d cron_history records", count)
	return nil
}

func migrateCollectSettings(pgDB *sql.DB, tx *sql.Tx) error {
	log.Debug("Migrating collect_settings...")

	var tableExists bool
	err := pgDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'maestro_collect_settings'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if maestro_collect_settings table exists: %v", err)
	}

	if !tableExists {
		log.Debug("maestro_collect_settings table does not exist in PostgreSQL, skipping...")
		return nil
	}

	var settings CollectSettings
	err = pgDB.QueryRow("SELECT max_repos, since, spoken_language_code FROM maestro_collect_settings LIMIT 1").
		Scan(&settings.MaxRepos, &settings.Since, &settings.SpokenLanguageCode)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("No maestro_collect_settings found in PostgreSQL, skipping...")
			return nil
		}
		return fmt.Errorf("failed to query maestro_collect_settings: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE collect_settings
		SET max_repos = ?, since = ?, spoken_language_code = ?
		WHERE id = 1
	`, settings.MaxRepos, settings.Since, settings.SpokenLanguageCode)
	if err != nil {
		return fmt.Errorf("failed to update collect_settings in SQLite: %v", err)
	}

	log.Debug("Migrated collect_settings successfully")
	return nil
}

func migratePromptSettings(pgDB *sql.DB, tx *sql.Tx) error {
	log.Debug("Migrating prompt settings...")

	var tableExists bool
	err := pgDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'think_prompt'
		)
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if think_prompt table exists: %v", err)
	}

	if !tableExists {
		log.Debug("think_prompt table does not exist in PostgreSQL, skipping...")
		return nil
	}

	var settings models.PromptSettings
	err = pgDB.QueryRow(`
		SELECT use_direct_url, llm_provider, temperature, content, model, llm_output_language, updated_at 
		FROM think_prompt 
		LIMIT 1
	`).Scan(&settings.UseDirectURL, &settings.LlmProvider, &settings.Temperature, &settings.Content, &settings.Model, &settings.LlmOutputLanguage, &settings.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("No think_prompt settings found in PostgreSQL, skipping...")
			return nil
		}
		return fmt.Errorf("failed to query think_prompt settings: %v", err)
	}

	useDirectURLInt := 0
	if settings.UseDirectURL {
		useDirectURLInt = 1
	}

	_, err = tx.Exec(`
		UPDATE prompt 
		SET use_direct_url = ?, llm_provider = ?, temperature = ?, content = ?, model = ?, llm_output_language = ?, updated_at = ?
		WHERE id = 1
	`, useDirectURLInt, settings.LlmProvider, settings.Temperature, settings.Content, settings.Model, settings.LlmOutputLanguage, settings.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update prompt settings in SQLite: %v", err)
	}

	log.Debug("Migrated prompt settings successfully")
	return nil
}

func ShouldMigrate(sqliteStore *SQLiteStore, config *PostgresConfig) (bool, error) {
	hasMigrationFlag, err := sqliteStore.HasMigrationFlag()
	if err != nil {
		return false, fmt.Errorf("failed to check migration flag: %v", err)
	}

	if hasMigrationFlag {
		log.Debug("Migration flag found, skipping migration")
		return false, nil
	}

	if config == nil {
		log.Debug("PostgreSQL config not found in environment")
		return false, nil
	}

	if !IsPostgresAvailable(config) {
		log.Debug("PostgreSQL is not available, skipping migration")
		return false, nil
	}

	return true, nil
}
