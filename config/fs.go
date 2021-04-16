package config

import (
	"io/fs"
	"os"
)

// ChooseFS returns a local (DirFS) file system when on 'dev' environment and the given go-embed file system otherwise
func ChooseFS(localDir string, embeddedFS fs.FS) fs.FS {
	if Get().IsDev() {
		return os.DirFS(localDir)
	}
	return embeddedFS
}
