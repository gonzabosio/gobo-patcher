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
	t.Run("detect differences in slice and add it(default behavior)", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brazil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brazil", "United States"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
		expected := []interface{}{"Argentina", "Brazil", "Canada", "United States"}
		assert.Equal(t, expected, diff["countries"])
		assert.Equal(t, "Jane", diff["name"])
	})
	t.Run("using 'appendNewSlice' option", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brazil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brazil"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseAddNewSlice())
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
		expected := []interface{}{"Argentina", "Brazil", "Canada", "Argentina", "Brazil"}
		assert.Equal(t, expected, diff["countries"])
		assert.Equal(t, "Jane", diff["name"])
	})
	t.Run("using 'replaceSlice' option", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brazil", "Canada"]}`
		newData := `{"name":"Jane", "countries":["Argentina", "Brazil", "United States"]}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseReplaceSlice())
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
		expected := []interface{}{"Argentina", "Brazil", "United States"}
		assert.Equal(t, expected, diff["countries"])
		assert.Equal(t, "Jane", diff["name"])
	})
}

func TestDoPatchExtended(t *testing.T) {
	t.Run("nested json slices numbers", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":
			[
				{
					"user1":{"posts":["Post 1", "Post 2", "Post 3"], "age":45}
				}
			]
		}`
		newData := `{"name":"Jane", "meta":
			[
				{
					"user1":{"posts":["Post 4", "Post 5", "Post 6"], "age":41}
				}
			]
		}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseReplaceSlice())
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string]interface{}{
			"age":   41.0,
			"posts": []interface{}{"Post 4", "Post 5", "Post 6"},
		}
		assert.Equal(t, expected, diff["meta"])
		assert.Equal(t, "Jane", diff["name"])
	})
	t.Run("using default behavior of slices and detect map inside", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":
			[
				{
					"user1":{"posts":["Post 1", "Post 2", "Post 3"], "age":45}
				}
			]
		}`
		newData := `{"name":"Jane", "meta":
			[
				{
					"user1":{"posts":["Post 4", "Post 5", "Post 6"], "age":41}
				}
			]
		}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string]interface{}{
			"meta": map[string]interface{}{
				"age": 41.0,
				"posts": []interface{}{
					"Post 1", "Post 2", "Post 3", "Post 4", "Post 5", "Post 6",
				},
			},
			"name": "Jane",
		}
		assert.Equal(t, expected, diff)
	})
	t.Run("using AddNewSlice behavior and detect map inside", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":
			[
				{
					"user1":{"posts":["Post 1", "Post 2", "Post 3"], "age":45}
				}
			]
		}`
		newData := `{"name":"Jane", "meta":
			[
				{
					"user1":{"posts":["Post 4", "Post 5", "Post 3"], "age":41}
				}
			]
		}`
		diff, err := DoPatch([]byte(dbRec), []byte(newData), UseAddNewSlice())
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string]interface{}{
			"meta": map[string]interface{}{
				"age": 41.0,
				"posts": []interface{}{
					"Post 1", "Post 2", "Post 3", "Post 4", "Post 5", "Post 3",
				},
			},
			"name": "Jane",
		}
		assert.Equal(t, expected, diff)
	})
}

func TestDoPatchWithQuery(t *testing.T) {
	t.Run("full condition argument", func(t *testing.T) {
		db := `{"name": "Gonzalo", "age": 19, "phoneNumber": "1 1234 5678"}`
		new := `{"name": "Gonza", "age": 20}`
		condition := `WHERE phoneNumber = '1 1234 5678'`
		query, err := DoPatchWithQuery([]byte(db), []byte(new), "user", condition, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(query)
		expected := fmt.Sprintf(`UPDATE "user" SET ("name"='Gonza', "age"=20) %v`, condition)
		assert.Equal(t, expected, query)
	})
	t.Run("where id(number)", func(t *testing.T) {
		db := `{"id":1234, "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := DoPatchWithQuery([]byte(db), []byte(new), "user", "id", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(query)
		expected := `UPDATE "user" SET ("name"='Gonza', "age"=20) WHERE "id"=1234`
		assert.Equal(t, expected, query)
	})
	t.Run("where id(string)", func(t *testing.T) {
		db := `{"id":"1234", "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := DoPatchWithQuery([]byte(db), []byte(new), "user", "id", nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(query)
		expected := `UPDATE "user" SET ("name"='Gonza', "age"=20) WHERE "id"='1234'`
		assert.Equal(t, expected, query)
	})
	t.Run("error no condition", func(t *testing.T) {
		db := `{"id":1234, "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := DoPatchWithQuery([]byte(db), []byte(new), "user", "", nil)
		expected := `UPDATE "user" SET ("name"='Gonza', "age"=20) WHERE "id"=1234`
		assert.NotEqual(t, expected, query)
		assert.Equal(t, fmt.Errorf("method did not receive query conditions"), err)
	})
	t.Run("apply attributes relationship", func(t *testing.T) {
		db := `{"id":"1234", "name": "Gonzalo", "age": 19, "country": "Argentina"}`
		new := `{"name": "Gonza", "age": 20, "country": "Greenland"}`
		rel := map[string]string{
			"name": "Name",
			"age":  "Age",
		}
		query, err := DoPatchWithQuery([]byte(db), []byte(new), "user", "id", rel)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(query)
		expected := `UPDATE "user" SET ("Name"='Gonza', "Age"=20, "country"='Greenland') WHERE "id"='1234'`
		assert.Equal(t, expected, query)
	})
}
