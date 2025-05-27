package main

import (
	"content-maestro/internal/logger"
	"content-maestro/internal/middleware"
	"content-maestro/internal/schedule"
	"content-maestro/internal/server"
	"content-maestro/internal/store"
	"net/http"
	"os"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error("Error loading .env file")
		return
	}

	dbPath := "data/badger"
	if err := os.MkdirAll(dbPath, 0777); err != nil {
		log.Error("Error creating database directory: %v", err)
		return
	}
	store, err := store.NewStore(dbPath)
	if err != nil {
		log.Error("Error initializing store: %v", err)
		return
	}
	defer store.Close()

	if err := os.MkdirAll("tmp/gh_project_img", 0777); err != nil {
		log.Error("Error creating tmp/gh_project_img directory: %v", err)
		return
	}

	if err := store.InitializeDefaultSettings(); err != nil {
		log.Error("Error initializing default settings: %v", err)
		return
	}

	settings, _ := store.GetAllCronSettings()
	schedulers := make(map[string]*gocron.Scheduler)

	for _, setting := range settings {
		schedulers[setting.Name] = schedule.NewScheduler(store, setting.Name, setting.Schedule)
	}

	jobs := schedule.InitJobs(store)
	cronAPI := server.NewCronAPI(store, schedulers, jobs)

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
