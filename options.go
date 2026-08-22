package vent

import (
	"encoding/json"
	"strconv"
	"strings"
)

// DefaultOptionLimit is the hard cap on FK option search hits.
// Selected IDs are loaded separately and are not counted against this cap.
const DefaultOptionLimit = 100

// OptionSearch trims an FK option search query.
func OptionSearch(raw string) string {
	return strings.TrimSpace(raw)
}

// OptionEdgeFromPath returns the FK edge name from an /options/{edge}/ path.
// Nested StripPrefix routers can leave URL.Path and PathValue("edge") pointing
// at another group's wildcard, so RequestURI is the reliable source.
func OptionEdgeFromPath(path string) string {
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	const marker = "/options/"
	i := strings.LastIndex(path, marker)
	if i < 0 {
		return ""
	}
	rest := strings.Trim(path[i+len(marker):], "/")
	if rest == "" {
		return ""
	}
	if j := strings.IndexByte(rest, '/'); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// UnionByID merges search hits with selected entities, preserving search order
// and appending selected rows that were outside the search window.
// LoadOptions uses this only for an empty query (the first page of options).
func UnionByID[T any](search []T, selected []T, id func(T) int) []T {
	seen := make(map[int]struct{}, len(search)+len(selected))
	out := make([]T, 0, len(search)+len(selected))
	appendUnique := func(items []T) {
		for _, item := range items {
			key := id(item)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	appendUnique(search)
	appendUnique(selected)
	return out
}

// ParseOptionValue parses a unique FK signal (string id). Empty is unset.
func ParseOptionValue(raw string) (id int, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return id, true
}

// ParseOptionIDs parses unique (string) or M2M ([]string) Datastar FK signals.
func ParseOptionIDs(raw json.RawMessage) []int {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return optionIDsFromStrings(values)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		if id, ok := ParseOptionValue(value); ok {
			return []int{id}
		}
		return nil
	}

	var nums []int
	if err := json.Unmarshal(raw, &nums); err == nil {
		out := make([]int, 0, len(nums))
		for _, n := range nums {
			out = append(out, n)
		}
		return out
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return []int{n}
	}

	return nil
}

func optionIDsFromStrings(values []string) []int {
	out := make([]int, 0, len(values))
	for _, value := range values {
		if id, ok := ParseOptionValue(value); ok {
			out = append(out, id)
		}
	}
	return out
}
