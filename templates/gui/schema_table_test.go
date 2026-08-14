package gui

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/troygilman/vent"
)

func TestTableFilterSignalsBoolJSON(t *testing.T) {
	got := tableFilterSignals([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: "admin"},
		{Name: "is_staff", Type: "bool", Value: ""},
		{Name: "is_active", Type: "bool", Value: vent.BoolFilterTrue},
	})
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(`{"filter":{"email":"admin","is_active":%q,"is_staff":%q}}`, vent.BoolFilterTrue, vent.BoolFilterAll)
	if string(raw) != want {
		t.Fatalf("signals JSON = %s, want %s", raw, want)
	}
}

func TestTableFilterValueNormalizesBool(t *testing.T) {
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: ""}); got != vent.BoolFilterAll {
		t.Fatalf("empty bool = %q, want all", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: vent.BoolFilterAll}); got != vent.BoolFilterAll {
		t.Fatalf("all = %q, want all", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: vent.BoolFilterTrue}); got != vent.BoolFilterTrue {
		t.Fatalf("true = %q", got)
	}
	if got := tableFilterValue(SchemaTableFilterableColumn{Type: "bool", Value: vent.BoolFilterFalse}); got != vent.BoolFilterFalse {
		t.Fatalf("false = %q", got)
	}
}

func TestTableFilterBoolExprs(t *testing.T) {
	columns := []SchemaTableFilterableColumn{
		{Name: "email", Type: "string"},
		{Name: "is_staff", Type: "bool"},
	}
	wantActive := fmt.Sprintf("$filter.email !== '' || $filter.is_staff !== '%s'", vent.BoolFilterAll)
	if got := tableFilterActiveExpr(columns); got != wantActive {
		t.Fatalf("active expr = %q", got)
	}
	wantClear := fmt.Sprintf("$filter.email = ''; $filter.is_staff = '%s'; @get(location.pathname, {filterSignals: {include: /^filter\\./}})", vent.BoolFilterAll)
	if got := tableFilterClearExpr(columns); got != wantClear {
		t.Fatalf("clear expr = %q", got)
	}
}
