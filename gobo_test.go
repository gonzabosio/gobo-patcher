package gobo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoPatch(t *testing.T) {
	t.Run("detect differences", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe"}`
		newData := `{"name":"Jane", "last_name":"Doe"}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Equal(t, "Jane", diff["name"])
	})
	t.Run("detect differences with missing fields", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "age":32}`
		newData := `{"name":"John","age": 36}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Equal(t, float64(36), diff["age"])
	})
	t.Run("unmarshal parse fail", func(t *testing.T) {
		dbRec := `"name":"John", "last_name":"Doe", "age":32}`
		newData := `{"name":"John","age": 36}`
		diff, _ := DoPatch([]byte(dbRec), []byte(newData))
		assert.NotEqual(t, float64(36), diff["age"])
	})
	t.Run("key conflicts", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe"}`
		newData := `{"name":"Jane", "lastname":"Doe"}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		assert.Equal(t, map[string]interface{}(nil), diff)
		assert.Equal(t, err, fmt.Errorf("failed maps iteration: %w", ErrKeyConflict))
	})
	t.Run("detect differences in complex json", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":{"country":"Argentina", "age":45}}`
		newData := `{"name":"Jane", "meta":{"country":"Argentina", "age":40}}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(diff)
		assert.Equal(t, "Jane", diff["name"])
		assert.Equal(t, float64(40), diff["age"])
	})
	t.Run("detect differences in slice", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brasil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brasil", "United States"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
	})
	t.Run("using 'appendNewSlice' option", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brasil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brasil"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseAddNewSlice())
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
	})
	t.Run("using 'appendNewSliceDiffs' option", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brasil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brasil"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseReplaceSlice())
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
	})
	t.Run("using default option", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brasil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brasil", "United States"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
	})
}
