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
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	f.mtx.RLock()
	maxCacheSize := f.maxCacheSize
	f.mtx.RUnlock()
	if info.Size() > maxCacheSize {
		f.mtx.Lock()
		f.toPipe[absPath] = struct{}{}
		f.mtx.Unlock()
		return nil
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	f.mtx.Lock()
	f.cached[absPath] = data
	f.mtx.Unlock()
	return nil
}

func (f *FileCache) Delete(path string) {
	absPath, err := filepath.Abs(path)
	if err == nil {
		f.mtx.Lock()
		delete(f.cached, absPath)
		delete(f.toPipe, absPath)
		f.mtx.Unlock()
	}
}

func (f *FileCache) Get(path string) (io.Reader, *io.PipeWriter, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}
	f.mtx.RLock()
	_, ok := f.toPipe[absPath]
	f.mtx.RUnlock()
	if ok {
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
	f.mtx.RLock()
	data, ok := f.cached[absPath]
	f.mtx.RUnlock()
	if ok {
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
	absPath, err := filepath.Abs(path)
	if err != nil {
		return NotFound
	}
	f.mtx.RLock()
	_, ok := f.cached[absPath]
	f.mtx.RUnlock()
	if ok {
		return Cached
	}
	f.mtx.RLock()
	_, ok = f.toPipe[absPath]
	f.mtx.RUnlock()
	if ok {
		return Piped
	}
	return NotFound
}

func (f *FileCache) GetCached(path string) (io.Reader, int) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, NotFound
	}
	f.mtx.RLock()
	data, ok := f.cached[absPath]
	f.mtx.RUnlock()
	if ok {
		return bytes.NewReader(data), Cached
	}
	f.mtx.RLock()
	_, ok = f.toPipe[absPath]
	f.mtx.RUnlock()
	if ok {
		return nil, Piped
	}
	return nil, NotFound
}
