package utils

import (
	"os"
)

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			log.Error(err)
		}
		log.Debug("Dir created successfully!")
	}
}
