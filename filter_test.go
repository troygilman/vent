package vent

import "testing"

func TestParseBoolFilter(t *testing.T) {
	cases := []struct {
		in     string
		want   bool
		wantOK bool
	}{
		{BoolFilterTrue, true, true},
		{BoolFilterFalse, false, true},
		{BoolFilterAll, false, false},
		{"", false, false},
	}
	for _, tc := range cases {
		got, ok := ParseBoolFilter(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("ParseBoolFilter(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestNormalizeBoolFilter(t *testing.T) {
	if got := NormalizeBoolFilter(BoolFilterTrue); got != BoolFilterTrue {
		t.Fatalf("true = %q", got)
	}
	if got := NormalizeBoolFilter(BoolFilterFalse); got != BoolFilterFalse {
		t.Fatalf("false = %q", got)
	}
	if got := NormalizeBoolFilter(""); got != BoolFilterAll {
		t.Fatalf("empty = %q, want all", got)
	}
	if got := NormalizeBoolFilter(BoolFilterAll); got != BoolFilterAll {
		t.Fatalf("all = %q", got)
	}
}
