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
	LogCronExecution(name string, status int, output string) error
	GetCronHistoryCount(name string, status *int, startDate, endDate *time.Time) (int, error)
	GetCronHistory(name string, status *int, offset, limit int, sortOrder string, startDate, endDate *time.Time) ([]models.CronHistory, error)
	GetCollectSettings() (*CollectSettings, error)
	UpdateCollectSettings(settings *CollectSettings) error
	GetPromptSettings() (*models.PromptSettings, error)
	UpdatePromptSettings(settings *models.UpdatePromptSettingsRequest) error
	GetAPIConfig(name string) (*models.APIConfigModel, error)
	GetAllAPIConfigs() ([]models.APIConfigModel, error)
	CreateAPIConfig(config *models.CreateAPIConfigRequest) (*models.APIConfigModel, error)
	UpdateAPIConfig(name string, config *models.UpdateAPIConfigRequest) (*models.APIConfigModel, error)
	DeleteAPIConfig(name string) error
}
