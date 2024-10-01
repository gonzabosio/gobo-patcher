package gobo

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
)

// Detect all kind of changes such as slices and nested json.
// The goal is for it to be general purpose differences detector while simpleMapIterator is used to build sql queries from a flat structure.
func iterateMaps(original, new map[string]interface{}, opts Options) (map[string]interface{}, error) {
	diff := make(map[string]interface{})
	for k, v := range new {
		for k2, v2 := range original {
			if k != k2 && v == v2 {
				return nil, ErrKeyConflict
			} else if k == k2 {
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
									return nil, err
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
					} else if v != v2 {
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

// Detect changes in flat json structures such as strings and numbers. Used in DoPatchWithQuery method to create the queries.
func simpleMapIterator(original, new map[string]interface{}) (map[string]interface{}, error) {
	diff := make(map[string]interface{})
	for k, v := range new {
		for k2, v2 := range original {
			if k != k2 && v == v2 {
				return nil, ErrKeyConflict
			} else if k == k2 {
				switch reflect.TypeOf(v).Kind() {
				case reflect.Float64:
					diff[k] = v
				case reflect.String:
					if v != v2 {
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

func appendNewSlice(original, new []interface{}) []interface{} {
	original = append(original, new...)
	return original
}

func appendNewSliceDiffs(original, new []interface{}) []interface{} {
	var diff []interface{}
	var found bool
	for i := range new {
		found = true
		for j := range original {
			if new[i] == original[j] {
				found = false
				break
			}
		}
		if found {
			diff = append(diff, new[i])
		}
	}
	original = append(original, diff...)
	return original
}

func equalSlices(originalSlice, newSlice []interface{}) (map[string]interface{}, map[string]interface{}, bool) {
	if len(originalSlice) != len(newSlice) {
		return nil, nil, false
	}
	for i := range originalSlice {
		if !reflect.DeepEqual(originalSlice[i], newSlice[i]) {
			if reflect.TypeOf(originalSlice[i]).Kind() == reflect.Map && reflect.TypeOf(newSlice[i]).Kind() == reflect.Map {
				return originalSlice[i].(map[string]interface{}), newSlice[i].(map[string]interface{}), false
			} else {
				return nil, nil, false
			}
		}
	}
	return nil, nil, true
}

func convertToMap[T reflect.Value | interface{}](original, new T) (originalMap, newMap map[string]interface{}) {
	originalMap = make(map[string]interface{})
	newMap = make(map[string]interface{})
	newBytes, err := json.Marshal(new)
	if err != nil {
		log.Fatalf("Failed new map marshal: %v", err)
	}
	originalBytes, err := json.Marshal(original)
	if err != nil {
		log.Fatalf("Failed original map marshal: %v", err)
	}
	err = json.Unmarshal(newBytes, &newMap)
	if err != nil {
		log.Fatalf("Failed new unmarshal: %v", err)
	}
	err = json.Unmarshal(originalBytes, &originalMap)
	if err != nil {
		log.Fatalf("Failed new unmarshal: %v", err)
	}
	return
}

func handleSlice(v, v2 interface{}, diff map[string]interface{}, key string, opts Options) map[string]interface{} {
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
	if opts.AddNewSlice {
		diff[key] = appendNewSlice(origSli, newSli)
	} else if opts.ReplaceSlice {
		diff[key] = newSli
	} else {
		diff[key] = appendNewSliceDiffs(origSli, newSli)
	}
	return diff
}

func findDiffsForQuery(original, new []byte, idKey string) (diff map[string]interface{}, idVal interface{}, err error) {
	var originalMap map[string]interface{}
	err = json.Unmarshal(original, &originalMap)
	if err != nil {
		return nil, idVal, fmt.Errorf("original json-encoded parse failed: %w", err)
	}
	var newMap map[string]interface{}
	err = json.Unmarshal(new, &newMap)
	if err != nil {
		return nil, idVal, fmt.Errorf("new json-encoded parse failed: %w", err)
	}
	idVal = originalMap[idKey]
	diff, err = simpleMapIterator(originalMap, newMap)
	if err != nil {
		return nil, idVal, err
	}
	return diff, idVal, nil
}

func buildSetClause(diff map[string]interface{}, rel map[string]string) (set string) {
	keys := make([]string, 0, len(diff))
	for key := range diff {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sets []string
	var attr string
	if rel != nil {
		for _, k := range keys {
			attr = k
			if dbAttr, found := rel[k]; found {
				attr = dbAttr
			}
			if value, ok := diff[k].(string); ok {
				sets = append(sets, fmt.Sprintf(`"%s"='%s'`, attr, value))
			} else {
				sets = append(sets, fmt.Sprintf(`"%s"=%v`, attr, diff[k]))
			}
		}
		set = strings.Join(sets, ", ")
	} else {
		for _, k := range keys {
			if value, ok := diff[k].(string); ok {
				sets = append(sets, fmt.Sprintf(`"%s"='%s'`, k, value))
			} else {
				sets = append(sets, fmt.Sprintf(`"%s"=%v`, k, diff[k]))
			}
		}
		set = strings.Join(sets, ", ")
	}
	return set
}
