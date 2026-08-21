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
	return fmt.Sprintf("$entity.%s = '%d'; $fklabel.%s = %s", name, id, name, strconv.Quote(label))
}

func fkClearUnique(name string) string {
	return fmt.Sprintf("$entity.%s = ''; $fklabel.%s = ''", name, name)
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
	}
}
