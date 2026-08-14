package gui

import (
	"fmt"
	"testing"

	"github.com/troygilman/vent"
)

func TestNewTableFilters(t *testing.T) {
	got := newTableFilters([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: "admin"},
		{Name: "is_staff", Type: "bool", Value: ""},
		{Name: "is_active", Type: "bool", Value: vent.BoolFilterTrue.String()},
	})
	wantShow := fmt.Sprintf("$filter.email !== '' || $filter.is_staff !== '%s' || $filter.is_active !== '%s'", vent.BoolFilterAll, vent.BoolFilterAll)
	if got.ShowClear != wantShow {
		t.Fatalf("ShowClear = %q", got.ShowClear)
	}
	wantClear := fmt.Sprintf("$filter.email = ''; $filter.is_staff = '%s'; $filter.is_active = '%s'; @get(location.pathname, {filterSignals: {include: /^filter\\./}})", vent.BoolFilterAll, vent.BoolFilterAll)
	if got.OnClear != wantClear {
		t.Fatalf("OnClear = %q", got.OnClear)
	}
}
