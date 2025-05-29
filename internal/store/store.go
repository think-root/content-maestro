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
			Schedule:  "13 13 * * 6",
			IsActive:  true,
			UpdatedAt: time.Now(),
		},
		{
			Name:      "message",
			Schedule:  "12 12 * * *",
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

	var collectSettings CollectSettings
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("collect_settings"))
		if err == badger.ErrKeyNotFound {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &collectSettings)
		})
	})

	if err == badger.ErrKeyNotFound {
		defaultCollectSettings := CollectSettings{
			MaxRepos:           5,
			Since:              "daily",
			SpokenLanguageCode: "en",
		}
		return s.db.Update(func(txn *badger.Txn) error {
			data, err := json.Marshal(defaultCollectSettings)
			if err != nil {
				return err
			}
			return txn.Set([]byte("collect_settings"), data)
		})
	}

	return err
}

func (s *Store) LogCronExecution(name string, success bool, errorMsg string, message string) error {
	history := models.CronHistory{
		Name:      name,
		Timestamp: time.Now(),
		Success:   success,
		Error:     errorMsg,
		Message:   message,
	}

	data, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal cron history: %v", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("cron_history:" + name + ":" + time.Now().Format(time.RFC3339))
		return txn.Set(key, data)
	})
}

// GetCronHistoryCount returns the total count of cron history records matching the filters
func (s *Store) GetCronHistoryCount(name string, success *bool, startDate, endDate *time.Time) (int, error) {
	count := 0

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("cron_history:")
		if name != "" {
			prefix = []byte("cron_history:" + name)
		}

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var hist models.CronHistory
				if err := json.Unmarshal(val, &hist); err != nil {
					return err
				}
				// Filter by name if specified
				if name != "" && hist.Name != name {
					return nil
				}
				// Filter by success if specified
				if success != nil && hist.Success != *success {
					return nil
				}
				// Filter by date range if specified
				if startDate != nil && hist.Timestamp.Before(*startDate) {
					return nil
				}
				if endDate != nil && hist.Timestamp.After(endDate.Add(24*time.Hour-time.Nanosecond)) {
					return nil
				}
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count cron history: %v", err)
	}

	return count, nil
}

func (s *Store) GetCronHistory(name string, success *bool, offset, limit int, sortOrder string, startDate, endDate *time.Time) ([]models.CronHistory, error) {
	var allHistory []models.CronHistory

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("cron_history:")
		if name != "" {
			prefix = []byte("cron_history:" + name)
		}

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var hist models.CronHistory
				if err := json.Unmarshal(val, &hist); err != nil {
					return err
				}
				// Filter by name if specified
				if name != "" && hist.Name != name {
					return nil
				}
				// Filter by success if specified
				if success != nil && hist.Success != *success {
					return nil
				}
				// Filter by date range if specified
				if startDate != nil && hist.Timestamp.Before(*startDate) {
					return nil
				}
				if endDate != nil && hist.Timestamp.After(endDate.Add(24*time.Hour-time.Nanosecond)) {
					return nil
				}
				allHistory = append(allHistory, hist)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get cron history: %v", err)
	}

	// Sort by timestamp using sort package
	if sortOrder == "asc" {
		// Sort ascending (oldest first)
		for i := 0; i < len(allHistory); i++ {
			for j := i + 1; j < len(allHistory); j++ {
				if allHistory[i].Timestamp.After(allHistory[j].Timestamp) {
					allHistory[i], allHistory[j] = allHistory[j], allHistory[i]
				}
			}
		}
	} else {
		// Sort descending (newest first) - default
		for i := 0; i < len(allHistory); i++ {
			for j := i + 1; j < len(allHistory); j++ {
				if allHistory[i].Timestamp.Before(allHistory[j].Timestamp) {
					allHistory[i], allHistory[j] = allHistory[j], allHistory[i]
				}
			}
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(allHistory) {
		return []models.CronHistory{}, nil
	}
	if end > len(allHistory) {
		end = len(allHistory)
	}

	return allHistory[start:end], nil
}
