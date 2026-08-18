package gui

import (
	"testing"

	"github.com/troygilman/vent"
)

func TestTableFiltersActive(t *testing.T) {
	if tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: ""},
		{Name: "is_staff", Type: "bool", Value: vent.BoolFilterAll.String()},
	}) {
		t.Fatal("empty filters should be inactive")
	}
	if !tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: "admin"},
	}) {
		t.Fatal("string value should be active")
	}
	if !tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "is_staff", Type: "bool", Value: vent.BoolFilterTrue.String()},
	}) {
		t.Fatal("bool true should be active")
	}
	if tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "is_staff", Type: "bool", Value: ""},
	}) {
		t.Fatal("empty bool should be inactive")
	}
}

func TestTableFilterChipValue(t *testing.T) {
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "string", Value: "Dune",
	}); got != "Dune" {
		t.Fatalf("string chip = %q, want Dune", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "int", Value: "412",
	}); got != "412" {
		t.Fatalf("int chip = %q, want 412", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterTrue.String(),
	}); got != "Yes" {
		t.Fatalf("bool true chip = %q, want Yes", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterFalse.String(),
	}); got != "No" {
		t.Fatalf("bool false chip = %q, want No", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterAll.String(),
	}); got != "" {
		t.Fatalf("bool all chip = %q, want empty", got)
	}
}

func TestTableFilterClearClick(t *testing.T) {
	got := tableFilterClearClick("title")
	want := "$filter.title = ''; el.closest('form').dispatchEvent(new Event('change'))"
	if got != want {
		t.Fatalf("clear click = %q, want %q", got, want)
	}
}
