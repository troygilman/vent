package vent

import (
	"embed"
	"net/http"
)

//go:embed static
var static embed.FS

// StaticDirHandler returns an HTTP handler that serves static files embedded in the vent package.
func StaticDirHandler() http.Handler {
	return http.FileServerFS(static)
}
