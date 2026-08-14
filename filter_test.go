package vent

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestIsDatastarRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/admin/users/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if IsDatastarRequest(req) {
		t.Fatal("expected false without header")
	}
	req.Header.Set("Datastar-Request", "true")
	if !IsDatastarRequest(req) {
		t.Fatal("expected true with header")
	}
}

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
		{"all", false, false},
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
	if got := BoolFilter("all").Normalize(); got != BoolFilterAll {
		t.Fatalf("legacy all = %q, want empty", got)
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
