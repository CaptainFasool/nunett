package repositories_gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/db/repositories"
)

// TestHandleDBError tests the handleDBError function for proper error translation.
// It asserts that different GORM errors are correctly translated into corresponding custom repository errors.
func TestHandleDBError(t *testing.T) {
	// Test case: GORM ErrRecordNotFound should result in NotFoundError
	err := handleDBError(gorm.ErrRecordNotFound)
	assert.Equal(t, repositories.NotFoundError, err)

	// Test case: GORM ErrInvalidData should result in InvalidDataError
	err = handleDBError(gorm.ErrInvalidData)
	assert.Equal(t, repositories.InvalidDataError, err)

	// Test case: GORM ErrInvalidDB should result in DatabaseError
	err = handleDBError(gorm.ErrInvalidDB)
	assert.Equal(t, repositories.DatabaseError, err)
}
