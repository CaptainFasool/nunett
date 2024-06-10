package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
)

func GetDirectorySize(fs afero.Fs, path string) (int64, error) {
	var size int64
	err := afero.Walk(fs, path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to calculate volume size: %w", err)
	}

	return size, nil
}

// CopyFile copies file from an afero FS to another afero FS. It's useful when
// testing with in-memory file systems.
func CopyFile(srcFs afero.Fs, dstFs afero.Fs, srcPath string, dstPath string) error {
	// Open the source file
	srcFile, err := srcFs.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := dstFs.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}
