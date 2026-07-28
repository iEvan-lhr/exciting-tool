# Migrating to v0.2

## Requirements

v0.2 requires Go 1.26 or newer.

## HTTP

Replace panic-based root helpers:

```go
body := tools.Do(url)
```

with an error-returning client:

```go
client := httpx.New(httpx.WithTimeout(10 * time.Second))
response, err := client.Do(ctx, http.MethodGet, url, nil, nil)
```

Use `GetJSON`, `PostJSON`, or `DoJSON` when decoding JSON.

## SQL

Legacy SQL helpers return complete SQL strings. New helpers return SQL and arguments separately:

```go
type User struct {
    ID   int    `db:"id,where"`
    Name string `db:"name"`
}

query, args, err := tools.UpdateArgs(User{ID: 7, Name: "Ada"})
```

The `db` tag supports:

- a column name: `db:"display_name"`
- an update filter: `db:"id,where"`
- an auto-generated field: `db:"id,auto"`
- an ignored field: `db:"-"`

Parameterized updates and deletes reject empty filters.

## String indexing

Replace panic-based access:

```go
value := text.GetByte(index)
```

with:

```go
value, ok := text.ByteAt(index)
part, err := text.Slice(start, end)
runeValue, ok := text.RuneAt(index)
runePart, err := text.SliceRunes(start, end)
```

## Locks

`LockFunc` now completes the function before returning. For manual locking:

```go
lock := tools.Lock("resource")
defer lock.Unlock()
```

## Release compatibility

The root package is retained, but HTTP shortcuts and unchecked string indexing are deprecated. Migrate incrementally; subpackages do not depend on the root compatibility API.
