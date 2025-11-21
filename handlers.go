package vent

import (
	"embed"
	"net/http"
)

//go:embed static
var static embed.FS

func StaticDirHandler() http.Handler {
	return http.FileServerFS(static)
}
