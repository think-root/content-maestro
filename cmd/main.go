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

	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDBName := os.Getenv("POSTGRES_DB")

	if pgHost == "" || pgPort == "" || pgUser == "" || pgDBName == "" {
		log.Error("PostgreSQL environment variables are missing:")
		if pgHost == "" { log.Error("- POSTGRES_HOST is empty") }
		if pgPort == "" { log.Error("- POSTGRES_PORT is empty") }
		if pgUser == "" { log.Error("- POSTGRES_USER is empty") }
		if pgDBName == "" { log.Error("- POSTGRES_DB is empty") }
		log.Error("These are required to run the application.")
		return
	}

	log.Debugf("Attempting to connect to PostgreSQL...")
	pgStore, err := store.NewPostgresStore(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		log.Errorf("Error initializing PostgreSQL store: %v", err)
		log.Errorf("Connection details: host=%s port=%s user=%s dbname=%s", pgHost, pgPort, pgUser, pgDBName)
		return
	}
	storeInstance = pgStore
	defer storeInstance.Close()
	log.Debug("PostgreSQL store initialized successfully")

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
	mux.Handle("/api/collect-settings", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.GetCollectSettings)))))
	mux.Handle("/api/collect-settings/update", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateCollectSettings)))))
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
