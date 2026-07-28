# exciting-tool

[English](./README.md) | 简体中文

exciting-tool 是一个面向 Go 1.26+ 的轻量工具集合。v0.3 增加了流式 HTTP
和结构化模型文本处理，同时保留根包兼容层，方便现有项目逐步迁移。

## 安装

```bash
go get github.com/iEvan-lhr/exciting-tool
```

要求 Go 1.26 或更高版本。

## 子包

### textutil

安全提取文本，不使用无边界下标：

```go
value, ok := textutil.Between("before<a>内容</a>after", "<a>", "</a>")
items := textutil.AllBetween("<b>one</b><b>two</b>", "<b>", "</b>")
part, err := textutil.SliceRunes("A中文B", 1, 3)
```

### httpx

支持 context、超时、响应大小限制、状态码、JSON、重试和流式传输：

```go
client := httpx.New(
    httpx.WithTimeout(10*time.Second),
    httpx.WithMaxBodyBytes(2<<20),
)

var result User
response, err := client.GetJSON(ctx, endpoint, &result)
```

非 2xx JSON 请求会返回 `*httpx.StatusError`。普通 `Do` 始终返回状态码和 Header，由调用者决定如何处理。

大文件上传和下载不需要完整缓存在内存中：

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

重试默认关闭。默认重试方法均为幂等方法；只有服务端允许重复处理时，才应显式加入 `POST`：

```go
client := httpx.New(httpx.WithRetry(httpx.RetryPolicy{
    MaxAttempts:          3,
    BaseDelay:            200 * time.Millisecond,
    MaxDelay:             2 * time.Second,
    RetryTransportErrors: true,
    RespectRetryAfter:    true,
}))
```

### structuredtext

从模型输出中提取 JSON，或处理跨流式 chunk 的标记：

```go
jsonText, ok := structuredtext.ExtractJSON(llmResponse)

tokenizer, _ := structuredtext.NewMarkerTokenizer("(img:", ")")
tokens, err := tokenizer.Push(streamChunk)
```

`ExtractJSONWithRepair` 接受修复回调，因此业务项目可以继续使用已有的
JSON 修复库，`exciting-tool` 不会额外绑定一套实现。

### sqlbuilder

生成参数化 SQL，不把值拼入语句：

```go
type User struct {
    ID   int    `db:"id,where"`
    Name string `db:"name"`
}

query, args, err := sqlbuilder.New(sqlbuilder.PostgreSQL).
    UpdateStruct(User{ID: 7, Name: "Ada"})
// UPDATE "user" SET "name" = $1 WHERE "id" = $2
// args: []any{"Ada", 7}
```

`Update` 和 `Delete` 在没有过滤条件时返回 `ErrUnsafeMutation`。

### orderedmap

泛型、并发安全并保留插入顺序：

```go
values := orderedmap.New[string, int]()
values.Set("first", 1)
values.Set("second", 2)
keys := values.Keys()
```

## 兼容层

原有 `tools.String`、`Do`、`Query` 等 API 仍保留。新代码建议使用：

- `NewHTTPClient` 或 `httpx.New`
- `InsertArgs`、`QueryArgs`、`UpdateArgs`
- `String.ByteAt`、`String.Slice`、`String.RuneAt`、`String.SliceRunes`

旧的 HTTP 快捷函数依赖 panic 传递错误，已标记为 Deprecated。

## 开发

```bash
go test ./...
go vet ./...
go test -race ./...
```

升级现有项目请阅读 [MIGRATION.md](./MIGRATION.md)，版本变化见 [CHANGELOG.md](./CHANGELOG.md)。

## License

项目代码使用 Apache-2.0。部分源文件改编自 Go 标准库，并按照
[LICENSE-GO](./LICENSE-GO) 和 [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md)
保留其原始版权与许可证说明。
