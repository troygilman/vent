package gui

import (
	"fmt"
	"strconv"
)

func entitySignal(name string) string {
	return "entity." + name
}

func fkOptionsElementID(name string) string {
	return "entity-" + name + "-options"
}

func fkSearchSignal(name string) string {
	return "fksearch." + name
}

func fkOpenSignal(name string) string {
	return "fkopen." + name
}

func fkSelectedID(options []SelectOption) string {
	for _, opt := range options {
		if opt.Selected {
			return strconv.Itoa(opt.Value)
		}
	}
	return ""
}

func fkSelectedIDs(options []SelectOption) []string {
	ids := make([]string, 0, len(options))
	for _, opt := range options {
		if opt.Selected {
			ids = append(ids, strconv.Itoa(opt.Value))
		}
	}
	return ids
}

func fkSelectedLabel(options []SelectOption) string {
	for _, opt := range options {
		if opt.Selected {
			return opt.Label
		}
	}
	return ""
}

func FkOptionsFetch(optionsURL, name string) string {
	return fmt.Sprintf(
		`@get(%q + '?q=' + encodeURIComponent($fksearch.%s || ""), {filterSignals: {include: /^entity\.%s$/}})`,
		optionsURL,
		name,
		name,
	)
}

func fkFetchIfOpen(optionsURL, name string) string {
	return fmt.Sprintf("$fkopen.%s && %s", name, FkOptionsFetch(optionsURL, name))
}

func fkSelectUnique(name string, id int, label string) string {
	quoted := strconv.Quote(label)
	return fmt.Sprintf(
		"$fkopen.%s = false; $entity.%s = '%d'; $fklabel.%s = %s; $fksearch.%s = \"\"",
		name, name, id, name, quoted, name,
	)
}

func fkClearUnique(name string) string {
	return fmt.Sprintf("$fkopen.%s = false; $entity.%s = \"\"; $fklabel.%s = \"\"; $fksearch.%s = \"\"", name, name, name, name)
}

func fkHasLabelExpr(name string) string {
	return fmt.Sprintf("$fklabel.%s !== ''", name)
}

func fkLabelTextExpr(name string) string {
	return "$fklabel." + name
}

func fkRenderChipsEffect(name string) string {
	return fmt.Sprintf("fkCombobox.renderChips(el, $fkchips.%s)", name)
}

func fkRemoveManyClick(name string) string {
	return fmt.Sprintf(
		"evt.target.closest('.fk-clear') && ($entity.%s = $entity.%s.filter((id) => id !== evt.target.closest('.fk-clear').dataset.id), $fkchips.%s = $fkchips.%s.filter((c) => String(c.id) !== evt.target.closest('.fk-clear').dataset.id))",
		name, name, name, name,
	)
}

func fkOpen(name string) string {
	return fmt.Sprintf("$fkopen.%s = true", name)
}

func fkClose(name string) string {
	return fmt.Sprintf("$fkopen.%s = false, $fksearch.%s = \"\"", name, name)
}

func fkOnInput(name string) string {
	return fkOpen(name)
}

func fkBlur(name string) string {
	return fmt.Sprintf(
		"(!evt.relatedTarget || !evt.currentTarget.closest('.fk-combobox').contains(evt.relatedTarget)) && (%s)",
		fkClose(name),
	)
}

func fkEscape(name string) string {
	return fmt.Sprintf("evt.key === 'Escape' && (%s)", fkClose(name))
}

func fkKeydown(name string) string {
	return fmt.Sprintf("fkCombobox.keydown(evt, %q); %s", name, fkEscape(name))
}

func fkIsOpenExpr(name string) string {
	return "$" + fkOpenSignal(name)
}

func fkAddMany(name string, id int, label string) string {
	quoted := strconv.Quote(label)
	return fmt.Sprintf(
		"!$entity.%s.includes('%d') && ($entity.%s = [...$entity.%s, '%d'], $fkchips.%s = [...$fkchips.%s, {id:'%d', label:%s}])",
		name, id, name, name, id, name, name, id, quoted,
	)
}

type fkChipJSON struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func fkSelectedChips(options []SelectOption) []fkChipJSON {
	chips := make([]fkChipJSON, 0, len(options))
	for _, opt := range options {
		if opt.Selected {
			chips = append(chips, fkChipJSON{ID: strconv.Itoa(opt.Value), Label: opt.Label})
		}
	}
	return chips
}

func fkUniqueSignals(name, selectedID, selectedLabel string) map[string]any {
	return map[string]any{
		"entity": map[string]any{
			name: selectedID,
		},
		"fksearch": map[string]any{
			name: "",
		},
		"fklabel": map[string]any{
			name: selectedLabel,
		},
		"fkopen": map[string]any{
			name: false,
		},
	}
}

func fkManySignals(name string, chips []SelectOption) map[string]any {
	selectedIDs := fkSelectedIDs(chips)
	return map[string]any{
		"entity": map[string]any{
			name: selectedIDs,
		},
		"fkchips": map[string]any{
			name: fkSelectedChips(chips),
		},
		"fksearch": map[string]any{
			name: "",
		},
		"fkopen": map[string]any{
			name: false,
		},
	}
}
