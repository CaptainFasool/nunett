package repositories_gorm

import (
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/db/repositories"
)

const structFieldNameDeletedAt = "DeletedAt"

// handleDBError is a utility function that translates GORM database errors into custom repository errors.
// It takes a GORM database error as input and returns a corresponding custom error from the repositories package.
func handleDBError(err error) error {
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			// Return NotFoundError for record not found errors
			return repositories.NotFoundError
		case gorm.ErrInvalidData, gorm.ErrInvalidField, gorm.ErrInvalidValue:
			// Return InvalidDataError for various invalid data errors
			return repositories.InvalidDataError
		default:
			// Return DatabaseError for other unspecified database errors
			return repositories.DatabaseError
		}
	}
	return nil
}
