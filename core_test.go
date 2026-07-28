package tools

import (
	stdjson "encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtoiNegativeDoesNotMutate(t *testing.T) {
	value := Make("-123")
	got, err := value.Atoi()
	if err != nil {
		t.Fatal(err)
	}
	if got != -123 {
		t.Fatalf("Atoi() = %d, want -123", got)
	}
	if value.String() != "-123" {
		t.Fatalf("Atoi mutated receiver to %q", value.String())
	}
}

func TestUnicodeExtraction(t *testing.T) {
	value := Make("前【中文】后")
	if got := value.GetRune("【】").String(); got != "中文" {
		t.Fatalf("GetRune() = %q, want 中文", got)
	}
	if got := value.GetContent("<missing>", "</missing>"); got != "" {
		t.Fatalf("GetContent() = %q, want empty", got)
	}
}

func TestStringFloat64AndRunes(t *testing.T) {
	value := Make("中")
	value.Append("文")
	if got := string(value.Runes()); got != "中文" {
		t.Fatalf("Runes() = %q, want 中文", got)
	}
	if got := Make(float64(1e100)).String(); strings.Contains(got, "Inf") {
		t.Fatalf("float64 was reduced to float32: %q", got)
	}
	if !value.Check([]rune("中文")) {
		t.Fatal("Check([]rune) used stale rune data")
	}
	quoted := Make("a\nb")
	Quote(quoted)
	if quoted.String() != `"a\nb"` {
		t.Fatalf("Quote() = %q", quoted.String())
	}
}

func TestJSONRoundTripString(t *testing.T) {
	type payload struct {
		Name  string  `json:"name"`
		Count *String `json:"count"`
	}
	var decoded payload
	Unmarshal(`{"name":"<tool>","count":12}`, &decoded)
	if decoded.Name != "<tool>" || decoded.Count.String() != "12" {
		t.Fatalf("Unmarshal decoded %+v", decoded)
	}
	if got := string(Marshal(decoded)); !strings.Contains(got, "<tool>") {
		t.Fatalf("Marshal escaped HTML: %s", got)
	}
	literal := struct {
		Value string `json:"value"`
	}{Value: `\u003c`}
	data := Marshal(literal)
	var roundTrip struct {
		Value string `json:"value"`
	}
	if err := stdjson.Unmarshal(data, &roundTrip); err != nil || roundTrip.Value != literal.Value {
		t.Fatalf("Marshal literal escape = %s, %v", data, err)
	}
}

func TestHTTPHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Method))
	}))
	defer server.Close()

	if got := Do(server.URL).String(); got != http.MethodGet {
		t.Fatalf("Do(GET) = %q", got)
	}
	if got := Do(server.URL, "body").String(); got != http.MethodPost {
		t.Fatalf("Do(POST) = %q", got)
	}
}

func TestSpiderConcurrentAccess(t *testing.T) {
	spider := MakeSpider("initial", 0)
	var wait sync.WaitGroup
	for i := 0; i < 100; i++ {
		wait.Add(1)
		go func(value int) {
			defer wait.Done()
			spider.Add(value, value)
			if got := spider.Get(value); got != value {
				t.Errorf("Get(%d) = %v", value, got)
			}
		}(i)
	}
	wait.Wait()
	if spider.Len() != 101 {
		t.Fatalf("Len() = %d, want 101", spider.Len())
	}
}

func TestLockFuncSerializesByName(t *testing.T) {
	var active int32
	var maximum int32
	var wait sync.WaitGroup
	for i := 0; i < 10; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			LockFunc("serial", func() {
				current := atomic.AddInt32(&active, 1)
				for {
					previous := atomic.LoadInt32(&maximum)
					if current <= previous || atomic.CompareAndSwapInt32(&maximum, previous, current) {
						break
					}
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt32(&active, -1)
			})
		}()
	}
	wait.Wait()
	if maximum != 1 {
		t.Fatalf("maximum concurrency = %d, want 1", maximum)
	}
	namedLockRegistry.Lock()
	remaining := len(namedLockRegistry.entries)
	namedLockRegistry.Unlock()
	if remaining != 0 {
		t.Fatalf("named lock registry retained %d entries", remaining)
	}
}

func TestSQLGeneration(t *testing.T) {
	type UserProfile struct {
		ID   int    `marshal:"check" db:"id,where"`
		Name string `db:"name"`
	}
	model := UserProfile{ID: 7, Name: "O'Reilly"}
	query := Query(model)
	if !strings.Contains(query, "user_profile") || !strings.Contains(query, "O''Reilly") {
		t.Fatalf("Query() = %q", query)
	}
	update := Update(model)
	if !strings.Contains(update, "set `name`='O''Reilly'") || !strings.Contains(update, "where `id`='7'") {
		t.Fatalf("Update() = %q", update)
	}
	if create := Create(model); !strings.Contains(create, "CREATE TABLE `user_profile`") {
		t.Fatalf("Create() = %q", create)
	}
	if unsafe := Update(struct{ Name string }{Name: "all rows"}); unsafe != "" {
		t.Fatalf("Update without check field = %q", unsafe)
	}
	parameterized, args, err := UpdateArgs(UserProfile{ID: 7, Name: "Ada"})
	if err != nil || parameterized != "UPDATE `user_profile` SET `name` = ? WHERE `id` = ?" || len(args) != 2 {
		t.Fatalf("UpdateArgs() = %q, %#v, %v", parameterized, args, err)
	}
}

func TestBoundsSafeStringAccess(t *testing.T) {
	value := Make("A中文")
	if got, ok := value.RuneAt(2); !ok || got != '文' {
		t.Fatalf("RuneAt() = %q, %v", got, ok)
	}
	if _, ok := value.ByteAt(99); ok {
		t.Fatal("ByteAt accepted an invalid index")
	}
	if got, err := value.SliceRunes(1, 3); err != nil || got != "中文" {
		t.Fatalf("SliceRunes() = %q, %v", got, err)
	}
}

func TestUpdateLayoutReturnsParseError(t *testing.T) {
	value := Make("not-a-date")
	got, err := value.UpdateLayout()
	if err == nil || got != "not-a-date" {
		t.Fatalf("UpdateLayout() = %q, %v", got, err)
	}
}
