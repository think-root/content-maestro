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
		// Subdirectories are owned by whoever created them - a manual retry keeps
		// its images in one so this cleanup cannot delete a file that is still
		// being uploaded.
		if info, err := os.Stat(file); err == nil && info.IsDir() {
			continue
		}

		// A file that another writer removed in the meantime is not a failure:
		// treating it as one used to turn a successful run into a reported one.
		if err := remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	log.Debug("Files removed successfully!")
	return nil
}
