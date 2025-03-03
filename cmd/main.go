package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"content-maestro/internal/logger"
	"content-maestro/internal/scheduler"
	"content-maestro/internal/utils"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if os.Getenv("APP_VERSION") == "" {
		os.Setenv("APP_VERSION", "dev")
	}
}

func main() {
	log := logger.NewLogger()
	defer log.Sync()

	utils.CreateDirIfNotExist("tmp/gh_project_img")

	log.Debug("content-maestro application starting...")
	log.Debug("APP_VERSION:", os.Getenv("APP_VERSION"))
	log.Debug("environment loaded successfully")
	log.Debug("starting scheduler...")

	s := scheduler.StartCron()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Debug("received signal: ", sig)
	log.Debug("shutting down...")

	s.Stop()
}
