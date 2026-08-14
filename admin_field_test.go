package vent

import (
	"testing"
)

func TestFormatFormValueTime(t *testing.T) {
	if got := FormatFormValue(mustParseTime(t, "2024-05-01T15:04:00Z")); got != "2024-05-01T15:04" {
		t.Fatalf("FormatFormValue() = %q", got)
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
