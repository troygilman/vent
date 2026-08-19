package vent

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"time"
)

//go:embed static
var static embed.FS

const immutableCacheControl = "public, max-age=31536000, immutable"

// StaticAssetVersion is a short content hash of embedded static files.
// Append it to asset URLs so browsers fetch new files after deploys
// instead of keeping a stale immutable cache.
var StaticAssetVersion = staticAssetVersion()

var versionedStyleCSS = loadVersionedStyleCSS()

func staticAssetVersion() string {
	h := sha256.New()
	_ = fs.WalkDir(static, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := static.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path))
		h.Write(b)
		return nil
	})
	return hex.EncodeToString(h.Sum(nil)[:8])
}

func loadVersionedStyleCSS() []byte {
	b, err := static.ReadFile("static/css/style.css")
	if err != nil {
		return nil
	}
	return versionCSSURLs(b, StaticAssetVersion)
}

// versionCSSURLs appends the content-hash query to relative url() refs so
// fonts and images referenced from CSS share the same cache-busting scheme
// as HTML-linked assets.
func versionCSSURLs(css []byte, version string) []byte {
	for _, path := range []string{
		"../fonts/inter-latin.woff2",
		"../img/vent-logo.png",
	} {
		old := []byte(`url("` + path + `")`)
		neu := []byte(`url("` + path + `?v=` + version + `")`)
		css = bytes.ReplaceAll(css, old, neu)
	}
	return css
}

// StaticDirHandler returns an HTTP handler that serves static files embedded in the vent package.
func StaticDirHandler() http.Handler {
	fileServer := http.FileServerFS(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/css/style.css" && len(versionedStyleCSS) > 0 {
			serveVersionedCSS(w, r, versionedStyleCSS)
			return
		}
		fileServer.ServeHTTP(&cacheControlWriter{ResponseWriter: w}, r)
	})
}

func serveVersionedCSS(w http.ResponseWriter, r *http.Request, css []byte) {
	w.Header().Set("Cache-Control", immutableCacheControl)
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	http.ServeContent(w, r, "style.css", time.Time{}, bytes.NewReader(css))
}

// cacheControlWriter applies a long-lived immutable cache only to successful
// responses so 404s are not stored for a year.
type cacheControlWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *cacheControlWriter) WriteHeader(code int) {
	if !w.wrote {
		w.wrote = true
		if code == http.StatusOK {
			w.Header().Set("Cache-Control", immutableCacheControl)
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *cacheControlWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *cacheControlWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
