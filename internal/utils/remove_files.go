package utils

import (
	"os"
	"path/filepath"
)

func RemoveAllFilesInFolder(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return err
		}
	}
	log.Debug("Files removed successfully!")
	return nil
}
