# exciting-tool

English | [简体中文](./README-CN.md)

exciting-tool is a lightweight utility collection for Go 1.26+. The v0.2 API is organized into focused, type-safe subpackages while the root package remains available as a compatibility layer.

## Install

```bash
go get github.com/iEvan-lhr/exciting-tool
```

## Packages

- `textutil`: bounds-safe text extraction and rune slicing.
- `httpx`: context-aware HTTP, bounded response reads, status handling, and JSON helpers.
- `sqlbuilder`: deterministic parameterized SQL for MySQL, PostgreSQL, and SQLite.
- `orderedmap`: a generic concurrency-safe map that preserves insertion order.

### HTTP

```go
client := httpx.New(
    httpx.WithTimeout(10*time.Second),
    httpx.WithMaxBodyBytes(2<<20),
)

var result User
response, err := client.GetJSON(ctx, endpoint, &result)
```

### Parameterized SQL

```go
type User struct {
    ID   int    `db:"id,where"`
    Name string `db:"name"`
}

query, args, err := sqlbuilder.New(sqlbuilder.PostgreSQL).
    UpdateStruct(User{ID: 7, Name: "Ada"})
```

Updates and deletes without filters return `sqlbuilder.ErrUnsafeMutation`.

### Ordered map

```go
values := orderedmap.New[string, int]()
values.Set("first", 1)
values.Set("second", 2)
keys := values.Keys()
```

## Compatibility

The original `tools.String`, `Do`, and SQL string APIs remain available. New code should use `NewHTTPClient`, `InsertArgs`, `QueryArgs`, `UpdateArgs`, and the focused subpackages. Legacy HTTP shortcuts are deprecated because they report errors through panic.

See [MIGRATION.md](./MIGRATION.md) before upgrading an existing application and [CHANGELOG.md](./CHANGELOG.md) for release details.

## Development

```bash
go test ./...
go vet ./...
go test -race ./...
```

## License

Project code is Apache-2.0. Files derived from the Go standard library retain
their original notices under [LICENSE-GO](./LICENSE-GO), as documented in
[THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).
