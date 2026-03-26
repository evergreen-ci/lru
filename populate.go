package lru

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// DirectoryContents takes a path and builds a cache object. If
// skipDir is true, this option does not include any directories,
// otherwise all directories are included in the cache. When including
// directories in the cache, lru includes the aggregate size of files
// in the directory.
func DirectoryContents(path string, skipDir bool) (*Cache, error) {
	ctx := context.Background()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "getting absolute path for path '%s'", path)
	}
	path = absPath

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "getting directory contents for path '%s'", path)
	}

	c := NewCache()
	catcher := grip.NewCatcher()

	for _, info := range infos {
		if info.IsDir() && skipDir {
			continue
		}

		fn := filepath.Join(path, info.Name())

		catcher.Add(c.AddStat(fn, info))
	}

	if catcher.HasErrors() {
		return nil, errors.Wrapf(err, "building cache with %d items (of %d)", catcher.Len(), c.Count())
	}

	grip.Debugf(ctx, "created new cache, with %d items and %d bytes", c.Count(), c.Size())

	return c, nil
}

// TreeContents adds all file system items, excluding directories, to
// a cache object.
func TreeContents(root string) (*Cache, error) {
	ctx := context.Background()

	absPath, err := filepath.Abs(root)
	if err != nil {
		return nil, errors.Wrapf(err, "getting absolute path for path '%s'", root)
	}
	root = absPath

	c := NewCache()
	catcher := grip.NewCatcher()
	catcher.Add(filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fn := filepath.Join(path, info.Name())
		catcher.Add(c.AddStat(fn, info))

		return nil
	}))

	if catcher.HasErrors() {
		return nil, errors.Wrapf(err, "building cache with %d items (of %d)", catcher.Len(), c.Count())
	}

	grip.Debugf(ctx, "created new cache, with %d items and %d bytes", c.Count(), c.Size())

	return c, nil
}
