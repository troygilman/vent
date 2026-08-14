package gui

import (
	"encoding/json"
	"testing"
)

func TestBoolFilterValue(t *testing.T) {
	if got := BoolFilterValue(nil); got != "" {
		t.Fatalf("nil = %q, want empty", got)
	}
	yes, no := true, false
	if got := BoolFilterValue(&yes); got != "true" {
		t.Fatalf("true = %q, want true", got)
	}
	if got := BoolFilterValue(&no); got != "false" {
		t.Fatalf("false = %q, want false", got)
	}
}

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
	want := `{"filter":{"email":"admin","is_active":true,"is_staff":null}}`
	if string(raw) != want {
		t.Fatalf("signals JSON = %s, want %s", raw, want)
	}
}

func TestTableFilterBoolExprs(t *testing.T) {
	columns := []SchemaTableFilterableColumn{
		{Name: "email", Type: "string"},
		{Name: "is_staff", Type: "bool"},
	}
	if got := tableFilterActiveExpr(columns); got != "$filter.email !== '' || $filter.is_staff !== null" {
		t.Fatalf("active expr = %q", got)
	}
	if got := tableFilterClearExpr(columns); got != "$filter.email = ''; $filter.is_staff = null; @get(location.pathname, {filterSignals: {include: /^filter\\./}})" {
		t.Fatalf("clear expr = %q", got)
	}
}
