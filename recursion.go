package gobo

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

func appendNewSlice(original, new []interface{}) []interface{} {
	fmt.Println("In appendNewSlice method", original, new)
	original = append(original, new...)
	return original
}

func appendNewSliceDiffs(original, new []interface{}) []interface{} {
	fmt.Println("In appendNewSliceDiffs(default) method", original, new)
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
