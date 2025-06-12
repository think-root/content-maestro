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
	if err := godotenv.Load(); err != nil {
		log.Error("Error loading .env file")
		return
	}

	var storeInstance store.StoreInterface
	var err error

	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDBName := os.Getenv("POSTGRES_DB")

	if pgHost == "" || pgPort == "" || pgUser == "" || pgDBName == "" {
		log.Error("PostgreSQL environment variables are missing (POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_DB). These are required to run the application.")
		return
	}

	log.Debug("Initializing PostgreSQL store")
	pgStore, err := store.NewPostgresStore(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		log.Error("Error initializing PostgreSQL store: %v", err)
		return
	}
	storeInstance = pgStore
	defer storeInstance.Close()
	log.Debug("PostgreSQL store initialized")

	if err := os.MkdirAll("tmp/gh_project_img", 0777); err != nil {
		log.Error("Error creating tmp/gh_project_img directory: %v", err)
		return
	}

	if err := storeInstance.InitializeDefaultSettings(); err != nil {
		log.Error("Error initializing default settings: %v", err)
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
		log.Error("Error starting server: %v", err)
	}
}
