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
		`@get(%q + '?q=' + encodeURIComponent($fksearch.%s || ''), {filterSignals: {include: /^entity\.%s$/}})`,
		optionsURL,
		name,
		name,
	)
}

func fkSelectUnique(name string, id int, label string) string {
	quoted := strconv.Quote(label)
	return fmt.Sprintf(
		"$entity.%s = '%d'; $fklabel.%s = %s; $fksearch.%s = ''; $fkopen.%s = false",
		name, id, name, quoted, name, name,
	)
}

func fkClearUnique(name string) string {
	return fmt.Sprintf("$entity.%s = ''; $fklabel.%s = ''; $fksearch.%s = ''; $fkopen.%s = false", name, name, name, name)
}

func fkOpen(name string) string {
	return fmt.Sprintf("$fkopen.%s = true", name)
}

func fkClose(name string) string {
	return fmt.Sprintf("$fkopen.%s = false; $fksearch.%s = ''", name, name)
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

func fkAddMany(name string, id int) string {
	return fmt.Sprintf(
		"!$entity.%s.includes('%d') && ($entity.%s = [...$entity.%s, '%d'])",
		name, id, name, name, id,
	)
}

func fkRemoveMany(name string, id int) string {
	return fmt.Sprintf(
		"$entity.%s = $entity.%s.filter((id) => id !== '%d')",
		name, name, id,
	)
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

func fkManySignals(name string, selectedIDs []string) map[string]any {
	if selectedIDs == nil {
		selectedIDs = []string{}
	}
	return map[string]any{
		"entity": map[string]any{
			name: selectedIDs,
		},
		"fksearch": map[string]any{
			name: "",
		},
		"fklabel": map[string]any{
			name: "",
		},
		"fkopen": map[string]any{
			name: false,
		},
	}
}
