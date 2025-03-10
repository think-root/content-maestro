package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirIfNotExist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "create_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Create new directory", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "new_dir")

		if _, err := os.Stat(newDir); !os.IsNotExist(err) {
			t.Fatalf("Test directory already exists before test: %s", newDir)
		}

		CreateDirIfNotExist(newDir)

		if _, err := os.Stat(newDir); os.IsNotExist(err) {
			t.Errorf("Directory was not created: %s", newDir)
		}
	})

	t.Run("Directory already exists", func(t *testing.T) {
		existingDir := filepath.Join(tempDir, "existing_dir")

		err := os.Mkdir(existingDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		CreateDirIfNotExist(existingDir)

		if _, err := os.Stat(existingDir); os.IsNotExist(err) {
			t.Errorf("Existing directory no longer exists")
		}
	})
}

func TestCreateDirIfNotExist_Success(t *testing.T) {
	temp := t.TempDir()
	newDir := filepath.Join(temp, "testDir")
	CreateDirIfNotExist(newDir)
	info, err := os.Stat(newDir)
	if err != nil {
		t.Errorf("Directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Not a directory")
	}
	CreateDirIfNotExist(newDir)
}

func TestCreateDirIfNotExist_Error(t *testing.T) {
	temp := t.TempDir()
	errPath := filepath.Join(temp, "noPermission", "sub")
	os.Mkdir(temp+"/noPermission", 0o400)
	CreateDirIfNotExist(errPath)
}
