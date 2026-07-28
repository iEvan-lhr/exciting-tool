package sqlbuilder

import (
	"errors"
	"reflect"
	"testing"
)

func TestMapBuilders(t *testing.T) {
	mysql := New(MySQL)
	query, args, err := mysql.Select("users", map[string]any{"id": 7, "active": true})
	if err != nil {
		t.Fatal(err)
	}
	if query != "SELECT * FROM `users` WHERE `active` = ? AND `id` = ?" {
		t.Fatalf("Select() = %q", query)
	}
	if !reflect.DeepEqual(args, []any{true, 7}) {
		t.Fatalf("args = %#v", args)
	}

	postgres := New(PostgreSQL)
	query, args, err = postgres.Insert("public.users", map[string]any{"name": "Ada", "age": 36})
	if err != nil {
		t.Fatal(err)
	}
	if query != `INSERT INTO "public"."users" ("age", "name") VALUES ($1, $2)` {
		t.Fatalf("Insert() = %q", query)
	}
	if !reflect.DeepEqual(args, []any{36, "Ada"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestUnsafeMutationRejected(t *testing.T) {
	builder := New(MySQL)
	if _, _, err := builder.Update("users", map[string]any{"name": "Ada"}, nil); !errors.Is(err, ErrUnsafeMutation) {
		t.Fatalf("Update error = %v", err)
	}
	if _, _, err := builder.Delete("users", nil); !errors.Is(err, ErrUnsafeMutation) {
		t.Fatalf("Delete error = %v", err)
	}
}

func TestStructBuilders(t *testing.T) {
	type UserProfile struct {
		ID        int    `db:"id,where"`
		Name      string `db:"display_name"`
		CreatedAt string `db:",auto"`
	}
	model := UserProfile{ID: 7, Name: "O'Reilly"}
	builder := New(PostgreSQL)
	query, args, err := builder.UpdateStruct(model)
	if err != nil {
		t.Fatal(err)
	}
	if query != `UPDATE "user_profile" SET "display_name" = $1 WHERE "id" = $2` {
		t.Fatalf("UpdateStruct() = %q", query)
	}
	if !reflect.DeepEqual(args, []any{"O'Reilly", 7}) {
		t.Fatalf("args = %#v", args)
	}
}
