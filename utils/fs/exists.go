package fs

import (
	"github.com/patrickmn/go-cache"
	"io/fs"
	"net/http"
	"strings"
)

func NewExistsFS(fs fs.FS) ExistsFS {
	return ExistsFS{
		FS:    fs,
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

type ExistsFS struct {
	fs.FS
	UseCache bool
	cache    *cache.Cache
}

func (efs ExistsFS) WithCache(withCache bool) ExistsFS {
	efs.UseCache = withCache
	return efs
}

func (efs ExistsFS) Exists(name string) bool {
	if efs.UseCache {
		if result, ok := efs.cache.Get(name); ok {
			return result.(bool)
		}
	}
	_, err := fs.Stat(efs.FS, name)
	result := err == nil
	if efs.UseCache {
		efs.cache.SetDefault(name, result)
	}
	return result
}

func (efs ExistsFS) Open(name string) (fs.File, error) {
	return efs.FS.Open(name)
}

// ---

type ExistsHttpFS struct {
	ExistsFS
	httpFs http.FileSystem
}

func NewExistsHttpFS(fs ExistsFS) ExistsHttpFS {
	return ExistsHttpFS{
		ExistsFS: fs,
		httpFs:   http.FS(fs),
	}
}

func (ehfs ExistsHttpFS) Exists(name string) bool {
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	return ehfs.ExistsFS.Exists(name)
}

func (ehfs ExistsHttpFS) Open(name string) (http.File, error) {
	return ehfs.httpFs.Open(name)
}
