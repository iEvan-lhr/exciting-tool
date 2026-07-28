package structuredtext

import (
	"errors"
	"strings"
	"testing"
)

func TestStripFence(t *testing.T) {
	input := " \n```json\n{\"ok\":true}\n```\n "
	if got := StripFence(input); got != `{"ok":true}` {
		t.Fatalf("StripFence() = %q", got)
	}
	if got := StripFence("prefix ```json\n{}\n``` suffix"); got != "prefix ```json\n{}\n``` suffix" {
		t.Fatalf("surrounding text changed: %q", got)
	}
	if got := StripFence("~~~\n~~~"); got != "" {
		t.Fatalf("empty fence = %q", got)
	}
}

func TestExtractJSON(t *testing.T) {
	input := `Answer: {"text":"brace } and quote \"","items":[1,{"ok":true}]} done`
	got, ok := ExtractJSON(input)
	if !ok {
		t.Fatal("JSON not found")
	}
	want := `{"text":"brace } and quote \"","items":[1,{"ok":true}]}`
	if got != want {
		t.Fatalf("ExtractJSON() = %q", got)
	}
}

func TestJSONCandidates(t *testing.T) {
	got := JSONCandidates(`first {"a":1} second [2,3]`)
	if len(got) != 2 || got[0] != `{"a":1}` || got[1] != `[2,3]` {
		t.Fatalf("JSONCandidates() = %#v", got)
	}
}

func TestExtractJSONWithRepair(t *testing.T) {
	got, err := ExtractJSONWithRepair("result: {\"ok\":true,}", func(candidate string) (string, error) {
		return strings.ReplaceAll(candidate, ",}", "}"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"ok":true}` {
		t.Fatalf("repaired JSON = %q", got)
	}
}

func TestExtractJSONWithInvalidRepair(t *testing.T) {
	_, err := ExtractJSONWithRepair("{bad}", func(string) (string, error) {
		return "still bad", nil
	})
	if !errors.Is(err, ErrJSONNotFound) {
		t.Fatalf("error = %v", err)
	}
}

func FuzzExtractJSONDoesNotPanic(f *testing.F) {
	f.Add(`prefix {"a":[1,2]} suffix`)
	f.Add(`{"quote":"\\\"}"}`)
	f.Add(`[[[[`)
	f.Fuzz(func(t *testing.T, value string) {
		_, _ = ExtractJSON(value)
		_ = JSONCandidates(value)
	})
}
