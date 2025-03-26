package main

import (
	"content-maestro/internal/api"
	"content-maestro/internal/logger"
	"content-maestro/internal/middleware"
	"content-maestro/internal/schedule"
	"content-maestro/internal/store"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error("Error loading .env file")
		return
	}

	dbPath := filepath.Join("data", "badger")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		log.Error("Error creating database directory: %v", err)
		return
	}

	store, err := store.NewStore(dbPath)
	if err != nil {
		log.Error("Error initializing store: %v", err)
		return
	}
	defer store.Close()

	if err := store.InitializeDefaultSettings(); err != nil {
		log.Error("Error initializing default settings: %v", err)
		return
	}

	collectScheduler := schedule.CollectCron(store)
	messageScheduler := schedule.MessageCron(store)

	cronAPI := api.NewCronAPI(store, map[string]*gocron.Scheduler{
		"collect": collectScheduler,
		"message": messageScheduler,
	})

	http.Handle("/api/crons", middleware.AuthMiddleware(http.HandlerFunc(cronAPI.GetCrons)))
	http.Handle("/api/crons/collect/schedule", middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))
	http.Handle("/api/crons/message/schedule", middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))
	http.Handle("/api/crons/collect/status", middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))
	http.Handle("/api/crons/message/status", middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Debugf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Error("Error starting server: %v", err)
	}
}
