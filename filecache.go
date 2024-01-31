package filecache

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileCache struct {
	mtx          sync.RWMutex
	maxCacheSize int64
	cached       map[string][]uint8
	toPipe       map[string]struct{}
}

func NewFileCache(maxCacheSize int64) *FileCache {
	return &FileCache{
		mtx:          sync.RWMutex{},
		maxCacheSize: maxCacheSize,
		cached:       make(map[string][]uint8),
		toPipe:       make(map[string]struct{}),
	}
}

func (f *FileCache) Update(path string) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if info.Size() > f.maxCacheSize {
		f.toPipe[absPath] = struct{}{}
		return nil
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	f.cached[absPath] = data
	return nil
}

func (f *FileCache) Delete(path string) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	absPath, err := filepath.Abs(path)
	if err == nil {
		delete(f.cached, absPath)
		delete(f.toPipe, absPath)
	}
}

func (f *FileCache) Get(path string) (io.Reader, *io.PipeWriter, error) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}
	if _, ok := f.toPipe[absPath]; ok {
		file, err := os.OpenFile(absPath, os.O_RDONLY, 0666)
		if err != nil {
			return nil, nil, err
		}
		r, w := io.Pipe()
		go func() {
			io.Copy(w, file)
			w.Close()
		}()
		return r, w, nil
	}
	if data, ok := f.cached[absPath]; ok {
		return bytes.NewReader(data), nil, nil
	}
	return nil, nil, os.ErrNotExist
}

const (
	NotFound = iota
	Cached
	Piped
)

func (f *FileCache) Identify(path string) int {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return NotFound
	}
	if _, ok := f.cached[absPath]; ok {
		return Cached
	}
	if _, ok := f.toPipe[absPath]; ok {
		return Piped
	}
	return NotFound
}

func (f *FileCache) GetCached(path string) (io.Reader, int) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, NotFound
	}
	if data, ok := f.cached[absPath]; ok {
		return bytes.NewReader(data), Cached
	}
	if _, ok := f.toPipe[absPath]; ok {
		return nil, Piped
	}
	return nil, NotFound
}
