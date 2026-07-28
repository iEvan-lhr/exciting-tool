# Changelog

## Unreleased (v0.2.0)

### Added

- Go 1.26 module and CI baseline.
- `textutil` bounds-safe text helpers.
- `httpx` context-aware client, JSON helpers, status errors, and response limits.
- `sqlbuilder` parameterized MySQL, PostgreSQL, and SQLite statements.
- Generic concurrency-safe `orderedmap`.
- Root migration APIs: `NewHTTPClient`, `InsertArgs`, `QueryArgs`, and `UpdateArgs`.
- Bounds-safe `String` byte and rune accessors.

### Changed

- `LockFunc` executes synchronously while serializing by name.
- Named locks are removed from the registry after the final user releases them.
- Root `Update` refuses to produce an UPDATE without a `marshal:"check"` field.
- Root JSON marshaling now uses standard encoding with HTML escaping disabled.
- `Update` and `Create` return generated SQL.

### Fixed

- Negative integer parsing mutated the input and lost its sign.
- Unicode extraction could panic or loop forever.
- HTTP response bodies were not closed and default requests had no timeout.
- Float64 values were formatted through float32.
- `Quote` did not update its input and could panic during escaping.
- Experimental maps and locks contained data races.

### Deprecated

- Root `Do`, `DoReq`, and `DoUseHeader`; use `httpx.Client`.
- Root `String.GetByte` and `String.GetStr`; use bounds-safe accessors.
