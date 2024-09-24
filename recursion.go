package gobo

import (
	"fmt"
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

func equalSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}
