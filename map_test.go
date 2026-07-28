package tools

import "testing"

func TestMap(t *testing.T) {
	var ordered Map
	ordered.Set("first", 1)
	ordered.Set("second", 2)
	ordered.Set("first", 3)
	if ordered.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", ordered.Len())
	}
	value, ok := ordered.Get("first")
	if !ok || value != 3 {
		t.Fatalf("Get(first) = %v, %v; want 3, true", value, ok)
	}
}

func TestSpider(t *testing.T) {
	spider := MakeSpider(0, "999999")
	for i := 1; i < 1000; i++ {
		spider.Add(i, 1000-i)
	}
	if spider.Len() != 1000 {
		t.Fatalf("Len() = %d, want 1000", spider.Len())
	}
	if value := spider.Get(232); value != 768 {
		t.Fatalf("Get(232) = %v, want 768", value)
	}
}
