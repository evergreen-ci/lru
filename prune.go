package lru

import (
	"github.com/pkg/errors"
	"github.com/tychoish/grip"
)

// Prune removes files (when dryRun is false), from the file system
// until the total size of the cache is less than the maxSize (in
// bytes.)
func (c *LruCache) Prune(maxSize int, dryRun bool) error {
	catcher := grip.NewCatcher()

	for {
		if c.underQuota(maxSize) {
			break
		}

		if err := c.prunePass(dryRun); err != nil {
			grip.Noticef("cache pruning ended early due to error, (size=%d, count=%d)",
				c.Size(), c.Count())
			catcher.Add(err)
		}
	}

	grip.Infof("cache pruning is complete; cache is %d bytes with %d items",
		c.Size(), c.Count())

	return catcher.Resolve()
}

func (c *LruCache) underQuota(maxSize int) bool {
	if c.Count() == 0 {
		grip.Info("there are no items in the cache")
		return true
	}

	size := c.Size()
	if size <= maxSize {
		grip.Infof("cache size %d is under target size of %d",
			size, maxSize)
		return true
	}

	return false
}

func (c *LruCache) prunePass(dryRun bool) error {
	f, err := c.Pop()
	if err != nil {
		return errors.Wrap(err, "problem retrieving item from cache")
	}

	if dryRun {
		grip.Noticef("[dry-run]: would delete '%s'", f.Path)
		return nil
	}

	if err := f.Remove(); err != nil {
		return errors.Wrap(err, "problem removing item")
	}

	return nil
}
