package fs

import (
	"io/fs"
	"net/http"
	"strings"
)

type ExistsFS struct {
	Fs fs.FS
}

func (efs ExistsFS) Exists(name string) bool {
	_, err := fs.Stat(efs.Fs, name)
	return err == nil
}

func (efs ExistsFS) Open(name string) (fs.File, error) {
	return efs.Fs.Open(name)
}

// ---

type ExistsHttpFS struct {
	Fs     ExistsFS
	httpFs http.FileSystem
}

func NewExistsHttpFs(fs ExistsFS) ExistsHttpFS {
	return ExistsHttpFS{
		Fs:     fs,
		httpFs: http.FS(fs),
	}
}

func (ehfs ExistsHttpFS) Exists(name string) bool {
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	return ehfs.Fs.Exists(name)
}

func (ehfs ExistsHttpFS) Open(name string) (http.File, error) {
	return ehfs.httpFs.Open(name)
}
