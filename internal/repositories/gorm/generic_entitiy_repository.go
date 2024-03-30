package repositories_gorm

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gitlab.com/nunet/device-management-service/internal/repositories"
)

// GenericEntityRepositoryGORM is a generic single entity repository implementation using GORM as an ORM.
// It is intended to be embedded in single entity model repositories to provide basic database operations.
type GenericEntityRepositoryGORM[T interface{}] struct {
	db      *gorm.DB    // db is the GORM database instance.
	pkField string      // pkField is the primary key field of the entity.
	pkValue interface{} // pkValue is the value of the primary key field.
}

// NewGenericEntityRepository creates a new instance of GenericEntityRepositoryGORM.
// It initializes and returns a repository with the provided GORM database, primary key field, and value.
func NewGenericEntityRepository[T interface{}](
	db *gorm.DB,
	pkField string,
	pkValue interface{},
) repositories.GenericEntityRepository[T] {
	return &GenericEntityRepositoryGORM[T]{db: db, pkField: pkField, pkValue: pkValue}
}

// Save creates or updates the record to the repository and returns the new/updated data.
func (repo *GenericEntityRepositoryGORM[T]) Save(ctx context.Context, data T) (T, error) {
	var updatedData T
	var err error
	db := repo.db.WithContext(ctx)

	// If a primary key field is specified, attempt to set the default primary key value to the data.
	if repo.pkField != "" {
		updatedData, err = updateField(data, repo.pkField, repo.pkValue)
		if err != nil {
			return data, handleDBError(gorm.ErrInvalidData)
		}
	} else {
		// If no primary key field is specified, allow global updates to update all records.
		db = db.Session(&gorm.Session{AllowGlobalUpdate: true})
		updatedData = data
	}

	// Check if the record already exists in the database.
	// Disable the logger to remove record not found messages if the record doesn't exist
	repo.db.Logger = logger.Discard
	existingData, err := repo.Get(ctx)
	repo.db.Logger = logger.Default

	// If the record does not exist, create a new record.
	if err != nil {
		if errors.Is(err, repositories.NotFoundError) {
			err := db.Create(&updatedData).Error
			return updatedData, handleDBError(err)
		} else {
			return existingData, handleDBError(err)
		}
	}

	// If the record exists, update the existing record.
	err = db.Model(&existingData).Updates(updatedData).Error
	return existingData, handleDBError(err)
}

// Get retrieves the record from the repository.
func (repo *GenericEntityRepositoryGORM[T]) Get(ctx context.Context) (T, error) {
	var result T
	var err error

	// If a primary key field is specified, retrieve the record by the primary key value.
	if repo.pkField != "" {
		tableName := repo.db.NamingStrategy.TableName(reflect.TypeOf(*new(T)).Name())
		columnName := repo.db.NamingStrategy.ColumnName(tableName, repo.pkField)
		err = repo.db.WithContext(ctx).
			Where(fmt.Sprintf("%s = ?", columnName), repo.pkValue).
			First(&result).
			Error
	} else {
		// If no primary key field is specified, retrieve the first record.
		err = repo.db.WithContext(ctx).First(&result).Error
	}

	return result, handleDBError(err)
}

// Delete removes the record from the repository.
func (repo *GenericEntityRepositoryGORM[T]) Delete(ctx context.Context) error {
	var err error

	// If a primary key field is specified, delete the record by the primary key value.
	if repo.pkField != "" {
		tableName := repo.db.NamingStrategy.TableName(reflect.TypeOf(*new(T)).Name())
		columnName := repo.db.NamingStrategy.ColumnName(tableName, repo.pkField)
		err = repo.db.WithContext(ctx).
			Delete(new(T), fmt.Sprintf("%s = ?", columnName), repo.pkValue).
			Error
	} else {
		// If no primary key field is specified, allow global updates and delete all records.
		err = repo.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(new(T)).Error
	}

	return err
}
