package sqlbuilder_test

import (
	"fmt"

	"github.com/iEvan-lhr/exciting-tool/sqlbuilder"
)

func ExampleBuilder_Update() {
	builder := sqlbuilder.New(sqlbuilder.PostgreSQL)
	query, args, _ := builder.Update(
		"users",
		map[string]any{"name": "Ada"},
		map[string]any{"id": 7},
	)
	fmt.Println(query)
	fmt.Println(args)
	// Output:
	// UPDATE "users" SET "name" = $1 WHERE "id" = $2
	// [Ada 7]
}
