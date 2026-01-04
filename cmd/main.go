package main

import (
	"content-maestro/internal/logger"
	"content-maestro/internal/middleware"
	"content-maestro/internal/models"
	"content-maestro/internal/schedule"
	"content-maestro/internal/server"
	"content-maestro/internal/store"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger()

type StoreInterface interface {
	Close() error
	InitializeDefaultSettings() error
	GetAllCronSettings() ([]models.CronSetting, error)
}

func main() {
	log.Debug("Starting content-maestro application")

	if err := godotenv.Load(); err != nil {
		log.Errorf("Error loading .env file: %v", err)
		log.Debug("Continuing without .env file - will use environment variables")
	} else {
		log.Debug(".env file loaded successfully")
	}

	var storeInstance store.StoreInterface
	var err error

	sqlitePath := os.Getenv("SQLITE_DB_PATH")
	if sqlitePath == "" {
		// Use current working directory for a stable default path
		sqlitePath = filepath.Join(".", "data", "content-maestro.db")
		log.Debugf("SQLITE_DB_PATH not set, using default: %s", sqlitePath)
	}

	if err := os.MkdirAll(filepath.Dir(sqlitePath), 0755); err != nil {
		log.Errorf("Error creating database directory: %v", err)
		return
	}

	log.Debugf("Attempting to connect to SQLite database at %s...", sqlitePath)
	sqliteStore, err := store.NewSQLiteStore(sqlitePath)
	if err != nil {
		log.Errorf("Error initializing SQLite store: %v", err)
		return
	}
	storeInstance = sqliteStore
	defer storeInstance.Close()
	log.Debug("SQLite store initialized successfully")

	log.Debug("Creating tmp/gh_project_img directory")
	if err := os.MkdirAll("tmp/gh_project_img", 0777); err != nil {
		log.Errorf("Error creating tmp/gh_project_img directory: %v", err)
		return
	}
	log.Debug("Directory tmp/gh_project_img created successfully")

	if err := storeInstance.InitializeDefaultSettings(); err != nil {
		log.Errorf("Error initializing default settings: %v", err)
		return
	}
	schedulers := map[string]*gocron.Scheduler{
		"collect": schedule.CollectCron(storeInstance),
		"message": schedule.MessageCron(storeInstance),
	}

	jobs := schedule.InitJobs(storeInstance)

	cronAPI := server.NewCronAPI(storeInstance, schedulers, jobs)

	mux := http.NewServeMux()

	mux.Handle("/api/crons", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.GetCrons)))))
	mux.Handle("/api/crons/collect/schedule", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))))
	mux.Handle("/api/crons/message/schedule", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))))
	mux.Handle("/api/crons/collect/status", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))))
	mux.Handle("/api/crons/message/status", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))))
	mux.Handle("/api/collect-settings", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.HandleCollectSettings)))))
	mux.Handle("/api/prompt-settings", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.HandlePromptSettings)))))
	mux.Handle("/api/cron-history", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.GetCronHistory)))))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Debugf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Errorf("Error starting server: %v", err)
	}
}
