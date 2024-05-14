package repositories_clover

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/ostafen/clover/v2"
	clover_d "github.com/ostafen/clover/v2/document"
	"gitlab.com/nunet/device-management-service/db/repositories"
)

func handleDBError(err error) error {
	if err != nil {
		switch err {
		case clover.ErrDocumentNotExist:
			// Return NotFoundError for record not found errors
			return repositories.NotFoundError
		case clover.ErrDuplicateKey:
			// Return InvalidDataError for various invalid data errors
			return repositories.InvalidDataError
		default:
			// Return DatabaseError for other unspecified database errors
			return repositories.DatabaseError
		}
	}
	return nil
}

func toCloverDoc[T repositories.ModelType](data T) *clover_d.Document {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return clover_d.NewDocument()
	}
	mappedData := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &mappedData)
	if err != nil {
		return clover_d.NewDocument()
	}

	doc := clover_d.NewDocumentOf(mappedData)
	return doc
}

func toModel[T repositories.ModelType](doc *clover_d.Document) T {
	var model T
	err := doc.Unmarshal(&model)
	model, err = repositories.UpdateField(model, "ID", doc.ObjectId())
	if err != nil {
		return model
	}
	return model
}

func fieldJSONTag[T repositories.ModelType](field string) string {
	fieldName := field
	if field, ok := reflect.TypeOf(*new(T)).FieldByName(field); ok {
		if tag, ok := field.Tag.Lookup("json"); ok {
			fieldName = strings.Split(tag, ",")[0]
		}
	}
	return fieldName
}
