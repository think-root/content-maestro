package utils

import (
	"os"
	"path/filepath"
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
