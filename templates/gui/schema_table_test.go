package gui

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/troygilman/vent"
	"github.com/troygilman/vent/requestctx"
)

func TestTableFiltersActive(t *testing.T) {
	if tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: ""},
		{Name: "is_staff", Type: "bool", Value: vent.BoolFilterAll.String()},
	}) {
		t.Fatal("empty filters should be inactive")
	}
	if !tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "email", Type: "string", Value: "admin"},
	}) {
		t.Fatal("string value should be active")
	}
	if !tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "is_staff", Type: "bool", Value: vent.BoolFilterTrue.String()},
	}) {
		t.Fatal("bool true should be active")
	}
	if tableFiltersActive([]SchemaTableFilterableColumn{
		{Name: "is_staff", Type: "bool", Value: ""},
	}) {
		t.Fatal("empty bool should be inactive")
	}
}

func TestTableFilterActiveCount(t *testing.T) {
	if got := tableFilterActiveCount([]SchemaTableFilterableColumn{
		{Name: "title", Type: "string", Value: "Dune"},
		{Name: "published", Type: "bool", Value: vent.BoolFilterTrue.String()},
		{Name: "pages", Type: "int", Value: ""},
	}); got != 2 {
		t.Fatalf("active count = %d, want 2", got)
	}
}

func TestTableFilterChipValue(t *testing.T) {
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "string", Value: "Dune",
	}); got != "Dune" {
		t.Fatalf("string chip = %q, want Dune", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "int", Value: "412",
	}); got != "412" {
		t.Fatalf("int chip = %q, want 412", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterTrue.String(),
	}); got != "Yes" {
		t.Fatalf("bool true chip = %q, want Yes", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterFalse.String(),
	}); got != "No" {
		t.Fatalf("bool false chip = %q, want No", got)
	}
	if got := tableFilterChipValue(SchemaTableFilterableColumn{
		Type: "bool", Value: vent.BoolFilterAll.String(),
	}); got != "" {
		t.Fatalf("bool all chip = %q, want empty", got)
	}
}

func TestTableFilterClearClick(t *testing.T) {
	got := tableFilterClearClick("title")
	want := `tableFilters.dispatch(el, {filterNames: ["title"]})`
	if got != want {
		t.Fatalf("clear click = %q, want %q", got, want)
	}
}

func TestTableFilterResetAllClick(t *testing.T) {
	got := tableFilterResetAllClick([]SchemaTableFilterableColumn{
		{Name: "email"},
		{Name: "is_staff"},
	})
	want := `tableFilters.dispatch(el, {filterNames: ["email", "is_staff"]})`
	if got != want {
		t.Fatalf("reset all click = %q, want %q", got, want)
	}
}

func TestTableWidgetsCookieExpr(t *testing.T) {
	got := tableWidgetsCookieExpr("/admin/")
	want := `{include: /^widgets\./, path: "/admin/"}`
	if got != want {
		t.Fatalf("widget cookie expr = %q, want %q", got, want)
	}
}

func TestTableColumnKind(t *testing.T) {
	cases := map[string]string{
		"bool":      "bool",
		"int":       "num",
		"int64":     "num",
		"uint":      "num",
		"float64":   "num",
		"time":      "time",
		"time.Time": "time",
		"string":    "text",
		"edge":      "text",
		"custom":    "text",
		"":          "text",
	}
	for colType, want := range cases {
		if got := tableColumnKind(colType); got != want {
			t.Errorf("tableColumnKind(%q) = %q, want %q", colType, got, want)
		}
	}
}

func TestTableColumnWidthPercentSumsTo100(t *testing.T) {
	columns := []SchemaTableColumn{
		{Type: "string"},
		{Type: "bool"},
		{Type: "bool"},
		{Type: "bool"},
		{Type: "time.Time"},
	}
	sum := 0
	for i := range columns {
		var pct int
		got := tableColumnWidthPercent(columns, i)
		if _, err := fmt.Sscanf(got, "%d%%", &pct); err != nil {
			t.Fatalf("column %d width %q: %v", i, got, err)
		}
		if pct <= 0 {
			t.Fatalf("column %d width %q should be positive", i, got)
		}
		sum += pct
	}
	if sum != 100 {
		t.Fatalf("widths sum to %d, want 100", sum)
	}
	email := tableColumnWidthPercent(columns, 0)
	boolCol := tableColumnWidthPercent(columns, 1)
	if email <= boolCol {
		t.Fatalf("text column width %s should exceed bool width %s", email, boolCol)
	}
}

func TestSchemaTableAddButtonIsALink(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "books",
		SingularDisplayName: "Book",
		PluralDisplayName:   "Books",
		FilterableColumns: []SchemaTableFilterableColumn{
			{Name: "title", Type: "string", Value: ""},
		},
		RenderContext: RenderContext{CanCreate: true},
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `<a class="btn btn-primary" href="/admin/books/add/">Add Book</a>`) {
		t.Fatal("add control should be an anchor button, not a nested submit button")
	}
	if strings.Contains(html, `<a href="/admin/books/add/"><button`) {
		t.Fatal("add control must not nest a button inside a link (submits the filter form)")
	}
}

func TestSchemaTableFilterChipDelimiter(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "users",
		SingularDisplayName: "User",
		PluralDisplayName:   "Users",
		FilterableColumns: []SchemaTableFilterableColumn{
			{Name: "email", Label: "Email", Type: "string", Value: "admin"},
			{Name: "is_staff", Label: "IsStaff", Type: "bool", Value: vent.BoolFilterFalse.String()},
		},
		RenderContext: RenderContext{CanCreate: true},
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, "IsStaff: <b>No</b>") {
		t.Fatal("active filter chips should delimit the label and value with \": \"")
	}
	if !strings.Contains(html, `data-on:table-filter-reset="tableFilters.onReset($filter, evt)"`) {
		t.Fatal("filter form should listen for table-filter-reset")
	}
	if strings.Contains(html, "table-filter-reset__document") {
		t.Fatal("filter fields should not listen for table-filter-reset")
	}
	if !strings.Contains(html, `tableFilters.dispatch(el, {filterNames: [&#34;is_staff&#34;]})`) {
		t.Fatal("chip remove should dispatch table-filter-reset with that filter name")
	}
	if !strings.Contains(html, `tableFilters.dispatch(el, {filterNames: [&#34;email&#34;, &#34;is_staff&#34;]})`) {
		t.Fatal("clear should dispatch table-filter-reset with all filter names")
	}
	if strings.Contains(html, "resetAll") {
		t.Fatal("filter resets should not use a resetAll flag")
	}
	if !strings.Contains(html, `data-cookie:vent-widgets="`) {
		t.Fatal("filter form should persist widget drawer signals to a named cookie")
	}
	if !strings.Contains(html, `{include: /^widgets\./, path: &#34;/admin/&#34;}`) &&
		!strings.Contains(html, `{include: /^widgets\./, path: "/admin/"}`) {
		t.Fatal("widget cookie should include widgets.* and the admin path")
	}
	if !strings.Contains(html, `data-signals__ifmissing="{&#34;widgets&#34;:{&#34;_open&#34;:false,&#34;active&#34;:&#34;&#34;}}"`) &&
		!strings.Contains(html, `data-signals__ifmissing='{"widgets":{"_open":false,"active":""}}'`) {
		t.Fatal("missing cookie should initialize the drawer closed")
	}
	if strings.Contains(html, `class="widget-drawer is-open"`) {
		t.Fatal("missing cookie should not render the drawer open")
	}
	if !strings.Contains(html, `id="schema-table-filters"`) || !strings.Contains(html, `data-indicator="_indicator"`) {
		t.Fatal("list fetches should drive the existing page overlay")
	}
	if strings.Contains(html, "No filters available") {
		t.Fatal("schemas with filterable columns should not show the empty filters message")
	}
}

func TestSchemaTableDrawerRendersOpenFromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/users/", nil)
	req.AddCookie(&http.Cookie{
		Name:  WidgetDrawerCookieName,
		Value: url.QueryEscape(`{"widgets":{"_open":true,"active":"filter"}}`),
	})
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")
	ctx = requestctx.WithHTTPRequest(ctx, req)

	props := SchemaTableProps{
		RouteName:           "users",
		SingularDisplayName: "User",
		PluralDisplayName:   "Users",
		FilterableColumns: []SchemaTableFilterableColumn{
			{Name: "email", Label: "Email", Type: "string", Value: ""},
		},
		RenderContext: RenderContext{CanCreate: true},
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, `class="widget-drawer is-open"`) {
		t.Fatal("open cookie should render the drawer with is-open before JS runs")
	}
	if !strings.Contains(html, `data-signals__ifmissing="{&#34;widgets&#34;:{&#34;_open&#34;:true,&#34;active&#34;:&#34;filter&#34;}}"`) &&
		!strings.Contains(html, `"_open":true`) {
		t.Fatal("open cookie should initialize widget signals as open")
	}
	if !strings.Contains(html, `aria-label="Collapse drawer"`) {
		t.Fatal("open cookie should label the toggle as collapse")
	}
	if !strings.Contains(html, `class="widget-drawer-icon is-active"`) {
		t.Fatal("open filter cookie should mark the Filters icon active")
	}
}

func TestSchemaTableRendersFixedColumnGroup(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "users",
		SingularDisplayName: "User",
		PluralDisplayName:   "Users",
		Columns: []SchemaTableColumn{
			{Name: "email", Label: "Email", Type: "string"},
			{Name: "is_staff", Label: "IsStaff", Type: "bool"},
			{Name: "pages", Label: "Pages", Type: "int"},
			{Name: "last_login", Label: "LastLogin", Type: "time.Time"},
			{Name: "author", Label: "Author", Type: "edge"},
		},
		Rows: []SchemaTableRow{{
			Cells: []SchemaTableCell{
				{Display: "admin@vent.com", LinkURL: "/admin/users/1/"},
				{Display: "true"},
				{Display: "412"},
				{Display: "2026-08-20T03:43"},
				{Display: "Frank Herbert"},
			},
		}},
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if strings.Count(html, "<col ") != 5 {
		t.Fatalf("want 5 col elements, html:\n%s", html)
	}
	if strings.Contains(html, `class="data-col-`) {
		t.Fatal("col widths should come from the width attribute, not type classes")
	}
	for _, width := range []string{`width="30%"`, `width="12%"`, `width="10%"`, `width="17%"`, `width="31%"`} {
		if !strings.Contains(html, "<col "+width) {
			t.Fatalf("colgroup missing %s in:\n%s", width, html)
		}
	}
	if !strings.Contains(html, `title="admin@vent.com"`) {
		t.Fatal("truncated cells should expose full value in title")
	}
}

func TestSchemaTableLoadingOmitsNoData(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "users",
		SingularDisplayName: "User",
		PluralDisplayName:   "Users",
		Columns: []SchemaTableColumn{
			{Name: "email", Label: "Email", Type: "string"},
		},
		FilterableColumns: []SchemaTableFilterableColumn{
			{Name: "email", Label: "Email", Type: "string", Value: ""},
		},
		Loading: true,
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if strings.Contains(html, "No data") {
		t.Fatal("chrome-first paint must not flash No data")
	}
	if !strings.Contains(html, `data-indicator="_indicator"`) {
		t.Fatal("loading list fetch should drive the existing page overlay")
	}
}

func TestSchemaTableLoadingWithoutFiltersFetchesRows(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "notes",
		SingularDisplayName: "Note",
		PluralDisplayName:   "Notes",
		Columns: []SchemaTableColumn{
			{Name: "title", Label: "Title", Type: "string"},
		},
		Loading: true,
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if strings.Contains(html, "No data") {
		t.Fatal("chrome-first paint must not flash No data")
	}
	if strings.Contains(html, `id="schema-table-page"`) {
		t.Fatal("unfiltered tables should use the same list form as filtered tables")
	}
	if !strings.Contains(html, `id="schema-table-filters"`) {
		t.Fatal("unfiltered tables should still wrap the page in the list form")
	}
	if !strings.Contains(html, `class="widget-drawer"`) {
		t.Fatal("unfiltered tables should still render the widget drawer")
	}
	if !strings.Contains(html, `data-on:query-string="@get(location.pathname, {filterSignals: {include: /^filter\./}})"`) {
		t.Fatal("unfiltered tables should lazy-load rows the same way as filtered tables")
	}
	if !strings.Contains(html, `data-indicator="_indicator"`) {
		t.Fatal("unfiltered list fetch should drive the existing page overlay")
	}
	if !strings.Contains(html, `class="widget-drawer-empty"`) || !strings.Contains(html, "No filters available") {
		t.Fatal("Filters panel should say no filters are available")
	}
	if strings.Contains(html, `data-init=`) {
		t.Fatal("unfiltered tables should not use a separate data-init fetch")
	}
}

func TestSchemaTableEmptyShowsNoDataAfterLoad(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "system")

	props := SchemaTableProps{
		RouteName:           "users",
		SingularDisplayName: "User",
		PluralDisplayName:   "Users",
		Columns: []SchemaTableColumn{
			{Name: "email", Label: "Email", Type: "string"},
		},
	}

	var buf bytes.Buffer
	if err := SchemaTablePage(props).Render(ctx, &buf); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, "No data") {
		t.Fatal("loaded empty table should say No data")
	}
}
