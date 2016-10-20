package lru

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type FileObject struct {
	Size  int
	Time  time.Time
	Path  string
	index int
}

func NewFile(fn string, info os.FileInfo) *FileObject {
	return &FileObject{
		Path: fn,
		Time: info.ModTime(),
		Size: int(stat.Size()),
	}
}

func (f *FileObject) Update() error {
	stat, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		return errors.Errorf("file %s no longer exists", f.Path)
	}

	f.Time = stat.ModTime()

	if stat.IsDir() {
		size, err := dirSize(f.Path)
		if err != nil {
			return errors.Wrapf(err, "problem finding size of directory %d", fn)
		}

		f.Size = int(size)
	} else {
		f.Size = int(stat.Size())
	}

	return nil
}

func (f *FileObject) Remove() error {
	return os.RemoveAll(f.Path)
}

func dirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, errors.Wrapf(err, "problem getting size of %s", path)
	}

	return size, nil
}
