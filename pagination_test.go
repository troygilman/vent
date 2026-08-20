package vent

import (
	"reflect"
	"testing"
)

func TestParseListPage(t *testing.T) {
	cases := []struct {
		raw      string
		pageSize int
		wantPage int
		wantSize int
	}{
		{"", 50, 1, 50},
		{"0", 50, 1, 50},
		{"-3", 50, 1, 50},
		{"nope", 50, 1, 50},
		{"2", 50, 2, 50},
		{"2", 0, 2, DefaultListPageSize},
		{"2", -10, 2, DefaultListPageSize},
	}
	for _, tc := range cases {
		got := ParseListPage(tc.raw, tc.pageSize)
		if got.Page != tc.wantPage || got.PageSize != tc.wantSize || got.Total != 0 {
			t.Fatalf("ParseListPage(%q, %d) = %+v, want page %d size %d total 0",
				tc.raw, tc.pageSize, got, tc.wantPage, tc.wantSize)
		}
	}
}

func TestListPageWithTotalClamps(t *testing.T) {
	got := ParseListPage("99", 10).WithTotal(25)
	if got.Page != 3 || got.Total != 25 {
		t.Fatalf("clamp = %+v, want page 3 total 25", got)
	}

	empty := ParseListPage("4", 10).WithTotal(0)
	if empty.Page != 1 || empty.Total != 0 || empty.TotalPages() != 0 {
		t.Fatalf("empty = %+v pages %d, want page 1 total 0 pages 0", empty, empty.TotalPages())
	}

	neg := ParseListPage("1", 10).WithTotal(-5)
	if neg.Total != 0 || neg.Page != 1 {
		t.Fatalf("neg total = %+v, want total 0 page 1", neg)
	}
}

func TestListPageOffsetLimitFromTo(t *testing.T) {
	p := ParseListPage("2", 10).WithTotal(25)
	if p.Offset() != 10 || p.Limit() != 10 {
		t.Fatalf("offset/limit = %d/%d, want 10/10", p.Offset(), p.Limit())
	}
	if p.From() != 11 || p.To() != 20 {
		t.Fatalf("from/to = %d/%d, want 11/20", p.From(), p.To())
	}
	if !p.HasPrev() || !p.HasNext() || p.TotalPages() != 3 {
		t.Fatalf("nav = prev %v next %v pages %d", p.HasPrev(), p.HasNext(), p.TotalPages())
	}

	last := ParseListPage("3", 10).WithTotal(25)
	if last.From() != 21 || last.To() != 25 || last.HasNext() {
		t.Fatalf("last = from %d to %d next %v", last.From(), last.To(), last.HasNext())
	}

	first := ParseListPage("", 10).WithTotal(25)
	if first.Offset() != 0 || first.From() != 1 || first.To() != 10 || first.HasPrev() {
		t.Fatalf("first = %+v from %d to %d prev %v", first, first.From(), first.To(), first.HasPrev())
	}

	empty := ParseListPage("1", 10).WithTotal(0)
	if empty.From() != 0 || empty.To() != 0 || empty.Offset() != 0 {
		t.Fatalf("empty from/to/offset = %d/%d/%d", empty.From(), empty.To(), empty.Offset())
	}
}

func TestListPageSignal(t *testing.T) {
	if got := ParseListPage("1", 10).Signal(); got != "" {
		t.Fatalf("page 1 signal = %q, want empty", got)
	}
	if got := ParseListPage("2", 10).Signal(); got != "2" {
		t.Fatalf("page 2 signal = %q, want 2", got)
	}
}

func TestListPageWindow(t *testing.T) {
	got := ParseListPage("5", 10).WithTotal(100).Window(2)
	want := []ListPageItem{
		{Page: 1},
		{Ellipsis: true},
		{Page: 3},
		{Page: 4},
		{Page: 5, Current: true},
		{Page: 6},
		{Page: 7},
		{Ellipsis: true},
		{Page: 10},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("window = %#v, want %#v", got, want)
	}

	small := ParseListPage("1", 10).WithTotal(30).Window(2)
	wantSmall := []ListPageItem{
		{Page: 1, Current: true},
		{Page: 2},
		{Page: 3},
	}
	if !reflect.DeepEqual(small, wantSmall) {
		t.Fatalf("small window = %#v, want %#v", small, wantSmall)
	}

	if got := ParseListPage("1", 10).WithTotal(0).Window(2); got != nil {
		t.Fatalf("empty window = %#v, want nil", got)
	}
}
