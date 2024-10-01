package gobo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNoDiff      = errors.New("there are no differences between values")
	ErrKeyConflict = errors.New("keys with equal values have different names")
	ErrNoCondition = errors.New("method did not receive query conditions")
)

// JSONDiff will handle the differences of the given structures.
// It checks values between original data and the new one and return the differences.
// Ensure given data is a json in bytes array format.
//
// To configure analysis of slices add UseReplaceSlice or UseAddNewSlice function as 'optFuncs' argument.
// If nothing is added, it will conserve original slice and add the differences of the new one. Slices with empty items won't throw an ErrEmptyFields like the others structures.
func JSONDiff(original, new []byte, optFuncs ...Option) (diff map[string]interface{}, err error) {
	opts := Options{}
	for _, optFunc := range optFuncs {
		optFunc(&opts)
	}

	var originalMap map[string]interface{}
	err = json.Unmarshal(original, &originalMap)
	if err != nil {
		return nil, fmt.Errorf("original json-encoded parse failed: %w", err)
	}
	var newMap map[string]interface{}
	err = json.Unmarshal(new, &newMap)
	if err != nil {
		return nil, fmt.Errorf("new json-encoded parse failed: %w", err)
	}

	diff, err = iterateMaps(originalMap, newMap, opts)
	if err != nil {
		return nil, err
	}
	return diff, nil
}

// PatchWithQuery will do the same tasks as DoPatch but instead of return the differences, it will return a PostgreSQL update query with only the necessary changes to be made.
//
// The 'condition' parameter can be completed as you want. It's added after the SET part.
// If the argument is "id", "Id" or "ID", method will consider this attribute as condition to the update. If string is empty, ErrNoCondition will be triggered.
// The 'rel' parameter works as the relationship of given json with database. Keys for json field names and the values as the associated table attributes.
// For example: map[string]string{} {"last_name"(json): "lastName"(database)}
// In the case there are no differences between database and json fields, set 'rel' as nil.
func PatchWithQuery(original, new []byte, table, condition string, rel map[string]string) (query string, err error) {
	switch condition {
	case "":
		return "", ErrNoCondition
	case "id", "Id", "ID":
		diff, idVal, err := findDiffsForQuery(original, new, condition)
		if err != nil {
			return "", err
		}
		switch reflect.TypeOf(idVal).Kind() {
		case reflect.String:
			set := buildSetClause(diff, rel)
			query = fmt.Sprintf(`UPDATE "%s" SET %s WHERE "%s"='%s'`, table, set, condition, idVal.(string))
		case reflect.Float64:
			set := buildSetClause(diff, rel)
			query = fmt.Sprintf(`UPDATE "%s" SET %s WHERE "%s"=%v`, table, set, condition, idVal)
		}
	default:
		diff, _, err := findDiffsForQuery(original, new, condition)
		if err != nil {
			return "", err
		}
		set := buildSetClause(diff, rel)
		query = fmt.Sprintf(`UPDATE "%s" SET %s %v`, table, set, condition)
	}
	return query, nil
}
