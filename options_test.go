package vent

import (
	"reflect"
	"testing"
)

func TestDefaultOptionLimit(t *testing.T) {
	if DefaultOptionLimit != 100 {
		t.Fatalf("DefaultOptionLimit = %d, want 100", DefaultOptionLimit)
	}
}

func TestOptionSearch(t *testing.T) {
	if got := OptionSearch("  Title  "); got != "Title" {
		t.Fatalf("OptionSearch = %q, want Title", got)
	}
	if got := OptionSearch("\n"); got != "" {
		t.Fatalf("OptionSearch blank = %q, want empty", got)
	}
}

func TestUnionByID(t *testing.T) {
	type row struct{ ID int }
	search := []row{{1}, {2}, {3}}
	selected := []row{{3}, {2502}}
	got := UnionByID(search, selected, func(r row) int { return r.ID })
	want := []row{{1}, {2}, {3}, {2502}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UnionByID = %#v, want %#v", got, want)
	}
}
