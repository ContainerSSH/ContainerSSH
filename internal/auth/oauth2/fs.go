package oauth2

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed webroot/*
var assets embed.FS

// GetFilesystem creates a scoped filesystem so we resolve file names relative to the webroot directory.
func GetFilesystem() http.FileSystem {
	subFS, err := fs.Sub(assets, "webroot")
	if err != nil {
		panic(err)
	}
	return http.FS(subFS)
}
