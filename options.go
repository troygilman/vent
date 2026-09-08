package vent

import (
	"net/http"
	"strings"
)

// DefaultOptionLimit is the hard cap on FK option search hits.
// Selected IDs are loaded separately and are not counted against this cap.
const DefaultOptionLimit = 100

// ProcessOptionSearchValue trims an FK option search query.
func ProcessOptionSearchValue(raw string) string {
	return strings.TrimSpace(raw)
}

// OptionEdgeFromRequest returns the FK edge name for an options GET.
// It prefers PathValue("edge") and otherwise uses the last path segment so
// literal /options/{edgeName}/ routes (password schemas) resolve per request.
func OptionEdgeFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if edge := strings.TrimSpace(r.PathValue("edge")); edge != "" {
		return edge
	}
	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// UnionSearchResultsByID merges search hits with selected entities, preserving search order
// and appending selected rows that were outside the search window.
// LoadOptions uses this only for an empty query (the first page of options).
func UnionSearchResultsByID[T any](search []T, selected []T, id func(T) int) []T {
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
