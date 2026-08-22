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
	if !strings.Contains(got, `$fklabel.book`) {
		t.Fatalf("fetch %q should omit q when the input still holds the selected label", got)
	}
}

func TestFkUniqueSignalsShowLabelWhenClosed(t *testing.T) {
	got := fkUniqueSignals("book", "12", "Needle Title")
	search, _ := got["fksearch"].(map[string]any)
	if search["book"] != "Needle Title" {
		t.Fatalf("closed unique input should show selected label, got %#v", search["book"])
	}
}
