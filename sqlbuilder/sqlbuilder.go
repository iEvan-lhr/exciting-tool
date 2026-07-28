// Package sqlbuilder builds deterministic parameterized SQL statements.
package sqlbuilder

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

var (
	ErrEmptyTable     = errors.New("table name cannot be empty")
	ErrNoValues       = errors.New("at least one value is required")
	ErrUnsafeMutation = errors.New("UPDATE and DELETE require at least one filter")
	ErrInvalidModel   = errors.New("model must be a struct or non-nil pointer to struct")
)

type Dialect uint8

const (
	MySQL Dialect = iota
	PostgreSQL
	SQLite
)

type Builder struct {
	dialect Dialect
}

func New(dialect Dialect) Builder {
	return Builder{dialect: dialect}
}

func (b Builder) Select(table string, filters map[string]any) (string, []any, error) {
	quotedTable, err := b.quoteIdentifier(table)
	if err != nil {
		return "", nil, err
	}
	query := "SELECT * FROM " + quotedTable
	where, args, err := b.conditions(filters, 1)
	if err != nil {
		return "", nil, err
	}
	if where != "" {
		query += " WHERE " + where
	}
	return query, args, nil
}

func (b Builder) Insert(table string, values map[string]any) (string, []any, error) {
	quotedTable, err := b.quoteIdentifier(table)
	if err != nil {
		return "", nil, err
	}
	keys := sortedKeys(values)
	if len(keys) == 0 {
		return "", nil, ErrNoValues
	}
	columns := make([]string, 0, len(keys))
	placeholders := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		column, err := b.quoteIdentifier(key)
		if err != nil {
			return "", nil, err
		}
		columns = append(columns, column)
		placeholders = append(placeholders, b.placeholder(index+1))
		args = append(args, values[key])
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		quotedTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	return query, args, nil
}

func (b Builder) Update(
	table string,
	values map[string]any,
	filters map[string]any,
) (string, []any, error) {
	quotedTable, err := b.quoteIdentifier(table)
	if err != nil {
		return "", nil, err
	}
	keys := sortedKeys(values)
	if len(keys) == 0 {
		return "", nil, ErrNoValues
	}
	if len(filters) == 0 {
		return "", nil, ErrUnsafeMutation
	}

	assignments := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys)+len(filters))
	for index, key := range keys {
		column, err := b.quoteIdentifier(key)
		if err != nil {
			return "", nil, err
		}
		assignments = append(assignments, column+" = "+b.placeholder(index+1))
		args = append(args, values[key])
	}
	where, filterArgs, err := b.conditions(filters, len(args)+1)
	if err != nil {
		return "", nil, err
	}
	args = append(args, filterArgs...)
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		quotedTable,
		strings.Join(assignments, ", "),
		where,
	)
	return query, args, nil
}

func (b Builder) Delete(table string, filters map[string]any) (string, []any, error) {
	quotedTable, err := b.quoteIdentifier(table)
	if err != nil {
		return "", nil, err
	}
	if len(filters) == 0 {
		return "", nil, ErrUnsafeMutation
	}
	where, args, err := b.conditions(filters, 1)
	if err != nil {
		return "", nil, err
	}
	return "DELETE FROM " + quotedTable + " WHERE " + where, args, nil
}

// InsertStruct uses the struct type name as the table name. The db tag may
// rename a field, skip it with "-", or mark it as auto-generated with "auto".
func (b Builder) InsertStruct(model any) (string, []any, error) {
	value, typ, err := structValue(model)
	if err != nil {
		return "", nil, err
	}
	values := make(map[string]any)
	for _, field := range modelFields(value, typ) {
		if !field.auto {
			values[field.name] = field.value
		}
	}
	return b.Insert(snakeCase(typ.Name()), values)
}

// SelectStruct uses non-zero fields as filters.
func (b Builder) SelectStruct(model any) (string, []any, error) {
	value, typ, err := structValue(model)
	if err != nil {
		return "", nil, err
	}
	filters := make(map[string]any)
	for _, field := range modelFields(value, typ) {
		if field.where || !field.zero {
			filters[field.name] = field.value
		}
	}
	return b.Select(snakeCase(typ.Name()), filters)
}

// UpdateStruct requires at least one field tagged with the "where" db option.
func (b Builder) UpdateStruct(model any) (string, []any, error) {
	value, typ, err := structValue(model)
	if err != nil {
		return "", nil, err
	}
	values := make(map[string]any)
	filters := make(map[string]any)
	for _, field := range modelFields(value, typ) {
		if field.where {
			filters[field.name] = field.value
		} else if !field.auto {
			values[field.name] = field.value
		}
	}
	return b.Update(snakeCase(typ.Name()), values, filters)
}

func (b Builder) conditions(filters map[string]any, start int) (string, []any, error) {
	keys := sortedKeys(filters)
	if len(keys) == 0 {
		return "", nil, nil
	}
	conditions := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	argumentIndex := start
	for _, key := range keys {
		column, err := b.quoteIdentifier(key)
		if err != nil {
			return "", nil, err
		}
		value := filters[key]
		if value == nil {
			conditions = append(conditions, column+" IS NULL")
			continue
		}
		conditions = append(conditions, column+" = "+b.placeholder(argumentIndex))
		args = append(args, value)
		argumentIndex++
	}
	return strings.Join(conditions, " AND "), args, nil
}

func (b Builder) placeholder(index int) string {
	if b.dialect == PostgreSQL {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

func (b Builder) quoteIdentifier(identifier string) (string, error) {
	if identifier == "" || strings.ContainsRune(identifier, 0) {
		return "", ErrEmptyTable
	}
	quote := "`"
	if b.dialect == PostgreSQL || b.dialect == SQLite {
		quote = `"`
	}
	parts := strings.Split(identifier, ".")
	for index, part := range parts {
		if part == "" {
			return "", ErrEmptyTable
		}
		parts[index] = quote + strings.ReplaceAll(part, quote, quote+quote) + quote
	}
	return strings.Join(parts, "."), nil
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func structValue(model any) (reflect.Value, reflect.Type, error) {
	if model == nil {
		return reflect.Value{}, nil, ErrInvalidModel
	}
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, nil, ErrInvalidModel
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, nil, ErrInvalidModel
	}
	return value, value.Type(), nil
}

type modelField struct {
	name  string
	value any
	auto  bool
	where bool
	zero  bool
}

func modelFields(value reflect.Value, typ reflect.Type) []modelField {
	var fields []modelField
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		if field.PkgPath != "" {
			continue
		}
		name, options := parseTag(field)
		if name == "-" {
			continue
		}
		fieldValue := value.Field(index)
		fields = append(fields, modelField{
			name:  name,
			value: fieldValue.Interface(),
			auto:  options["auto"],
			where: options["where"],
			zero:  fieldValue.IsZero(),
		})
	}
	return fields
}

func parseTag(field reflect.StructField) (string, map[string]bool) {
	name := snakeCase(field.Name)
	options := make(map[string]bool)
	tag := field.Tag.Get("db")
	if tag == "" {
		return name, options
	}
	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		name = parts[0]
	}
	for _, option := range parts[1:] {
		options[option] = true
	}
	return name, options
}

func snakeCase(value string) string {
	runes := []rune(value)
	var result strings.Builder
	for index, current := range runes {
		if unicode.IsUpper(current) {
			if index > 0 {
				previous := runes[index-1]
				nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
				if unicode.IsLower(previous) || unicode.IsDigit(previous) || nextIsLower {
					result.WriteByte('_')
				}
			}
			current = unicode.ToLower(current)
		}
		result.WriteRune(current)
	}
	return result.String()
}
