package fs

import (
	"io/fs"
	"path/filepath"
)

// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings

type NeuteredFileSystem struct {
	Fs fs.FS
}

func (nfs NeuteredFileSystem) Open(path string) (fs.File, error) {
	f, err := nfs.Fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.Fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
