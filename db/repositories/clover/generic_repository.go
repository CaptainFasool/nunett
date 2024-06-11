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
	pkField        = "_id"
	deletedAtField = "DeletedAt"
)

// GenericRepositoryClover is a generic repository implementation using Clover.
// It is intended to be embedded in model repositories to provide basic database operations.
type GenericRepositoryClover[T repositories.ModelType] struct {
	db         *clover.DB // db is the Clover database instance.
	collection string     // collection is the name of the collection in the database.
}

// NewGenericRepository creates a new instance of GenericRepositoryClover.
// It initializes and returns a repository with the provided Clover database.
func NewGenericRepository[T repositories.ModelType](
	db *clover.DB,
) repositories.GenericRepository[T] {
	collection := strcase.ToSnake(reflect.TypeOf(*new(T)).Name())
	return &GenericRepositoryClover[T]{db: db, collection: collection}
}

// GetQuery returns a clean Query instance for building queries.
func (repo *GenericRepositoryClover[T]) GetQuery() repositories.Query[T] {
	return repositories.Query[T]{}
}

func (repo *GenericRepositoryClover[T]) query(includeDeleted bool) *clover_q.Query {
	query := clover_q.NewQuery(repo.collection)
	if !includeDeleted {
		query = query.Where(clover_q.Field(deletedAtField).LtEq(time.Unix(0, 0)))
	}
	return query
}

func (repo *GenericRepositoryClover[T]) queryWithID(
	id interface{},
	includeDeleted bool,
) *clover_q.Query {
	return repo.query(includeDeleted).Where(clover_q.Field(pkField).Eq(id.(string)))
}

// Create adds a new record to the repository and returns the created data.
func (repo *GenericRepositoryClover[T]) Create(ctx context.Context, data T) (T, error) {
	doc := toCloverDoc(data)
	doc.Set("CreatedAt", time.Now())

	_, err := repo.db.InsertOne(repo.collection, doc)
	if err != nil {
		return data, handleDBError(err)
	}

	return toModel[T](doc), handleDBError(err)
}

// Get retrieves a record by its identifier.
func (repo *GenericRepositoryClover[T]) Get(ctx context.Context, id interface{}) (T, error) {
	var result T
	doc, err := repo.db.FindFirst(repo.queryWithID(id, false))
	if err != nil || doc == nil {
		return result, handleDBError(err)
	}
	return toModel[T](doc), handleDBError(err)
}

// Update modifies a record by its identifier.
func (repo *GenericRepositoryClover[T]) Update(
	ctx context.Context,
	id interface{},
	data T,
) (T, error) {
	updates := toCloverDoc(data).AsMap()
	updates["UpdatedAt"] = time.Now()

	err := repo.db.Update(repo.queryWithID(id, false), updates)
	if err != nil {
		return data, handleDBError(err)
	}

	data, err = repo.Get(ctx, id)
	return data, handleDBError(err)
}

// Delete removes a record by its identifier.
func (repo *GenericRepositoryClover[T]) Delete(ctx context.Context, id interface{}) error {
	err := repo.db.Update(
		repo.queryWithID(id, false),
		map[string]interface{}{"DeletedAt": time.Now()},
	)
	return err
}

// Find retrieves a single record based on a query.
func (repo *GenericRepositoryClover[T]) Find(
	ctx context.Context,
	query repositories.Query[T],
) (T, error) {
	var result T
	q := repo.query(false)
	q = applyConditions(q, query)
	doc, err := repo.db.FindFirst(q)
	if err != nil || doc == nil {
		return result, handleDBError(err)
	}

	return toModel[T](doc), handleDBError(err)
}

// FindAll retrieves multiple records based on a query.
func (repo *GenericRepositoryClover[T]) FindAll(
	ctx context.Context,
	query repositories.Query[T],
) ([]T, error) {
	var results []T
	q := repo.query(false)
	q = applyConditions(q, query)

	err := repo.db.ForEach(q, func(doc *clover_d.Document) bool {
		results = append(results, toModel[T](doc))
		return true
	})

	return results, handleDBError(err)
}

// applyConditions applies conditions, sorting, limiting, and offsetting to a Clover database query.
// It takes a Clover database instance (db) and a generic query (repositories.Query) as input.
// The function dynamically constructs the WHERE clause based on the provided conditions and instance values.
// It also includes sorting, limiting, and offsetting based on the query parameters.
// The modified Clover database instance is returned.
func applyConditions[T repositories.ModelType](
	q *clover_q.Query,
	query repositories.Query[T],
) *clover_q.Query {
	// Apply conditions specified in the query.
	for _, condition := range query.Conditions {
		// change the field name to json tag name if specified in the struct
		condition.Field = fieldJSONTag[T](condition.Field)
		switch condition.Operator {
		case "=":
			q = q.Where(clover_q.Field(condition.Field).Eq(condition.Value))
		case ">":
			q = q.Where(clover_q.Field(condition.Field).Gt(condition.Value))
		case ">=":
			q = q.Where(clover_q.Field(condition.Field).GtEq(condition.Value))
		case "<":
			q = q.Where(clover_q.Field(condition.Field).Lt(condition.Value))
		case "<=":
			q = q.Where(clover_q.Field(condition.Field).LtEq(condition.Value))
		case "!=":
			q = q.Where(clover_q.Field(condition.Field).Neq(condition.Value))
		case "IN":
			if values, ok := condition.Value.([]interface{}); ok {
				q = q.Where(clover_q.Field(condition.Field).In(values...))
			}
		case "LIKE":
			if value, ok := condition.Value.(string); ok {
				q = q.Where(clover_q.Field(condition.Field).Like(value))
			}
		}
	}

	// Apply conditions based on non-zero values in the query instance.
	if !repositories.IsEmptyValue(query.Instance) {
		exampleType := reflect.TypeOf(query.Instance)
		exampleValue := reflect.ValueOf(query.Instance)
		for i := 0; i < exampleType.NumField(); i++ {
			fieldName := exampleType.Field(i).Name
			fieldName = fieldJSONTag[T](fieldName)
			fieldValue := exampleValue.Field(i).Interface()
			if !repositories.IsEmptyValue(fieldValue) {
				q = q.Where(clover_q.Field(fieldName).Eq(fieldValue))
			}
		}
	}

	// Apply sorting if specified in the query.
	if query.SortBy != "" {
		dir := 1
		if query.SortBy[0] == '-' {
			dir = -1
			query.SortBy = fieldJSONTag[T](query.SortBy[1:])
		}
		q = q.Sort(clover_q.SortOption{Field: query.SortBy, Direction: dir})
	}

	// Apply limit if specified in the query.
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}

	// Apply offset if specified in the query.
	if query.Offset > 0 {
		q = q.Limit(query.Offset)
	}

	return q
}
