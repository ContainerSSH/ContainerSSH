package oauth2

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

//go:embed webroot/*
var assets embed.FS

// GetFilesystem creates a scoped filesystem so we resolve file names relative to the webroot directory.
func GetFilesystem() http.FileSystem {
	return http.FS(&scopedFilesystem{
		backend: assets,
		scope:   "webroot/",
	})
}

type scopedFilesystem struct {
	backend embed.FS
	scope   string
}

func (s scopedFilesystem) Open(name string) (fs.File, error) {
	return s.backend.Open(path.Join(s.scope, name))
}
