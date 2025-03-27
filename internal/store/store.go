package store

import (
	"encoding/json"
	"fmt"
	"time"

	"content-maestro/internal/models"

	"github.com/dgraph-io/badger/v3"
)

type Store struct {
	db *badger.DB
}

func NewStore(dbPath string) (*Store, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %v", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetCronSetting(name string) (*models.CronSetting, error) {
	var setting models.CronSetting

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("cron:" + name))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &setting)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get cron setting: %v", err)
	}

	return &setting, nil
}

func (s *Store) GetAllCronSettings() ([]models.CronSetting, error) {
	var settings []models.CronSetting

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("cron:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var setting models.CronSetting
				if err := json.Unmarshal(val, &setting); err != nil {
					return err
				}
				settings = append(settings, setting)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all cron settings: %v", err)
	}

	return settings, nil
}

func (s *Store) UpdateCronSetting(name string, schedule string, isActive bool) (*models.CronSetting, error) {
	setting := models.CronSetting{
		Name:      name,
		Schedule:  schedule,
		IsActive:  isActive,
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(setting)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cron setting: %v", err)
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("cron:"+name), data)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update cron setting: %v", err)
	}

	return &setting, nil
}

func (s *Store) InitializeDefaultSettings() error {
	defaults := []models.CronSetting{
		{
			Name:      "collect",
			Schedule:  "0 13 * * 6",
			IsActive:  true,
			UpdatedAt: time.Now(),
		},
		{
			Name:      "message",
			Schedule:  "12 10 * * *",
			IsActive:  true,
			UpdatedAt: time.Now(),
		},
	}

	for _, setting := range defaults {
		exists, err := s.GetCronSetting(setting.Name)
		if err != nil {
			return fmt.Errorf("failed to check existing setting: %v", err)
		}

		if exists == nil {
			if _, err := s.UpdateCronSetting(setting.Name, setting.Schedule, setting.IsActive); err != nil {
				return fmt.Errorf("failed to initialize default setting: %v", err)
			}
		}
	}

	return nil
}
