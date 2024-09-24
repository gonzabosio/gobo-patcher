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
)

// DoPatch will handle the update of the given structures.
// It checks values between database record and new request.
//
// Pass the database content as original and the request body for the update as new. Ensure given data will be a json in bytes array format
//
// To configure analysis of slices add UseReplaceSlice or UseAddNewSlice function as DoPatch argument
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

func iterateMaps(original, new map[string]interface{}, opts Options) (map[string]interface{}, error) {
	diff := make(map[string]interface{})
	for k, v := range new {
		for k2, v2 := range original {
			// fmt.Println(reflect.TypeOf(v2))
			if k != k2 && v == v2 {
				return nil, ErrKeyConflict
			} else if _, ok := v2.(map[string]interface{}); ok && k == k2 {
				// nested json
				vby, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				vby2, err := json.Marshal(v2)
				if err != nil {
					return nil, err
				}
				var vmap map[string]interface{}
				if err := json.Unmarshal(vby, &vmap); err != nil {
					return nil, err
				}
				var vmap2 map[string]interface{}
				if err := json.Unmarshal(vby2, &vmap2); err != nil {
					return nil, err
				}
				for k, v := range vmap {
					for k2, v2 := range vmap2 {
						if k == k2 && v != v2 {
							diff[k] = v
						}
					}
				}
			} else if _, ok := v2.([]interface{}); ok {
				// slices
				new := reflect.ValueOf(v)
				orig := reflect.ValueOf(v2)
				var newSli []interface{}
				var origSli []interface{}
				for i := range new.Len() {
					if new.Kind() == reflect.ValueOf(v).Kind() {
						newSli = append(newSli, new.Index(i).Interface())
					} else {
						newSli = append(newSli, new.Index(i))
					}
				}
				for i := range orig.Len() {
					if orig.Kind() == reflect.ValueOf(v2).Kind() {
						origSli = append(origSli, orig.Index(i).Interface())
					} else {
						origSli = append(origSli, orig.Index(i))
					}
				}
				if areEqual := equalSlices(newSli, origSli); !areEqual {
					if opts.AddNewSlice {
						diff[k] = appendNewSlice(origSli, newSli)
					} else if opts.ReplaceSlice {
						diff[k] = newSli
					} else {
						diff[k] = appendNewSliceDiffs(origSli, newSli)
					}
				}
			} else if k == k2 && v != "" && v != v2 {
				diff[k] = v
			}
		}
	}
	if len(diff) == 0 {
		return nil, ErrNoDiff
	}
	return diff, nil
}
