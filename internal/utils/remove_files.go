package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

var remove = os.Remove

func RemoveAllFilesInFolder(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = remove(file)
		if err != nil {
			return err
		}
	}

	log.Debug("Files removed successfully!")
	return nil
}
