package gui

import (
	"encoding/json"
	"testing"
)

func TestTableFilterSignalsBoolJSON(t *testing.T) {
	got := tableFilterSignals([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: "admin"},
		{Name: "is_staff", Type: "bool", Value: ""},
		{Name: "is_active", Type: "bool", Value: "true"},
	})
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"filter":{"email":"admin","is_active":"true","is_staff":"all"}}`
	if string(raw) != want {
		t.Fatalf("signals JSON = %s, want %s", raw, want)
	}
}

func TestTableFilterValueNormalizesBool(t *testing.T) {
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: ""}); got != "all" {
		t.Fatalf("empty bool = %q, want all", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: "all"}); got != "all" {
		t.Fatalf("all = %q, want all", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: "true"}); got != "true" {
		t.Fatalf("true = %q", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: "false"}); got != "false" {
		t.Fatalf("false = %q", got)
	}
}

func TestTableFilterBoolExprs(t *testing.T) {
	columns := []SchemaTableFilterableColumn{
		{Name: "email", Type: "string"},
		{Name: "is_staff", Type: "bool"},
	}
	if got := tableFilterActiveExpr(columns); got != "$filter.email !== '' || $filter.is_staff !== 'all'" {
		t.Fatalf("active expr = %q", got)
	}
	if got := tableFilterClearExpr(columns); got != "$filter.email = ''; $filter.is_staff = 'all'; @get(location.pathname, {filterSignals: {include: /^filter\\./}})" {
		t.Fatalf("clear expr = %q", got)
	}
}
