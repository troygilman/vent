package vent

import "strconv"

// DefaultListPageSize is the list page size when VentSchemaAnnotation.PageSize is unset.
const DefaultListPageSize = 100

// ListPage is a 1-based offset page over a filtered list query.
type ListPage struct {
	Page     int
	PageSize int
	Total    int
}

// ListPageItem is one slot in a compact page-number window.
type ListPageItem struct {
	Page     int
	Current  bool
	Ellipsis bool
}

// ParseListPage parses the Datastar $page signal.
// Empty, zero, negative, and non-numeric values become page 1.
// Non-positive pageSize falls back to DefaultListPageSize.
func ParseListPage(raw string, pageSize int) ListPage {
	if pageSize <= 0 {
		pageSize = DefaultListPageSize
	}
	page := 1
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		page = n
	}
	return ListPage{Page: page, PageSize: pageSize}
}

// WithTotal records the filtered row count and clamps Page onto the last page.
func (p ListPage) WithTotal(total int) ListPage {
	if total < 0 {
		total = 0
	}
	p.Total = total
	pages := p.TotalPages()
	if pages == 0 {
		p.Page = 1
		return p
	}
	if p.Page > pages {
		p.Page = pages
	}
	if p.Page < 1 {
		p.Page = 1
	}
	return p
}

// Limit is the Ent Limit() for this page.
func (p ListPage) Limit() int {
	if p.PageSize <= 0 {
		return DefaultListPageSize
	}
	return p.PageSize
}

// Offset is the Ent Offset() for this page.
func (p ListPage) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

// TotalPages is ceil(Total / PageSize), or 0 when there are no rows.
func (p ListPage) TotalPages() int {
	if p.Limit() <= 0 || p.Total <= 0 {
		return 0
	}
	return (p.Total + p.Limit() - 1) / p.Limit()
}

// From is the 1-based index of the first row on this page, or 0 when empty.
func (p ListPage) From() int {
	if p.Total <= 0 {
		return 0
	}
	return p.Offset() + 1
}

// To is the 1-based index of the last row on this page, or 0 when empty.
func (p ListPage) To() int {
	if p.Total <= 0 {
		return 0
	}
	to := p.Offset() + p.Limit()
	if to > p.Total {
		return p.Total
	}
	return to
}

// HasPrev reports whether a previous page exists.
func (p ListPage) HasPrev() bool {
	return p.Page > 1
}

// HasNext reports whether a next page exists.
func (p ListPage) HasNext() bool {
	return p.Page < p.TotalPages()
}

// Signal is the $page value to bind: empty for page 1 so the URL stays clean.
func (p ListPage) Signal() string {
	if p.Page <= 1 {
		return ""
	}
	return strconv.Itoa(p.Page)
}

// Window returns first, last, and current±radius page numbers, with ellipsis gaps.
func (p ListPage) Window(radius int) []ListPageItem {
	total := p.TotalPages()
	if total <= 0 {
		return nil
	}
	if radius < 0 {
		radius = 0
	}

	show := make(map[int]struct{}, total)
	show[1] = struct{}{}
	show[total] = struct{}{}
	for i := p.Page - radius; i <= p.Page+radius; i++ {
		if i >= 1 && i <= total {
			show[i] = struct{}{}
		}
	}

	items := make([]ListPageItem, 0, len(show)+2)
	prev := 0
	for i := 1; i <= total; i++ {
		if _, ok := show[i]; !ok {
			continue
		}
		if prev != 0 && i > prev+1 {
			items = append(items, ListPageItem{Ellipsis: true})
		}
		items = append(items, ListPageItem{Page: i, Current: i == p.Page})
		prev = i
	}
	return items
}
