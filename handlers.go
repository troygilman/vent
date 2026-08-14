package vent

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static
var static embed.FS

// StaticAssetVersion is a short content hash of embedded static files.
// Append it to JS/CSS URLs so browsers fetch new assets after deploys
// instead of keeping a stale hour-long cache.
var StaticAssetVersion = staticAssetVersion()

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

// StaticDirHandler returns an HTTP handler that serves static files embedded in the vent package.
func StaticDirHandler() http.Handler {
	fileServer := http.FileServerFS(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		fileServer.ServeHTTP(w, r)
	})
}
