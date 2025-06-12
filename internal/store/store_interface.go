package store

import (
	"content-maestro/internal/models"
	"time"
)

type StoreInterface interface {
	Close() error
	InitializeDefaultSettings() error
	GetCronSetting(name string) (*models.CronSetting, error)
	GetAllCronSettings() ([]models.CronSetting, error)
	UpdateCronSetting(name string, schedule string, isActive bool) (*models.CronSetting, error)
	LogCronExecution(name string, success bool, output string) error
	GetCronHistoryCount(name string, success *bool, startDate, endDate *time.Time) (int, error)
	GetCronHistory(name string, success *bool, offset, limit int, sortOrder string, startDate, endDate *time.Time) ([]models.CronHistory, error)
	GetCollectSettings() (*CollectSettings, error)
	UpdateCollectSettings(settings *CollectSettings) error
}
