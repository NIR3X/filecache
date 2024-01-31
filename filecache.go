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

func (fc *FileCache) Update(path string) error {
	fc.mtx.Lock()
	defer fc.mtx.Unlock()
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > fc.maxCacheSize {
		fc.toPipe[path] = struct{}{}
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fc.cached[path] = data
	return nil
}

func (fc *FileCache) Delete(path string) {
	fc.mtx.Lock()
	defer fc.mtx.Unlock()
	delete(fc.cached, path)
	delete(fc.toPipe, path)
}

func (fc *FileCache) Get(path string) (io.Reader, *io.PipeWriter, error) {
	fc.mtx.RLock()
	defer fc.mtx.RUnlock()
	if _, ok := fc.toPipe[path]; ok {
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
	if data, ok := fc.cached[path]; ok {
		return bytes.NewReader(data), nil, nil
	}
	return nil, nil, os.ErrNotExist
}
