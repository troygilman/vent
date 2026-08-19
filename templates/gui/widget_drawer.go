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

const widgetFilterName = "filter"

type widgetDrawerState struct {
	Open   bool   `json:"_open"`
	Active string `json:"active"`
}

type widgetDrawerSignals struct {
	Widgets widgetDrawerState `json:"widgets"`
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
	var payload widgetDrawerSignals
	if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
		return widgetDrawerState{}
	}
	state := widgetDrawerState{
		Open: payload.Widgets.Open,
	}
	if allowedWidgetName(payload.Widgets.Active) {
		state.Active = payload.Widgets.Active
	} else if state.Open {
		state.Active = widgetFilterName
	}
	return state
}

func allowedWidgetName(name string) bool {
	switch name {
	case widgetFilterName:
		return true
	default:
		return false
	}
}

func tableWidgetsFilterActive(state widgetDrawerState) bool {
	return state.Open && strings.EqualFold(state.Active, widgetFilterName)
}
