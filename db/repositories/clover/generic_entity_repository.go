package repositories_clover

import (
	"context"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	clover "github.com/ostafen/clover/v2"
	clover_d "github.com/ostafen/clover/v2/document"
	clover_q "github.com/ostafen/clover/v2/query"

	"gitlab.com/nunet/device-management-service/db/repositories"
)

const (
	pKField = "_id"
)

// GenericEntityRepositoryClover is a generic single entity repository implementation using Clover.
// It is intended to be embedded in single entity model repositories to provide basic database operations.
type GenericEntityRepositoryClover[T repositories.ModelType] struct {
	db         *clover.DB // db is the Clover database instance.
	collection string     // collection is the name of the collection in the database.
}

// NewGenericEntityRepository creates a new instance of GenericEntityRepositoryClover.
// It initializes and returns a repository with the provided Clover database, primary key field, and value.
func NewGenericEntityRepository[T repositories.ModelType](
	db *clover.DB,
) repositories.GenericEntityRepository[T] {
	collection := strcase.ToSnake(reflect.TypeOf(*new(T)).Name())
	return &GenericEntityRepositoryClover[T]{db: db, collection: collection}
}

// GetQuery returns a clean Query instance for building queries.
func (repo *GenericEntityRepositoryClover[T]) GetQuery() repositories.Query[T] {
	return repositories.Query[T]{}
}

func (repo *GenericEntityRepositoryClover[T]) query() *clover_q.Query {
	return clover_q.NewQuery(repo.collection)
}

// Save creates or updates the record to the repository and returns the new/updated data.
func (repo *GenericEntityRepositoryClover[T]) Save(ctx context.Context, data T) (T, error) {
	doc := toCloverDoc(data)
	doc.Set("CreatedAt", time.Now())

	_, err := repo.db.InsertOne(repo.collection, doc)
	if err != nil {
		return data, handleDBError(err)
	}

	return toModel[T](doc), handleDBError(err)
}

// Get retrieves the record from the repository.
func (repo *GenericEntityRepositoryClover[T]) Get(ctx context.Context) (T, error) {
	var result T
	q := repo.query().Sort(clover_q.SortOption{
		Field:     "CreatedAt",
		Direction: -1,
	})
	doc, err := repo.db.FindFirst(q)

	if err != nil || doc == nil {
		return result, handleDBError(err)
	}
	return toModel[T](doc), handleDBError(err)
}

// Clear removes the record with its history from the repository.
func (repo *GenericEntityRepositoryClover[T]) Clear(ctx context.Context) error {
	return repo.db.Delete(repo.query())
}

// History retrieves previous versions of the record from the repository.
func (repo *GenericEntityRepositoryClover[T]) History(
	ctx context.Context,
	query repositories.Query[T],
) ([]T, error) {
	var results []T
	q := repo.query()
	q = applyConditions(q, query)

	err := repo.db.ForEach(q, func(doc *clover_d.Document) bool {
		var result T
		err := doc.Unmarshal(&result)
		if err != nil {
			return false
		}
		results = append(results, result)
		return true
	})

	return results, handleDBError(err)
}
