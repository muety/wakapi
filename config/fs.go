package config

import (
	"io/fs"
	"os"
)

// ChooseFS returns a local (DirFS) file system when on 'dev' environment and the given go-embed file system otherwise
func ChooseFS(localDir string, embeddedFS fs.FS) fs.FS {
	if Get().IsDev() {
		if _, err := os.Stat(localDir); err == nil {
			return os.DirFS(localDir)
		}
		Log().Warn("attempted to use local fs for directory in dev mode, but failed", "path", localDir)
	}
	return embeddedFS
}
