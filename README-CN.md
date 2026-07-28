# exciting-tool

[English](./README.md) | 简体中文

exciting-tool 是一个面向 Go 1.26+ 的轻量工具集合。v0.2 将新 API 拆分为聚焦、类型安全的子包，同时保留根包兼容层，方便现有项目逐步迁移。

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

支持 context、超时、响应大小限制、状态码和 JSON：

```go
client := httpx.New(
    httpx.WithTimeout(10*time.Second),
    httpx.WithMaxBodyBytes(2<<20),
)

var result User
response, err := client.GetJSON(ctx, endpoint, &result)
```

非 2xx JSON 请求会返回 `*httpx.StatusError`。普通 `Do` 始终返回状态码和 Header，由调用者决定如何处理。

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
