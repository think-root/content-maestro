package utils

import (
	"content-maestro/internal/logger"
	"io"
	"os"
)

var log = logger.NewLogger()

func CopyFile(sourceFile, targetFile string) error {
	srcFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buffer := make([]byte, 32*1024)
	_, err = io.CopyBuffer(dstFile, srcFile, buffer)
	if err != nil {
		return err
	}

	return nil
}
