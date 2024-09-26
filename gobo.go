package gobo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoDiff      = errors.New("there are no differences between values")
	ErrKeyConflict = errors.New("keys with equal values have different names")
)

// DoPatch will handle the update of the given structures.
// It checks values between database record and new request.
//
// Pass the database content as original and the request body for the update as new. Ensure given data will be a json in bytes array format
//
// To configure analysis of slices add UseReplaceSlice or UseAddNewSlice function as 'optFuncs' argument.
// If nothing is added, it will conserve original data and add the differences with the new one when slices appears.
func DoPatch(original, new []byte, optFuncs ...Option) (diff map[string]interface{}, err error) {
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
		return nil, fmt.Errorf("failed maps iteration: %w", err)
	}
	return diff, nil
}

// DoPatchWithQuery will make the same tasks as DoPatch but instead of return the differences, it will return a PostgreSQL query with only the necessary changes.
//
// The parameter 'rel' works as the relationship of given json with database. Keys as table attributes and the values as the association with json keys.
// It isn't necessary to specify if there are no differences between database and json fields. In that case set 'rel' as nil.
// For example: map[string]string{} {"lastName"(database): "last_name"(json)}
func DoPatchWithQuery(original, new []byte, table string, rel map[string]string) (query string, err error) {
	diff, err := DoPatch(original, new)
	if err != nil {
		return "", err
	}
	var sets []string
	for k, v := range diff {
		if _, ok := v.(string); ok {
			sets = append(sets, fmt.Sprintf(`"%s"='%s'`, k, v))
		} else {
			sets = append(sets, fmt.Sprintf(`"%s"=%v`, k, v))
		}
	}
	set := strings.Join(sets, ", ")
	query = fmt.Sprintf(`UPDATE "%s" SET (%s) WHERE id=%v`, table, set, 1)
	return query, nil
}
