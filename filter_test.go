package vent

import (
	"encoding/json"
	"testing"
)

func TestBoolFilterBool(t *testing.T) {
	cases := []struct {
		in     BoolFilter
		want   bool
		wantOK bool
	}{
		{BoolFilterTrue, true, true},
		{BoolFilterFalse, false, true},
		{BoolFilterAll, false, false},
		{"", false, false},
	}
	for _, tc := range cases {
		got, ok := tc.in.Bool()
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("%q.Bool() = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestBoolFilterNormalize(t *testing.T) {
	if got := BoolFilterTrue.Normalize(); got != BoolFilterTrue {
		t.Fatalf("true = %q", got)
	}
	if got := BoolFilterFalse.Normalize(); got != BoolFilterFalse {
		t.Fatalf("false = %q", got)
	}
	if got := BoolFilter("").Normalize(); got != BoolFilterAll {
		t.Fatalf("empty = %q, want all", got)
	}
	if got := BoolFilterAll.Normalize(); got != BoolFilterAll {
		t.Fatalf("all = %q", got)
	}
}

func TestBoolFilterJSON(t *testing.T) {
	var signals struct {
		Filter struct {
			IsStaff BoolFilter `json:"is_staff"`
		} `json:"filter"`
	}
	if err := json.Unmarshal([]byte(`{"filter":{"is_staff":"false"}}`), &signals); err != nil {
		t.Fatal(err)
	}
	if signals.Filter.IsStaff != BoolFilterFalse {
		t.Fatalf("IsStaff = %q, want false", signals.Filter.IsStaff)
	}
	v, ok := signals.Filter.IsStaff.Bool()
	if !ok || v {
		t.Fatalf("Bool() = %v, %v; want false, true", v, ok)
	}
}
