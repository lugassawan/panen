package shared

import "testing"

type testItem struct {
	ID   string
	Name string
}

func TestIndexBy(t *testing.T) {
	items := []testItem{
		{ID: "a", Name: "Alpha"},
		{ID: "b", Name: "Beta"},
		{ID: "c", Name: "Charlie"},
	}

	m := IndexBy(items, func(i testItem) string { return i.ID })

	if len(m) != 3 {
		t.Fatalf("got %d entries, want 3", len(m))
	}
	if m["a"].Name != "Alpha" {
		t.Errorf("got %q, want Alpha", m["a"].Name)
	}
	if m["b"].Name != "Beta" {
		t.Errorf("got %q, want Beta", m["b"].Name)
	}
	if m["c"].Name != "Charlie" {
		t.Errorf("got %q, want Charlie", m["c"].Name)
	}
}

func TestIndexByEmpty(t *testing.T) {
	m := IndexBy([]string{}, func(s string) string { return s })
	if len(m) != 0 {
		t.Fatalf("got %d entries, want 0", len(m))
	}
}

func TestIndexByDuplicateKeys(t *testing.T) {
	items := []string{"apple", "avocado", "banana"}
	m := IndexBy(items, func(s string) byte { return s[0] })

	if len(m) != 2 {
		t.Fatalf("got %d entries, want 2", len(m))
	}
	// Last item with same key wins.
	if m['a'] != "avocado" {
		t.Errorf("got %q, want avocado", m['a'])
	}
}
