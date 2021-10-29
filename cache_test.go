package lru

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CacheSuite struct {
	tempDir string
	cache   *Cache
	require *require.Assertions
	suite.Suite
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}

func (s *CacheSuite) SetupSuite() {
	dir, err := ioutil.TempDir("", uuid.NewV4().String())
	s.Require().NoError(err)
	s.tempDir = dir
}

func (s *CacheSuite) TearDownSuite() {
	s.NoError(os.RemoveAll(s.tempDir))
}

func (s *CacheSuite) SetupTest() {
	s.cache = NewCache()
	s.Require().Len(s.cache.table, 0)
}

func (s *CacheSuite) TestInitialStateOfCacheObjectIsEmpty() {
	s.Equal(0, s.cache.size)
	s.Len(s.cache.table, 0)
	s.Equal(0, s.cache.heap.Len())

	s.Equal(0, s.cache.Count())
	s.Equal(0, s.cache.Size())
}

func (s *CacheSuite) TestAddFileThatDoeNotExistResultsInError() {
	s.Error(s.cache.AddFile(filepath.Join(s.tempDir, "DOES-NOT-EXIST")))
}

func (s *CacheSuite) TestAddDirectoryThatExistsSucceeds() {
	s.NoError(s.cache.AddFile(s.tempDir))
}

func (s *CacheSuite) TestAddRejectsFilesThatAlreadyExist() {
	s.NoError(s.cache.AddFile(s.tempDir))

	for i := 0; i < 40; i++ {
		s.Error(s.cache.AddFile(s.tempDir))
	}
}

func (s *CacheSuite) TestMutlithreadedFileAdds() {
	s.Equal(0, s.cache.Count())
	s.Equal(0, s.cache.Size())

	wg := &sync.WaitGroup{}
	var totalSize int
	for i := 0; i < 40; i++ {
		fn := filepath.Join(s.tempDir, uuid.NewV4().String())
		content := fmt.Sprintf("in %s is it %d", fn, i)
		s.NoError(ioutil.WriteFile(fn, []byte(content), 0644))
		totalSize += len(content)
		wg.Add(1)
		go func(f string) {
			s.NoError(s.cache.AddFile(f))
			wg.Done()
		}(fn)
	}
	wg.Wait()

	s.Equal(40, s.cache.Count())
	s.Equal(totalSize, s.cache.Size())
}

func (s *CacheSuite) TestAddStatRejectsFilesWithNilStats() {
	s.Error(s.cache.AddStat("foo", nil))
}

func (s *CacheSuite) TestUpdateRejectsRecordsThatAreNotLocal() {
	fn := filepath.Join(s.tempDir, "foo")
	fobj := &FileObject{
		Path: fn,
	}

	s.NoError(ioutil.WriteFile(fn, []byte("foo"), 0644))

	s.Error(s.cache.Update(fobj))
	s.NoError(s.cache.Add(fobj))
	s.NoError(s.cache.Update(fobj))
}
