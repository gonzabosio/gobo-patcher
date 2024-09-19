package gobo

import (
	"encoding/json"
	"fmt"
)

func DoPatch(original, new []byte) (diff map[string]interface{}, err error) {
	var originalMap map[string]interface{}
	err = json.Unmarshal(original, &originalMap)
	if err != nil {
		return nil, fmt.Errorf("original json-encoded parse failed %w: ", err)
	}
	var newMap map[string]interface{}
	err = json.Unmarshal(new, &newMap)
	if err != nil {
		return nil, fmt.Errorf("new json-encoded parse failed %w: ", err)
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
			if k == k2 && v != "" && v != v2 {
				diff[k] = v
				break
			}
		}
	}
	if len(diff) == 0 {
		return nil, fmt.Errorf("there are no differences between values")
	}
	return diff, nil
}
