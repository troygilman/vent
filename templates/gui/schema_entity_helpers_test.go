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
	chipsRoot, _ := got["fkchips"].(map[string]any)
	chips, _ := chipsRoot["groups"].([]fkChipJSON)
	if chips == nil {
		t.Fatal("fkchips must init [] not nil")
	}
	if len(chips) != 0 {
		t.Fatalf("fkchips = %#v, want empty", chips)
	}
}

func TestFkManySignalsHydratesChips(t *testing.T) {
	got := fkManySignals("groups", []SelectOption{{Value: 7, Label: "Staff", Selected: true}})
	entity, _ := got["entity"].(map[string]any)
	ids, _ := entity["groups"].([]string)
	if len(ids) != 1 || ids[0] != "7" {
		t.Fatalf("ids = %#v", ids)
	}
	chipsRoot, _ := got["fkchips"].(map[string]any)
	chips, _ := chipsRoot["groups"].([]fkChipJSON)
	if len(chips) != 1 || chips[0].ID != "7" || chips[0].Label != "Staff" {
		t.Fatalf("fkchips = %#v", chips)
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

func TestFkFetchIfOpenGatesGet(t *testing.T) {
	got := fkFetchIfOpen("/admin/reviews/options/book/", "book")
	if !strings.HasPrefix(got, "$fkopen.book && @get(") {
		t.Fatalf("fetch-if-open = %q", got)
	}
}

func TestFkSelectUniqueDoesNotFetch(t *testing.T) {
	got := fkSelectUnique("book", 9, "Needle Title")
	if strings.Contains(got, "@get") {
		t.Fatalf("pick must not @get: %q", got)
	}
	if !strings.Contains(got, "$entity.book = '9'") || !strings.Contains(got, "$fklabel.book") {
		t.Fatalf("pick = %q", got)
	}
}

func TestFkAddManyUpdatesChips(t *testing.T) {
	got := fkAddMany("groups", 3, "Staff")
	if strings.Contains(got, "@get") {
		t.Fatalf("M2M pick must not @get: %q", got)
	}
	if !strings.Contains(got, "$fkchips.groups") {
		t.Fatalf("M2M pick = %q", got)
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

func TestFkOnInputReopensDropdown(t *testing.T) {
	if got := fkOnInput("book"); got != "$fkopen.book = true" {
		t.Fatalf("fkOnInput = %q, want open the dropdown while typing", got)
	}
}

func TestFkBlurDoesNotEmbedSemicolons(t *testing.T) {
	got := fkBlur("book")
	if strings.Contains(got, ";") {
		t.Fatalf("Datastar && groups cannot contain semicolons: %q", got)
	}
}
