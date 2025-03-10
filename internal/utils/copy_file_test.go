package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		sourceData  string
		sourcePerm  os.FileMode
		targetPerm  os.FileMode
		expectError bool
	}{
		{
			name:        "successful copy",
			sourceData:  "test data",
			sourcePerm:  0644,
			targetPerm:  0644,
			expectError: false,
		},
		{
			name:        "empty file copy",
			sourceData:  "",
			sourcePerm:  0644,
			targetPerm:  0644,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			sourceFile := filepath.Join(tempDir, "source.txt")
			targetFile := filepath.Join(tempDir, "target.txt")

			err := os.WriteFile(sourceFile, []byte(tt.sourceData), tt.sourcePerm)
			if err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			err = CopyFile(sourceFile, targetFile)
			if (err != nil) != tt.expectError {
				t.Errorf("CopyFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				sourceContent, err := os.ReadFile(sourceFile)
				if err != nil {
					t.Fatalf("Failed to read source file: %v", err)
				}

				targetContent, err := os.ReadFile(targetFile)
				if err != nil {
					t.Fatalf("Failed to read target file: %v", err)
				}

				if string(sourceContent) != string(targetContent) {
					t.Errorf("File contents don't match. Got %s, want %s", string(targetContent), string(sourceContent))
				}
			}
		})
	}
}

func TestCopyFileErrors(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	t.Run("source file does not exist", func(t *testing.T) {
		err := CopyFile(nonExistentFile, targetFile)
		if err == nil {
			t.Error("Expected error when source file doesn't exist")
		}
	})

	t.Run("invalid target path", func(t *testing.T) {
		sourceFile := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourceFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		invalidTarget := filepath.Join(tempDir, "nonexistent", "target.txt")
		err = CopyFile(sourceFile, invalidTarget)
		if err == nil {
			t.Error("Expected error when target path is invalid")
		}
	})
}
