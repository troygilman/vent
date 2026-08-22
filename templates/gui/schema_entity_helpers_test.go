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
	if !strings.Contains(got, `encodeURIComponent($fksearch.book || "")`) {
		t.Fatalf("fetch %q should always send the search query", got)
	}
	if strings.Contains(got, `!== ($fklabel`) {
		t.Fatalf("fetch must not drop q when search equals the selected label: %q", got)
	}
}

func TestFkUniqueSignalsKeepSearchEmpty(t *testing.T) {
	got := fkUniqueSignals("book", "12", "Needle Title")
	search, _ := got["fksearch"].(map[string]any)
	if search["book"] != "" {
		t.Fatalf("fksearch must stay the query, not the selected label, got %#v", search["book"])
	}
	label, _ := got["fklabel"].(map[string]any)
	if label["book"] != "Needle Title" {
		t.Fatalf("fklabel = %#v, want Needle Title", label["book"])
	}
}

func TestFkBlurDoesNotEmbedSemicolons(t *testing.T) {
	got := fkBlur("book")
	if strings.Contains(got, ";") {
		t.Fatalf("Datastar && groups cannot contain semicolons: %q", got)
	}
}
