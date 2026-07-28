package textutil

import "testing"

func TestBetween(t *testing.T) {
	got, ok := Between("before<a>中文</a>after", "<a>", "</a>")
	if !ok || got != "中文" {
		t.Fatalf("Between() = %q, %v", got, ok)
	}
	if _, ok := Between("plain", "<a>", "</a>"); ok {
		t.Fatal("Between() found missing delimiters")
	}
}

func TestAllBetween(t *testing.T) {
	got := AllBetween("<b>one</b><b>二</b>", "<b>", "</b>")
	if len(got) != 2 || got[0] != "one" || got[1] != "二" {
		t.Fatalf("AllBetween() = %#v", got)
	}
}

func TestRuneHelpers(t *testing.T) {
	got, err := SliceRunes("A中文B", 1, 3)
	if err != nil || got != "中文" {
		t.Fatalf("SliceRunes() = %q, %v", got, err)
	}
	if _, err := SliceRunes("short", -1, 2); err == nil {
		t.Fatal("SliceRunes() accepted an invalid range")
	}
	truncated, err := TruncateRunes("一二三四", 2, "...")
	if err != nil || truncated != "一二..." {
		t.Fatalf("TruncateRunes() = %q, %v", truncated, err)
	}
}
