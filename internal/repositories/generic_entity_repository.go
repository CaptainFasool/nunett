package repositories

import (
	"context"
)

// GenericEntityRepository is an interface defining basic CRUD operations for repositories handling a single record.
type GenericEntityRepository[T interface{}] interface {
	// Save adds or updates a single record in the repository.
	Save(ctx context.Context, data T) (T, error)
	// Get retrieves the single record from the repository.
	Get(ctx context.Context) (T, error)
	// Delete removes the single record from the repository.
	Delete(ctx context.Context) error
}

