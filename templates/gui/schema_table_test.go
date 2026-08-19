package gui

import (
	"bytes"
	"context"
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
	if !strings.Contains(html, `data-cookie:widgets="`) {
		t.Fatal("filter form should persist widget drawer signals to a cookie")
	}
	if !strings.Contains(html, `{include: /^widgets\./, path: &#34;/admin/&#34;}`) &&
		!strings.Contains(html, `{include: /^widgets\./, path: "/admin/"}`) {
		t.Fatal("widget cookie should include widgets.* and the admin path")
	}
}
