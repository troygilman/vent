package vent

import (
	"encoding/json"
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

func TestOptionEdgeFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/admin/reviews/options/user/", "user"},
		{"/options/user/", "user"},
		{"/admin/reviews/options/book/?q=Needle", "book"},
		{"/admin/reviews/options/user/?q=ada&datastar=%7B%7D", "user"},
		{"/admin/reviews/options/not-an-edge/", "not-an-edge"},
		{"/admin/reviews/1/", ""},
	}
	for _, tt := range tests {
		if got := OptionEdgeFromPath(tt.path); got != tt.want {
			t.Fatalf("OptionEdgeFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
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

func TestParseOptionValue(t *testing.T) {
	id, ok := ParseOptionValue("42")
	if !ok || id != 42 {
		t.Fatalf("ParseOptionValue(42) = %d, %v", id, ok)
	}
	if _, ok := ParseOptionValue(""); ok {
		t.Fatal("empty should be unset")
	}
	if _, ok := ParseOptionValue("nope"); ok {
		t.Fatal("invalid should be unset")
	}
}

func TestParseOptionIDs(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []int
	}{
		{name: "null", raw: "null", want: nil},
		{name: "empty", raw: "", want: nil},
		{name: "unique string", raw: `"12"`, want: []int{12}},
		{name: "unique empty", raw: `""`, want: nil},
		{name: "m2m strings", raw: `["1","2"]`, want: []int{1, 2}},
		{name: "m2m empty", raw: `[]`, want: []int{}},
		{name: "int array", raw: `[3,4]`, want: []int{3, 4}},
		{name: "int", raw: `7`, want: []int{7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw json.RawMessage
			if tt.raw != "" {
				raw = json.RawMessage(tt.raw)
			}
			got := ParseOptionIDs(raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseOptionIDs(%s) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}
