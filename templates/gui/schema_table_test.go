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
