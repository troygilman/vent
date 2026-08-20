package vent

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
)

//go:embed static
var static embed.FS

// StaticAssetVersion is a short content hash of embedded static files.
// Append it to asset URLs so browsers fetch new files after deploys
// instead of keeping a stale immutable cache.
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
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(w, r)
	})
}
