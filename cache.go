// Package lru provides a tool to prune files from a cache based on
// LRU.
//
// lru implements a cache structure that tracks the size and
// (modified*) time of a file.
//
// * future versions of lru may use different time method to better
// approximate usage time rather than modification time.
package lru

import (
	"container/heap"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// LruCache provides tools to maintain an cache of file system
// objects, maintained on a least-recently-used basis.
type LruCache struct {
	size  int
	heap  *fileObjectHeap
	mutex sync.Mutex
	table map[string]*FileObject
}

// NewCache returns an initalized but unpopulated cache. Use the
// DirectoryContents and TreeContents constructors to populate a
// cache.
func NewCache() *LruCache {
	return &LruCache{
		table: make(map[string]*FileObject),
	}
}

// Size returns the total size of objects in the cache.
func (c *LruCache) Size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.size
}

// Count returns the total number of objects in the cache.
func (c *LruCache) Count() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return len(*c.heap)
}

func (c *LruCache) AddStat(fn string, stat os.FileInfo) error {
	if stat == nil {
		return errors.Errorf("file %s does not have a valid stat", fn)
	}

	f := &FileObject{
		Path: fn,
		Size: int(stat.Size()),
		Time: stat.ModTime(),
	}

	if stat.IsDir() {
		size, err := dirSize(fn)
		if err != nil {
			return errors.Wrapf(err, "problem finding size of directory %d", fn)
		}

		f.Size = int(size)
	}

	return errors.Wrapf(c.Add(f), "problem adding file (%s) by info", fn)
}

func (c *LruCache) AddFile(fn string) error {
	stat, err := os.Stat(fn)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "file %s does not exist", fn)
	}

	return errors.Wrap(c.AddStat(fn, stat), "problem adding file")
}

func (c *LruCache) Add(f *FileObject) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.table[f.Path]; ok {
		return errors.Errorf("cannot add object '%s' to cache: it already exists: %s",
			f.Path, "use Update() instead")
	}

	c.size += f.Size
	c.table[f.Path] = f
	c.heap.Push(f)

	return nil
}

func (c *LruCache) Update(f *FileObject) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	existing, ok := c.table[f.Path]
	if !ok {
		return errors.Errorf("cannot update '%s' in cache: it does not exist: %s",
			f.Path, "use Add() instead")
	}

	c.size -= existing.Size
	c.size += f.Size

	f.index = existing.index
	c.table[f.Path] = f
	heap.Fix(c.heap, f.index)

	return nil
}

func (c *LruCache) Pop() (*FileObject, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.heap.Len() == 0 {
		return nil, errors.New("cache listing is empty")
	}

	f := c.heap.Pop().(*FileObject)
	c.size -= f.Size
	delete(c.table, f.Path)

	return f, nil
}

func (c *LruCache) Get(path string) (*FileObject, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	f, ok := c.table[path]
	if !ok {
		return nil, errors.Errorf("file '%s' does not exist in cache", path)
	}

	return f, nil
}
