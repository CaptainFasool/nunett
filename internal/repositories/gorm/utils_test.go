package repositories_gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/internal/repositories"
	"gitlab.com/nunet/device-management-service/models"
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

// TestUpdateField tests the updateField function for updating struct fields using reflection.
// It covers cases where the input can be a struct or a pointer to a struct.
func TestUpdateField(t *testing.T) {
	// Test case: Updating a field in a struct (not a pointer)
	modified1, err := updateField(models.FreeResources{Vcpu: 1}, "Vcpu", 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, modified1.Vcpu)

	// Test case: Updating a field in a struct (using a pointer)
	modified2, err := updateField(&models.FreeResources{Vcpu: 1}, "Vcpu", 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, modified2.Vcpu)

	// Test case: Attempting to update a non-existent field results in an error
	_, err = updateField(models.FreeResources{Vcpu: 1}, "VCPU", 2)
	assert.Error(t, err)

	// Test case: Attempting to update with an incompatible value results in an error
	_, err = updateField(models.FreeResources{Vcpu: 1}, "Vcpu", "a")
	assert.Error(t, err)
}

// TestEmptyValue tests the isEmptyValue function for checking if a struct has non zero value.
// It asserts that the function correctly identifies empty and non-empty structs (or a pointer to a struct).
func TestEmptyValue(t *testing.T) {
	// Test case: nil should be considered empty
	assert.Equal(t, true, isEmptyValue(nil))

	// Test case: Empty struct and its pointer should be considered empty
	assert.Equal(t, true, isEmptyValue(models.FreeResources{}))
	assert.Equal(t, true, isEmptyValue(&models.FreeResources{}))

	// Test case: Struct and its pointer with non-zero field should not be considered empty
	assert.Equal(t, false, isEmptyValue(models.FreeResources{Vcpu: 1}))
	assert.Equal(t, false, isEmptyValue(&models.FreeResources{Vcpu: 1}))
}
