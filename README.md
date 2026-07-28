# exciting-tool

English | [简体中文](./README-CN.md)

exciting-tool is a lightweight utility collection for Go 1.26+. The v0.3 API
adds streaming HTTP and structured LLM text support while the root package
remains available as a compatibility layer.

## Install

```bash
go get github.com/iEvan-lhr/exciting-tool
```

## Packages

- `textutil`: bounds-safe text extraction and rune slicing.
- `httpx`: buffered JSON, streaming responses, multipart uploads, validation, and retries.
- `structuredtext`: JSON extraction, optional repair integration, and streaming marker tokenization.
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

### Streaming multipart

Large request and response bodies do not need to be buffered:

```go
form := httpx.NewMultipart()
_ = form.AddField("format", "pdf")
_ = form.AddFile("files", pptxPath)

response, err := client.PostMultipartStream(ctx, endpoint, form, nil)
if err != nil {
    return err
}
defer response.Close()
if err := response.CheckStatus(64 << 10); err != nil {
    return err
}
if err := response.RequireContentType("application/pdf"); err != nil {
    return err
}
response.LimitBody(100 << 20)
_, err = io.Copy(output, response.Body)
```

Retries are opt-in. The default methods are idempotent; include `POST`
explicitly only when the endpoint can safely receive it more than once:

```go
client := httpx.New(httpx.WithRetry(httpx.RetryPolicy{
    MaxAttempts:         3,
    BaseDelay:           200 * time.Millisecond,
    MaxDelay:            2 * time.Second,
    RetryTransportErrors: true,
    RespectRetryAfter:   true,
}))
```

### Structured model output

```go
jsonText, ok := structuredtext.ExtractJSON(llmResponse)

tokenizer, _ := structuredtext.NewMarkerTokenizer("(img:", ")")
tokens, err := tokenizer.Push(streamChunk)
```

`ExtractJSONWithRepair` accepts a repair callback, so applications can reuse
their preferred JSON repair library without adding one to exciting-tool.

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

The original `tools.String`, `Do`, and SQL string APIs remain available. New
code should use `NewHTTPClient`, `InsertArgs`, `QueryArgs`, `UpdateArgs`, and
the focused subpackages. Legacy HTTP shortcuts are deprecated because they
report errors through panic.

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
