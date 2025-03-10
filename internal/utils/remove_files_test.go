package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveAllFilesInFolder(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, error)
		wantErr bool
	}{
		{
			name: "successfully remove files",
			setup: func() (string, error) {
				tmpDir, err := os.MkdirTemp("", "test-remove-*")
				if err != nil {
					return "", err
				}

				testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
				for _, f := range testFiles {
					path := filepath.Join(tmpDir, f)
					if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
						return tmpDir, err
					}
				}
				return tmpDir, nil
			},
			wantErr: false,
		},
		{
			name: "empty directory",
			setup: func() (string, error) {
				return os.MkdirTemp("", "test-remove-empty-*")
			},
			wantErr: false,
		},
		{
			name: "non-existent directory",
			setup: func() (string, error) {
				return "/non/existent/directory", nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := tt.setup()
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}
			if dir != "/non/existent/directory" {
				defer os.RemoveAll(dir)
			}

			err = RemoveAllFilesInFolder(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveAllFilesInFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				files, err := filepath.Glob(filepath.Join(dir, "*"))
				if err != nil {
					t.Fatalf("failed to check directory contents: %v", err)
				}
				if len(files) > 0 {
					t.Errorf("directory not empty after removal, contains: %v", files)
				}
			}
		})
	}
}

func TestRemoveAllFilesInFolder_NonExistent(t *testing.T) {
	err := RemoveAllFilesInFolder("nonexistent_folder")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected error about nonexistent folder, got %v", err)
	}
}

func TestRemoveAllFilesInFolder_Success(t *testing.T) {
	testDir := t.TempDir()
	f, err := os.CreateTemp(testDir, "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	err = RemoveAllFilesInFolder(testDir)
	if err != nil {
		t.Errorf("unexpected error removing files: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(testDir, "*"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("files were not removed properly")
	}
}

func TestRemoveAllFilesInFolder_FailedRemove(t *testing.T) {
	testDir := t.TempDir()
	fp := filepath.Join(testDir, "locked")
	if err := os.WriteFile(fp, []byte("content"), 0400); err != nil {
		t.Fatalf("failed to write locked file: %v", err)
	}
	remove = func(name string) error {
		if name == fp {
			return errors.New("mock remove error")
		}
		return os.Remove(name)
	}
	defer func() { remove = os.Remove }()
	err := RemoveAllFilesInFolder(testDir)
	if err == nil || !strings.Contains(err.Error(), "mock remove error") {
		t.Errorf("expected mock remove error, got %v", err)
	}
}
