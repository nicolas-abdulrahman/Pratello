package common

import (
	"io"
	"io/fs"
	"os"
)

// ImageFiles serves the configured directory with OS-enforced root confinement,
// including when a path passes through a symbolic link.
func ImageFiles(dir string) *Images {
	return &Images{open: func(name string) (fs.File, error) {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return nil, err
		}
		defer root.Close()
		return openRegular(root.FS(), name)
	}}
}

// ImageFS serves an already scoped filesystem, such as embedded fixture media.
func ImageFS(files fs.FS) *Images {
	return &Images{open: func(name string) (fs.File, error) { return openRegular(files, name) }}
}

func openRegular(files fs.FS, name string) (fs.File, error) {
	info, err := fs.Stat(files, name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fs.ErrNotExist
	}
	return files.Open(name)
}

type seekableFile interface {
	fs.File
	io.Seeker
}
