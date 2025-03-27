package main

import (
	"content-maestro/internal/api"
	"content-maestro/internal/logger"
	"content-maestro/internal/middleware"
	"content-maestro/internal/schedule"
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

	imgPath := "tmp/gh_project_img"
	if err := os.MkdirAll(imgPath, 0777); err != nil {
		log.Error("Error creating image directory: %v", err)
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

	mux := http.NewServeMux()

	mux.Handle("/api/crons", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.GetCrons)))))
	mux.Handle("/api/crons/collect/schedule", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))))
	mux.Handle("/api/crons/message/schedule", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateSchedule)))))
	mux.Handle("/api/crons/collect/status", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))))
	mux.Handle("/api/crons/message/status", middleware.LoggingMiddleware(middleware.CorsMiddleware(middleware.AuthMiddleware(http.HandlerFunc(cronAPI.UpdateStatus)))))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Debugf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Error("Error starting server: %v", err)
	}
}
