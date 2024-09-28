# gobo-patcher
Simple map-based golang package to detect differences in JSON format and generate PostgreSQL update queries

## Installation

        $ go get github.com/gonzabosio/gobo-patcher

## Example
```go
package main

import (
	"fmt"

	"github.com/gonzabosio/gobo-patcher"
)

func main() {
	original := `{"name": "John Doe", "age": 32}`
	new := `{"name": "Jane Doe", "age": 30}`
	diff, err := gobo.JSONDiff([]byte(original), []byte(new))
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	for k, v := range diff {
		fmt.Println(k, v)
	}
	// output: name Jane Doe - age 30

	db := `{"id": 1234, "username": "johndoe", "age": 30}`
	update := `{"username": "janedoe"}`
	query, err := gobo.PatchWithQuery([]byte(db), []byte(update), "users", "id", nil)
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return
	}
	fmt.Println(query)
	// output: UPDATE "users" SET ("username"='janedoe') WHERE "id"=1234
}
```