package repositories_gorm

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/internal/repositories"
)

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

// updateField is a generic function that updates a field of a struct or a pointer to a struct.
// The function uses reflection to dynamically update the specified field of the input struct.
func updateField[T interface{}](input T, fieldName string, newValue interface{}) (T, error) {
	// Use reflection to get the struct's field
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		// If input is a pointer, get the underlying element
		val = val.Elem()
	} else {
		// If input is not a pointer, ensure it's addressable
		val = reflect.ValueOf(&input).Elem()
	}

	// Check if the input is a struct
	if val.Kind() != reflect.Struct {
		return input, fmt.Errorf("Not a struct: %T", input)
	}

	// Get the field by name
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return input, fmt.Errorf("Field not found: %v", fieldName)
	}

	// Check if the field is settable
	if !field.CanSet() {
		return input, fmt.Errorf("Field not settable: %v", fieldName)
	}

	// Check if types are compatible
	if !reflect.TypeOf(newValue).ConvertibleTo(field.Type()) {
		return input, fmt.Errorf("Incompatible value: %v", newValue)
	}

	// Convert the new value to the field type
	convertedValue := reflect.ValueOf(newValue).Convert(field.Type())

	// Set the new value to the field
	field.Set(convertedValue)

	return input, nil
}

// isEmptyValue checks if a value is considered empty (zero or nil).

// isEmptyValue checks if value represents a zero-value struct (or pointer to a zero-value struct) using reflection.
// The function is useful for determining if a struct or its pointer is empty, i.e., all fields have their zero-values.
func isEmptyValue(value interface{}) bool {
	// Check if the value is nil
	if value == nil {
		return true
	}

	// Use reflection to get the value's type and kind
	val := reflect.ValueOf(value)

	// If the value is a pointer, dereference it to get the underlying element
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if the value is zero (empty) based on its kind
	return val.IsZero()
}
