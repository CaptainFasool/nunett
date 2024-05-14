package utils

import (
	"fmt"
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
