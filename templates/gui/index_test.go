package gui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/troygilman/vent"
	"github.com/troygilman/vent/requestctx"
)

func TestIndexVersionsAllStaticAssets(t *testing.T) {
	ctx := requestctx.WithAdminPath(context.Background(), "/admin/")
	ctx = requestctx.WithCSRFToken(ctx, "test-csrf-token")
	ctx = requestctx.WithTheme(ctx, "light")

	var buf bytes.Buffer
	if err := Index().Render(ctx, &buf); err != nil {
		t.Fatalf("render Index: %v", err)
	}
	html := buf.String()
	v := vent.StaticAssetVersion
	for _, path := range []string{
		"static/js/datastar.js",
		"static/js/datastar-plugins.js",
		"static/js/common.js",
		"static/js/command-palette.js",
		"static/css/style.css",
		"static/fonts/inter-latin.woff2",
		"static/img/favicon.ico",
		"static/img/vent-logo.png",
	} {
		want := "/admin/" + path + "?v=" + v
		if !strings.Contains(html, want) {
			t.Errorf("Index HTML missing versioned asset %q", want)
		}
	}
}
