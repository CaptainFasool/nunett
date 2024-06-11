package utils

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCopyFile(t *testing.T) {
	// Create two in-memory filesystems
	srcFs := afero.NewMemMapFs()
	dstFs := afero.NewMemMapFs()

	// Create and write a file to the source filesystem
	srcFilePath := "/alice.txt"
	expectedContent := []byte("alice likes bob")
	err := afero.WriteFile(srcFs, srcFilePath, expectedContent, 0644)
	assert.NoError(t, err, "Failed to write file to source filesystem")

	// Copy the file from the source to the destination filesystem
	dstFilePath := "/alice_copy.txt"
	err = CopyFile(srcFs, dstFs, srcFilePath, dstFilePath)
	assert.NoError(t, err, "CopyFile returned an error")

	// Read and check the content of the copied file
	actualContent, err := afero.ReadFile(dstFs, dstFilePath)
	assert.NoError(t, err, "Failed to read file from destination filesystem")
	assert.Equal(t, expectedContent, actualContent, "Content of the copied file does not match the expected content")
}
