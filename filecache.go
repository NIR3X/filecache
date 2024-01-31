package filecache

import (
	"bytes"
	"io"
	"os"
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
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > f.maxCacheSize {
		f.toPipe[path] = struct{}{}
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	f.cached[path] = data
	return nil
}

func (f *FileCache) Delete(path string) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	delete(f.cached, path)
	delete(f.toPipe, path)
}

func (f *FileCache) Get(path string) (io.Reader, *io.PipeWriter, error) {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	if _, ok := f.toPipe[path]; ok {
		file, err := os.OpenFile(path, os.O_RDONLY, 0666)
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
	if data, ok := f.cached[path]; ok {
		return bytes.NewReader(data), nil, nil
	}
	return nil, nil, os.ErrNotExist
}
