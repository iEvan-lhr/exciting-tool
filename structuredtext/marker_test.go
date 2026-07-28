package structuredtext

import (
	"errors"
	"testing"
)

func TestMarkerTokenizerAcrossChunks(t *testing.T) {
	tokenizer, err := NewMarkerTokenizer("(img:", ")")
	if err != nil {
		t.Fatal(err)
	}
	var tokens []Token
	for _, chunk := range []string{"before (i", "mg:abc", ") middle ", "(img:x)", " after"} {
		produced, err := tokenizer.Push(chunk)
		if err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, produced...)
	}
	produced, err := tokenizer.Flush()
	if err != nil {
		t.Fatal(err)
	}
	tokens = append(tokens, produced...)

	want := []Token{
		{Kind: TextToken, Value: "before "},
		{Kind: MarkerToken, Value: "abc"},
		{Kind: TextToken, Value: " middle "},
		{Kind: MarkerToken, Value: "x"},
		{Kind: TextToken, Value: " after"},
	}
	if len(tokens) != len(want) {
		t.Fatalf("tokens = %#v", tokens)
	}
	for index := range want {
		if tokens[index] != want[index] {
			t.Fatalf("token %d = %#v, want %#v", index, tokens[index], want[index])
		}
	}
}

func TestMarkerTokenizerFlushesIncompleteMarkerAsText(t *testing.T) {
	tokenizer, err := NewMarkerTokenizer("[id:", "]")
	if err != nil {
		t.Fatal(err)
	}
	first, err := tokenizer.Push("before [id:unfinished")
	if err != nil {
		t.Fatal(err)
	}
	last, err := tokenizer.Flush()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].Value != "before " {
		t.Fatalf("first tokens = %#v", first)
	}
	if len(last) != 1 || last[0] != (Token{Kind: TextToken, Value: "[id:unfinished"}) {
		t.Fatalf("last tokens = %#v", last)
	}
}

func TestMarkerTokenizerLimit(t *testing.T) {
	tokenizer, err := NewMarkerTokenizerLimit("<", ">", 4)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tokenizer.Push("<1234")
	if !errors.Is(err, ErrMarkerTooLarge) {
		t.Fatalf("error = %v", err)
	}

	tokenizer, err = NewMarkerTokenizerLimit("<", ">", 4)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tokenizer.Push("<1234>")
	if !errors.Is(err, ErrMarkerTooLarge) {
		t.Fatalf("complete marker error = %v", err)
	}
}

func TestMarkerTokenizerRejectsEmptyDelimiter(t *testing.T) {
	_, err := NewMarkerTokenizer("", "]")
	if !errors.Is(err, ErrEmptyMarkerDelimiter) {
		t.Fatalf("error = %v", err)
	}
}
