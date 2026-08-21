package gui

import (
	"strings"
	"testing"
)

func TestFkManySignalsInitEmpty(t *testing.T) {
	got := fkManySignals("groups", nil)
	entity, _ := got["entity"].(map[string]any)
	ids, _ := entity["groups"].([]string)
	if ids == nil {
		t.Fatal("M2M add form must init [] not nil")
	}
	if len(ids) != 0 {
		t.Fatalf("ids = %#v, want empty slice", ids)
	}
}

func TestFkOptionsFetchIncludesEntitySignal(t *testing.T) {
	got := FkOptionsFetch("/admin/reviews/options/book/", "book")
	if !strings.Contains(got, `filterSignals: {include: /^entity\.book$/}`) {
		t.Fatalf("fetch %q missing entity filter", got)
	}
	if !strings.Contains(got, `/admin/reviews/options/book/`) {
		t.Fatalf("fetch %q missing url", got)
	}
}
