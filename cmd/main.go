package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"content-maestro/internal/logger"
	"content-maestro/internal/schedule"
	"content-maestro/internal/utils"
)

var log = logger.NewLogger()

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Debug("environment loaded successfully")
	if os.Getenv("APP_VERSION") == "" {
		os.Setenv("APP_VERSION", "dev")
	}
}

func main() {
	utils.CreateDirIfNotExist("tmp/gh_project_img")

	log.Debug("content-maestro application starting...")
	log.Debug("APP_VERSION:", os.Getenv("APP_VERSION"))
	log.Debug("starting schedules...")

	s1 := schedule.MessageCron()
	s2 := schedule.CollectCron()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan

	log.Debug("received signal: ", sig)
	log.Debug("shutting down...")

	s1.Stop()
	s2.Stop()
}
