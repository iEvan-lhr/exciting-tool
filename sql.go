package tools

import (
	"reflect"

	"github.com/iEvan-lhr/exciting-tool/sqlbuilder"
)

const sli = '_'

// Save assembles one or more INSERT statements.
// Deprecated: Use InsertArgs or sqlbuilder.Builder.InsertStruct.
func Save(model any) (result []*String) {
	return marshalStruct(model)
}

// Query assembles a SELECT statement from non-zero struct fields.
// Deprecated: Use QueryArgs or sqlbuilder.Builder.SelectStruct.
func Query(model any) string {
	s := String{}
	s.queryStruct(model)
	return s.string()
}

// Update assembles an UPDATE statement. Fields tagged marshal:"check" are used
// in the WHERE clause; other non-zero fields are included in SET.
// Deprecated: Use UpdateArgs or sqlbuilder.Builder.UpdateStruct.
func Update(model any) string {
	values, typ := returnValAndTyp(model)
	if !values.IsValid() || typ == nil || values.Kind() != reflect.Struct {
		return ""
	}

	query := Make("update ", humpName(typ.Name()), " set ")
	setCount := 0
	where := Make()
	whereCount := 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := values.Field(i)
		if field.PkgPath != "" || value.IsZero() || field.Tag.Get("marshal") == "off" {
			continue
		}
		literal := righteousCharacter(Make(value.Interface()))
		if field.Tag.Get("marshal") == "check" {
			if whereCount > 0 {
				where.Append(" and ")
			}
			where.Append("`", humpName(field.Name), "`='", literal, "'")
			whereCount++
			continue
		}
		if setCount > 0 {
			query.Append(",")
		}
		query.Append("`", humpName(field.Name), "`='", literal, "'")
		setCount++
	}
	if setCount == 0 {
		return ""
	}
	if whereCount == 0 {
		return ""
	}
	query.Append(" where ", where)
	return query.String()
}

// Check assembles a SELECT using fields tagged marshal:"check".
// Deprecated: Use QueryArgs or sqlbuilder.Builder.Select.
func Check(model any) string {
	s := String{}
	s.checkStruct(model)
	return s.string()
}

// Create assembles a CREATE TABLE statement.
// Deprecated: Use a schema migration tool for production DDL.
func Create(model any) string {
	if model == nil {
		return ""
	}
	return marshalTable(model).String()
}

// InsertArgs builds a parameterized INSERT using db struct tags.
func InsertArgs(model any, dialect ...sqlbuilder.Dialect) (string, []any, error) {
	return sqlbuilder.New(selectDialect(dialect)).InsertStruct(model)
}

// QueryArgs builds a parameterized SELECT using non-zero struct fields.
func QueryArgs(model any, dialect ...sqlbuilder.Dialect) (string, []any, error) {
	return sqlbuilder.New(selectDialect(dialect)).SelectStruct(model)
}

// UpdateArgs builds a parameterized UPDATE. At least one field must use the
// db:",where" option.
func UpdateArgs(model any, dialect ...sqlbuilder.Dialect) (string, []any, error) {
	return sqlbuilder.New(selectDialect(dialect)).UpdateStruct(model)
}

func selectDialect(dialects []sqlbuilder.Dialect) sqlbuilder.Dialect {
	if len(dialects) > 0 {
		return dialects[0]
	}
	return sqlbuilder.MySQL
}
