package vent

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDefaultOptionLimit(t *testing.T) {
	if DefaultOptionLimit != 100 {
		t.Fatalf("DefaultOptionLimit = %d, want 100", DefaultOptionLimit)
	}
}

func TestProcessOptionSearchValue(t *testing.T) {
	if got := ProcessOptionSearchValue("  Title  "); got != "Title" {
		t.Fatalf("ProcessOptionSearchValue = %q, want Title", got)
	}
	if got := ProcessOptionSearchValue("\n"); got != "" {
		t.Fatalf("ProcessOptionSearchValue blank = %q, want empty", got)
	}
}

func TestUnionSearchResultsByID(t *testing.T) {
	type row struct{ ID int }
	search := []row{{1}, {2}, {3}}
	selected := []row{{3}, {2502}}
	got := UnionSearchResultsByID(search, selected, func(r row) int { return r.ID })
	want := []row{{1}, {2}, {3}, {2502}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UnionSearchResultsByID = %#v, want %#v", got, want)
	}
}

func TestOptionEdgeFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/options/user/", nil)
	req.SetPathValue("edge", "user")
	if got := OptionEdgeFromRequest(req); got != "user" {
		t.Fatalf("PathValue edge = %q, want user", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/options/groups/", nil)
	if got := OptionEdgeFromRequest(req); got != "groups" {
		t.Fatalf("literal path edge = %q, want groups", got)
	}
}
