package store

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
)

type CollectSettings struct {
	MaxRepos           int    `json:"max_repos"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
}

func (s *Store) GetCollectSettings() (*CollectSettings, error) {
	var settings CollectSettings
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("collect_settings"))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &settings)
		})
	})
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *Store) UpdateCollectSettings(settings *CollectSettings) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(settings)
		if err != nil {
			return err
		}
		return txn.Set([]byte("collect_settings"), data)
	})
}
