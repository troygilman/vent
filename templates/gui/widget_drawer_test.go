package gui

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/troygilman/vent/requestctx"
)

func TestParseWidgetDrawerCookie(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		open   bool
		active string
	}{
		{name: "empty", raw: ""},
		{name: "invalid", raw: "not-json"},
		{
			name:   "open filter",
			raw:    `{"widgets":{"_open":true,"active":"filter"}}`,
			open:   true,
			active: "filter",
		},
		{
			name:   "urlencoded",
			raw:    url.QueryEscape(`{"widgets":{"_open":true,"active":"filter"}}`),
			open:   true,
			active: "filter",
		},
		{
			name:   "closed",
			raw:    `{"widgets":{"_open":false,"active":"filter"}}`,
			open:   false,
			active: "filter",
		},
		{
			name:   "open without active defaults to filter",
			raw:    `{"widgets":{"_open":true,"active":""}}`,
			open:   true,
			active: "filter",
		},
		{
			name: "rejects unsafe active",
			raw:  `{"widgets":{"_open":true,"active":"<script>"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWidgetDrawerCookie(tt.raw)
			if got.Open != tt.open || got.Active != tt.active {
				t.Fatalf("parseWidgetDrawerCookie(%q) = %+v, want open=%v active=%q", tt.raw, got, tt.open, tt.active)
			}
		})
	}
}

func TestTableWidgetsSignals(t *testing.T) {
	got := tableWidgetsSignals(widgetDrawerState{Open: true, Active: "filter"})
	want := `{"widgets":{"_open":true,"active":"filter"}}`
	if got != want {
		t.Fatalf("signals = %q, want %q", got, want)
	}
}

func TestTableWidgetsStateReadsRequestCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/users/", nil)
	req.AddCookie(&http.Cookie{
		Name:  WidgetDrawerCookieName,
		Value: url.QueryEscape(`{"widgets":{"_open":true,"active":"filter"}}`),
	})
	ctx := requestctx.WithHTTPRequest(requestctx.WithAdminPath(t.Context(), "/admin/"), req)

	got := tableWidgetsState(ctx)
	if !got.Open || got.Active != "filter" {
		t.Fatalf("state = %+v, want open filter", got)
	}
}

func TestTableWidgetsStateWithoutRequest(t *testing.T) {
	got := tableWidgetsState(t.Context())
	if got.Open || got.Active != "" {
		t.Fatalf("state = %+v, want closed empty", got)
	}
}
