package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
)

type BufferedZipEntry struct {
	fs.File
	*bytes.Reader
}

func (b *BufferedZipEntry) Read(p []byte) (n int, err error) {
	return b.Reader.Read(p)
}

func (b *BufferedZipEntry) ReadDir(n int) ([]fs.DirEntry, error) {
	if rdf, ok := b.File.(fs.ReadDirFile); ok {
		return rdf.ReadDir(n)
	}
	return nil, fmt.Errorf("fs.File instance does not implement ReadDir")
}

func NewBufferedZipEntry(f fs.File) *BufferedZipEntry {
	if f == nil {
		return nil
	}
	var buf bytes.Buffer
	io.Copy(&buf, f)
	return &BufferedZipEntry{f, bytes.NewReader(buf.Bytes())}
}

var _ io.ReadSeeker = &BufferedZipEntry{}
var _ fs.File = &BufferedZipEntry{}

var errBufferSizeExceeded = fmt.Errorf("size exceeds maximum allowed buffer size")

type ZipReaderFS struct {
	*zip.Reader
	maxBufferSize int64
}

func (z ZipReaderFS) Open(name string) (fs.File, error) {
	verbose("open: %s", name)
	f, err := z.Reader.Open(name)
	if err != nil || f == nil {
		verbose("error opening %s: %v", name, err)
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() && stat.Size() >= z.maxBufferSize {
		verbose("Buffer size exceeded for entry: %s, size: %d", name, stat.Size())
		return nil, errBufferSizeExceeded
	}
	return NewBufferedZipEntry(f), nil
}

var _ fs.FS = &ZipReaderFS{}

func OpenZipReaderFS(path string, maxBufferSize int64) fs.FS {
	f, err := os.Open(path)
	checkError(err, "cannot open input file")
	fstat, err := f.Stat()
	checkError(err, "Cannot stat input file")
	return GetZipReaderFS(f, fstat.Size(), maxBufferSize)
}

func GetZipReaderFS(reader io.ReaderAt, size int64,
	maxBufferSize int64) ZipReaderFS {
	zipReader, err := zip.NewReader(reader, size)
	checkError(err, "cannot open input ZIP file")
	return ZipReaderFS{zipReader, maxBufferSize}
}