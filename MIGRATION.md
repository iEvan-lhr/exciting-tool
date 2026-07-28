# Migrating to v0.3

## Requirements

v0.3 requires Go 1.26 or newer.

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

`Do` remains buffered and bounded. Use `DoStream` when the response should be
copied directly to a file or another writer:

```go
response, err := client.DoStream(ctx, http.MethodGet, url, nil, nil)
if err != nil {
    return err
}
defer response.Close()
response.LimitBody(100 << 20)
_, err = io.Copy(destination, response.Body)
```

The caller must close every successful streaming response. `CheckStatus`
consumes and closes only non-2xx bodies.

## Multipart requests

Replace manually buffered `multipart.Writer` code with `httpx.Multipart`:

```go
form := httpx.NewMultipart()
_ = form.AddFile("files", path)
response, err := client.PostMultipartStream(ctx, endpoint, form, nil)
```

`AddBytes`, `AddFile`, and `AddReaderFunc` can recreate their bodies for a
retry. `AddReader` is intentionally one-shot and is rejected when POST retries
are enabled.

## Retries

Retries are disabled unless `WithRetry` is used. `MaxAttempts` includes the
initial request. A custom `Request.Body` must return a fresh body every time.
`DoStream` accepts the replayable reader types recognized by `net/http`; it
returns `ErrBodyNotReplayable` before sending when the configured method may
retry but the supplied body cannot be replayed.

## Structured model output

Use `structuredtext.ExtractJSON` instead of trimming Markdown fences and JSON
substrings independently in each service. Use `ExtractJSONWithRepair` to
connect an existing repair dependency. `MarkerTokenizer` preserves delimiters
split across streaming chunks.

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
