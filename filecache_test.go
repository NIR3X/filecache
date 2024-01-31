package filecache

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"testing"
)

func TestFileCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmplreload")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	file1, err := os.CreateTemp(tmpDir, "file1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1.Name())
	file2, err := os.CreateTemp(tmpDir, "file2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2.Name())
	_, err = file1.Write(make([]byte, 1024*1024))
	if err != nil {
		t.Fatal(err)
	}
	_, err = file2.Write(make([]byte, 3*1024*1024))
	if err != nil {
		t.Fatal(err)
	}
	fc := NewFileCache(2 * 1024 * 1024)
	err = fc.Update(file1.Name())
	if err != nil {
		t.Fatal(err)
	}
	err = fc.Update(file2.Name())
	if err != nil {
		t.Fatal(err)
	}
	r, w, err := fc.Get(file1.Name())
	if err != nil {
		t.Fatal(err)
	}
	if w != nil {
		t.Fatalf("expected no writer, got %v", w)
	}
	buf := make([]byte, 1024*1024)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1024*1024 {
		t.Fatalf("expected to read 1MB, read %d bytes", n)
	}
	r, w, err = fc.Get(file2.Name())
	if err != nil {
		t.Fatal(err)
	}
	if w == nil {
		t.Fatalf("expected a writer, got nil")
	}
	defer w.Close()
	buf = make([]byte, 3*1024*1024)
	n, err = rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3*1024*1024 {
		t.Fatalf("expected to read 3MB, read %d bytes", n)
	}
	buf2 := make([]byte, 3*1024*1024)
	n, err = rand.Read(buf2)
	if err != nil {
		t.Fatal(err)
	}
	written, err := io.Copy(io.MultiWriter(bytes.NewBuffer(buf[:0]), bytes.NewBuffer(buf2[:0])), r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, buf2) {
		t.Fatalf("expected to read the same data twice, got different data")
	}
	if written != 3*1024*1024 {
		t.Fatalf("expected to read 3MB, read %d bytes", n)
	}
}
