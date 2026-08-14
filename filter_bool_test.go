package vent

import (
	"encoding/json"
	"testing"
)

func TestFilterBoolUnmarshal(t *testing.T) {
	cases := []struct {
		in      string
		wantSet bool
		wantVal bool
	}{
		{`true`, true, true},
		{`false`, true, false},
		{`"true"`, true, true},
		{`"false"`, true, false},
		{`null`, false, false},
		{`""`, false, false},
	}
	for _, tc := range cases {
		var f FilterBool
		if err := json.Unmarshal([]byte(tc.in), &f); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tc.in, err)
		}
		if f.set != tc.wantSet || f.val != tc.wantVal {
			t.Fatalf("Unmarshal(%s) = set=%v val=%v, want set=%v val=%v", tc.in, f.set, f.val, tc.wantSet, tc.wantVal)
		}
	}
}

func TestFilterBoolUnmarshalRejectsGarbage(t *testing.T) {
	var f FilterBool
	if err := json.Unmarshal([]byte(`"yes"`), &f); err == nil {
		t.Fatal("expected error")
	}
}

func TestFilterBoolStructPayload(t *testing.T) {
	var signals struct {
		Filter struct {
			Email    string     `json:"email"`
			IsStaff  FilterBool `json:"is_staff"`
			IsActive FilterBool `json:"is_active"`
		} `json:"filter"`
	}
	in := []byte(`{"filter":{"email":"","is_staff":"","is_active":false}}`)
	if err := json.Unmarshal(in, &signals); err != nil {
		t.Fatal(err)
	}
	if signals.Filter.IsStaff.Ptr() != nil {
		t.Fatal("empty is_staff should be unset")
	}
	if p := signals.Filter.IsActive.Ptr(); p == nil || *p {
		t.Fatalf("is_active = %v, want false", signals.Filter.IsActive.Ptr())
	}
}
