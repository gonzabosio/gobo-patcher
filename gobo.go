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
		// fmt.Println("Type", reflect.TypeOf(v))
		// fmt.Println(k, v)
		for k2, v2 := range original {
			if k != k2 && v == v2 {
				return nil, ErrKeyConflict
			} else if k == k2 {
				// fmt.Println(k2, v2)
				switch reflect.TypeOf(v).Kind() {
				case reflect.Float64:
					diff[k] = v
				case reflect.Slice:
					{
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
						if orig, new, areEqual := equalSlices(origSli, newSli); !areEqual {
							if orig != nil {
								diffOfMap, err := iterateMaps(orig, new, opts)
								if err != nil {
									fmt.Println(err)
								}
								diff[k] = diffOfMap
							} else if opts.AddNewSlice {
								diff[k] = appendNewSlice(origSli, newSli)
							} else if opts.ReplaceSlice {
								diff[k] = newSli
							} else {
								diff[k] = appendNewSliceDiffs(origSli, newSli)
							}
							break
						}
					}
				default:
					if _, ok := v2.(map[string]interface{}); ok {
						// nested json
						originalMap, newMap := convertToMap(v2, v)
						for k, v := range newMap {
							for k2, v2 := range originalMap {
								if k == k2 {
									if _, ok := v.([]interface{}); ok {
										diff = handleSlice(v, v2, diff, k, opts)
										break
									} else if v != v2 {
										diff[k] = v
									}
								}
							}
						}
						break
					} else if v != "" && v != v2 {
						diff[k] = v
						break
					}
				}
			}
		}
	}
	if len(diff) == 0 {
		return nil, ErrNoDiff
	}
	return diff, nil
}
