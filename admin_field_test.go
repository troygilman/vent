package vent

import (
	"testing"
	"time"
)

func TestFormatFormValueTime(t *testing.T) {
	if got := FormatFormValue(mustParseTime(t, "2024-05-01T15:04:00Z")); got != "2024-05-01T15:04" {
		t.Fatalf("FormatFormValue() = %q", got)
	}
}

func TestFormatFormValuePointers(t *testing.T) {
	t.Parallel()

	s := "A clear tour of Vent's schema features."
	if got := FormatFormValue(&s); got != s {
		t.Fatalf("FormatFormValue(*string) = %q, want %q", got, s)
	}
	var nilString *string
	if got := FormatFormValue(nilString); got != "" {
		t.Fatalf("FormatFormValue(nil *string) = %q, want empty", got)
	}

	n := 5
	if got := FormatFormValue(&n); got != "5" {
		t.Fatalf("FormatFormValue(*int) = %q, want 5", got)
	}

	b := true
	if got := FormatFormValue(&b); got != "true" {
		t.Fatalf("FormatFormValue(*bool) = %q, want true", got)
	}

	tm := mustParseTime(t, "2024-05-01T15:04:00Z")
	if got := FormatFormValue(&tm); got != "2024-05-01T15:04" {
		t.Fatalf("FormatFormValue(*time.Time) = %q", got)
	}
	var nilTime *time.Time
	if got := FormatFormValue(nilTime); got != "" {
		t.Fatalf("FormatFormValue(nil *time.Time) = %q, want empty", got)
	}

	if got := FormatFormValue(nil); got != "" {
		t.Fatalf("FormatFormValue(nil) = %q, want empty", got)
	}
}

func TestPasswordHashIsSet(t *testing.T) {
	hash := "secret"
	if !PasswordHashIsSet(&hash) {
		t.Fatal("PasswordHashIsSet(&hash) = false, want true")
	}
	if PasswordHashIsSet(nil) {
		t.Fatal("PasswordHashIsSet(nil) = true, want false")
	}
}
