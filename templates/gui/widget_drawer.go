package gui

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/troygilman/vent/requestctx"
)

// WidgetDrawerCookieName is the cookie written by data-cookie:vent-widgets.
const WidgetDrawerCookieName = "vent-widgets"

type widgetDrawerState struct {
	Open   bool
	Active string
}

func tableWidgetsState(ctx context.Context) widgetDrawerState {
	r := requestctx.HTTPRequest(ctx)
	if r == nil {
		return widgetDrawerState{}
	}
	cookie, err := r.Cookie(WidgetDrawerCookieName)
	if err != nil || cookie.Value == "" {
		return widgetDrawerState{}
	}
	return parseWidgetDrawerCookie(cookie.Value)
}

func parseWidgetDrawerCookie(raw string) widgetDrawerState {
	if raw == "" {
		return widgetDrawerState{}
	}
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		decoded = raw
	}
	var payload struct {
		Widgets struct {
			Open   bool   `json:"_open"`
			Active string `json:"active"`
		} `json:"widgets"`
	}
	if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
		return widgetDrawerState{}
	}
	state := widgetDrawerState{
		Open: payload.Widgets.Open,
	}
	if payload.Widgets.Active == "" {
		if state.Open {
			state.Active = "filter"
		}
		return state
	}
	state.Active = sanitizeWidgetName(payload.Widgets.Active)
	if state.Active == "" {
		return widgetDrawerState{}
	}
	return state
}

func tableWidgetsSignals(state widgetDrawerState) string {
	payload := struct {
		Widgets struct {
			Open   bool   `json:"_open"`
			Active string `json:"active"`
		} `json:"widgets"`
	}{}
	payload.Widgets.Open = state.Open
	payload.Widgets.Active = state.Active
	b, err := json.Marshal(payload)
	if err != nil {
		return `{"widgets":{"_open":false,"active":""}}`
	}
	return string(b)
}

func sanitizeWidgetName(name string) string {
	if name == "" {
		return ""
	}
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return ""
	}
	return name
}

func tableWidgetsToggleLabel(open bool) string {
	if open {
		return "Collapse drawer"
	}
	return "Expand drawer"
}

func tableWidgetsAriaExpanded(open bool) string {
	if open {
		return "true"
	}
	return "false"
}

func tableWidgetsFilterActive(state widgetDrawerState) bool {
	return state.Open && strings.EqualFold(state.Active, "filter")
}
