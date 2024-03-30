package repositories_gorm

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/internal/repositories"
)

// GenericRepositoryGORM is a generic repository implementation using GORM as an ORM.
// It is intended to be embedded in model repositories to provide basic database operations.
type GenericRepositoryGORM[T interface{}] struct {
	db *gorm.DB
}

// NewGenericRepository creates a new instance of GenericRepositoryGORM.
// It initializes and returns a repository with the provided GORM database.
func NewGenericRepository[T interface{}](db *gorm.DB) repositories.GenericRepository[T] {
	return &GenericRepositoryGORM[T]{db: db}
}

// GetQuery returns a clean Query instance for building queries.
func (repo *GenericRepositoryGORM[T]) GetQuery() repositories.Query[T] {
	return repositories.Query[T]{}
}

// Create adds a new record to the repository and returns the created data.
func (repo *GenericRepositoryGORM[T]) Create(ctx context.Context, data T) (T, error) {
	err := repo.db.WithContext(ctx).Create(&data).Error
	return data, handleDBError(err)
}

// Get retrieves a record by its identifier.
func (repo *GenericRepositoryGORM[T]) Get(ctx context.Context, id uint) (T, error) {
	var result T
	err := repo.db.WithContext(ctx).First(&result, id).Error
	if err != nil {
		return result, handleDBError(err)
	}
	return result, handleDBError(err)
}

// Update modifies a record by its identifier.
func (repo *GenericRepositoryGORM[T]) Update(ctx context.Context, id uint, data T) (T, error) {
	err := repo.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(data).Error
	return data, handleDBError(err)
}

// Delete removes a record by its identifier.
func (repo *GenericRepositoryGORM[T]) Delete(ctx context.Context, id uint) error {
	err := repo.db.WithContext(ctx).Delete(new(T), id).Error
	return err
}

// Find retrieves a single record based on a query.
func (repo *GenericRepositoryGORM[T]) Find(
	ctx context.Context,
	query repositories.Query[T],
) (T, error) {
	var result T
	db := repo.db.WithContext(ctx).Model(new(T))

	db = applyConditions(db, query)

	err := db.First(&result).Error
	return result, handleDBError(err)
}

// FindAll retrieves multiple records based on a query.
func (repo *GenericRepositoryGORM[T]) FindAll(
	ctx context.Context,
	query repositories.Query[T],
) ([]T, error) {
	var results []T
	db := repo.db.WithContext(ctx).Model(new(T))

	db = applyConditions(db, query)

	err := db.Find(&results).Error
	return results, handleDBError(err)
}

// applyConditions applies conditions, sorting, limiting, and offsetting to a GORM database query.
// It takes a GORM database instance (db) and a generic query (repositories.Query) as input.
// The function dynamically constructs the WHERE clause based on the provided conditions and instance values.
// It also includes sorting, limiting, and offsetting based on the query parameters.
// The modified GORM database instance is returned.
func applyConditions[T any](db *gorm.DB, query repositories.Query[T]) *gorm.DB {
	// Retrieve the table name using the GORM naming strategy.
	tableName := db.NamingStrategy.TableName(reflect.TypeOf(*new(T)).Name())

	// Apply conditions specified in the query.
	for _, condition := range query.Conditions {
		columnName := db.NamingStrategy.ColumnName(tableName, condition.Field)
		db = db.Where(
			fmt.Sprintf("%s %s ?", columnName, condition.Operator),
			condition.Value,
		)
	}

	// Apply conditions based on non-zero values in the query instance.
	if !isEmptyValue(query.Instance) {
		exampleType := reflect.TypeOf(query.Instance)
		exampleValue := reflect.ValueOf(query.Instance)
		for i := 0; i < exampleType.NumField(); i++ {
			fieldName := exampleType.Field(i).Name
			fieldValue := exampleValue.Field(i).Interface()
			if !isEmptyValue(fieldValue) {
				columnName := db.NamingStrategy.ColumnName(tableName, fieldName)
				db = db.Where(fmt.Sprintf("%s = ?", columnName), fieldValue)
			}
		}
	}

	// Apply sorting if specified in the query.
	if query.SortBy != "" {
		db = db.Order(query.SortBy)
	}

	// Apply limit if specified in the query.
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}

	// Apply offset if specified in the query.
	if query.Offset > 0 {
		db = db.Limit(query.Offset)
	}

	return db
}
