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

func DoPatch(original, new []byte) (diff map[string]interface{}, err error) {
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
	diff, err = iterateMaps(originalMap, newMap)
	if err != nil {
		return nil, fmt.Errorf("failed maps iteration: %w", err)
	}
	return diff, nil
}

func iterateMaps(original, new map[string]interface{}) (map[string]interface{}, error) {
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
				new := reflect.ValueOf(v)
				orig := reflect.ValueOf(v2)
				fmt.Println(new)
				fmt.Println(orig)
			} else if k == k2 && v != "" && v != v2 {
				diff[k] = v
				break
			}
		}
	}
	if len(diff) == 0 {
		return nil, ErrNoDiff
	}
	return diff, nil
}
