package gobo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDiff(t *testing.T) {
	t.Run("detect differences", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe"}`
		newData := `{"name":"Jane", "last_name":"Doe"}`
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Equal(t, "Jane", diff["name"])
	})
	t.Run("detect differences with missing fields", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "age":32}`
		newData := `{"name":"John","age": 36}`
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Equal(t, float64(36), diff["age"])
	})
	t.Run("unmarshal parse fail", func(t *testing.T) {
		dbRec := `"name":"John", "last_name":"Doe", "age":32}`
		newData := `{"name":"John","age": 36}`
		diff, _ := JSONDiff([]byte(dbRec), []byte(newData))
		assert.NotEqual(t, float64(36), diff["age"])
	})
	t.Run("key conflicts", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe"}`
		newData := `{"name":"Jane", "lastname":"Doe"}`
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
		assert.Equal(t, map[string]interface{}(nil), diff)
		assert.Equal(t, err, ErrKeyConflict)
	})
	t.Run("detect differences in complex json", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":{"country":"Argentina", "age":45}}`
		newData := `{"name":"Jane", "meta":{"country":"Argentina", "age":40}}`
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData), UseAddNewSlice())
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData), UseReplaceSlice())
		if err != nil {
			t.Fatalf("Error iterating on json with slice: %v", err)
		}
		t.Log("Differences ", diff)
		expected := []interface{}{"Argentina", "Brazil", "United States"}
		assert.Equal(t, expected, diff["countries"])
		assert.Equal(t, "Jane", diff["name"])
	})

	t.Run("empty field", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "countries":["Argentina", "Brazil", "Canada"]}`
		newData := `{"name":"", "countries":["Argentina", "Brazil", "United States"]}`
		query, err := JSONDiff([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, map[string]interface{}{"name": "", "countries": []interface{}{"Argentina", "Brazil", "Canada", "United States"}}, query)
	})
}

func TestJSONDiffExtended(t *testing.T) {
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData), UseReplaceSlice())
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
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
		diff, err := JSONDiff([]byte(dbRec), []byte(newData), UseAddNewSlice())
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
	t.Run("empty fields error in nested json", func(t *testing.T) {
		dbRec := `{"name":"John", "last_name":"Doe", "meta":{"country":"Argentina", "age":45}}`
		newData := `{"name":"Jane", "meta":{"country":"", "age":40}}`
		diff, err := JSONDiff([]byte(dbRec), []byte(newData))
		if err != nil {
			t.Fatal(err)
		}
		expected := map[string]interface{}{
			"age":     40.0,
			"country": "",
			"name":    "Jane",
		}
		assert.Equal(t, expected, diff)
	})
}

func TestPatchWithQuery(t *testing.T) {
	t.Run("full condition argument", func(t *testing.T) {
		db := `{"name": "Gonzalo", "age": 19, "phoneNumber": "1 1234 5678"}`
		new := `{"name": "Gonza", "age": 20}`
		condition := `WHERE phoneNumber = '1 1234 5678'`
		query, err := PatchWithQuery([]byte(db), []byte(new), "user", condition, true, nil)
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf(`UPDATE "user" SET age=20, name='Gonza' %v`, condition)
		assert.Equal(t, expected, query)
	})
	t.Run("where id(number)", func(t *testing.T) {
		db := `{"id":1234, "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := PatchWithQuery([]byte(db), []byte(new), "user", "id", true, nil)
		if err != nil {
			t.Fatal(err)
		}
		expected := `UPDATE "user" SET age=20, name='Gonza' WHERE id=1234`
		kwQuery := strings.Replace(query, `user`, `"user"`, -1)
		assert.Equal(t, expected, kwQuery)
	})
	t.Run("where id(string)", func(t *testing.T) {
		db := `{"id":"1234", "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := PatchWithQuery([]byte(db), []byte(new), "user", "id", true, nil)
		if err != nil {
			t.Fatal(err)
		}
		expected := `UPDATE "user" SET age=20, name='Gonza' WHERE id='1234'`
		kwQuery := strings.Replace(query, `user`, `"user"`, -1)
		assert.Equal(t, expected, kwQuery)
	})
	t.Run("error no condition", func(t *testing.T) {
		db := `{"id":1234, "name": "Gonzalo", "age": 19}`
		new := `{"name": "Gonza", "age": 20}`
		query, err := PatchWithQuery([]byte(db), []byte(new), "user", "", false, nil)
		expected := `UPDATE "user" SET age=20, name='Gonza' WHERE id=1234`
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
		query, err := PatchWithQuery([]byte(db), []byte(new), "user", "id", true, rel)
		if err != nil {
			t.Fatal(err)
		}
		expected := `UPDATE "user" SET Age=20, country='Greenland', Name='Gonza' WHERE id='1234'`
		kwQuery := strings.Replace(query, `user`, `"user"`, -1)
		assert.Equal(t, expected, kwQuery)
	})

	t.Run("empty field", func(t *testing.T) {
		dbRec := `{"id":"1234", "name":"John", "age": 30, "country": "Argentina"}`
		newData := `{"name": "Jane", "age": 28, "country": ""}`
		query, err := PatchWithQuery([]byte(dbRec), []byte(newData), "user", "id", false, nil)
		if err != nil {
			t.Fatal(err)
		}
		kwQuery := strings.Replace(query, `user`, `"user"`, -1)
		assert.Equal(t, `UPDATE "user" SET age=28, country='', name='Jane' WHERE id='1234'`, kwQuery)
	})

	t.Run("accurate filter", func(t *testing.T) {
		dbRec := `{"id":1014336373145370625,"name":"res-man","details":"details of project","team_id":1014110679220617217}`
		newData := `{"id":1014336373145370625,"name":"resources manager","details":"","team_id":1014110679220617217}`
		query, err := PatchWithQuery([]byte(dbRec), []byte(newData), "public.project", "id", true, nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, `UPDATE public.project SET name='resources manager' WHERE id=1014336373145370625`, query)
	})
	t.Run("accurate filter with relationships", func(t *testing.T) {
		dbRec := `{"id":1014336373145370625,"name":"res-man","details":"details of project","team_id":1014110679220617217}`
		newData := `{"id":1014336373145370625,"name":"resources manager","details":"","team_id":1014110679220617217}`
		query, err := PatchWithQuery([]byte(dbRec), []byte(newData), "public.project", "id", true, map[string]string{
			"name": `"name"`,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, `UPDATE public.project SET "name"='resources manager' WHERE id=1014336373145370625`, query)
	})
}
