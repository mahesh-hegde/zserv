package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

const sniffSize = 512

// StreamingZipEntryReader implements fs.File as well as io.ReaderAt and
// io.ReadSeekCloser without actually loading the entire file into the memory.

type StreamingZipEntryReader struct {
	fs.File
	*bytes.Reader
}

func (s *StreamingZipEntryReader) Read(buffer []byte) (int, error) {
	if s.Reader.Len() == 0 {
		return s.File.Read(buffer)
	}
	res, err := s.Reader.Read(buffer)
	if err != nil && errors.Is(err, io.EOF) {
		return res, nil
	}
	return res, err
}

func (s *StreamingZipEntryReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart && offset == 0 {
		return s.Reader.Seek(0, io.SeekStart)
	}
	if whence == io.SeekEnd && offset == 0 {
		stat, err := s.File.Stat()
		if err != nil {
			return -1, err
		}
		return stat.Size(), nil
	}
	return -1, fmt.Errorf("unsupported seek parameters")
}

var _ io.ReadSeekCloser = &StreamingZipEntryReader{}
var _ fs.File = &StreamingZipEntryReader{}

// NewStreamingZipEntryReader returns a lazy reader with limited seek.
// `file` must not be a directory.
func NewStreamingZipEntryReader(file fs.File) *StreamingZipEntryReader {
	var buffer bytes.Buffer
	_, err := io.CopyN(&buffer, file, sniffSize*2)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
	return &StreamingZipEntryReader{
		File:   file,
		Reader: bytes.NewReader(buffer.Bytes()),
	}
}

type StreamingZipFS struct {
	*zip.Reader
}

func GetStreamingZipFs(reader io.ReaderAt, size int64) *StreamingZipFS {
	zipReader, err := zip.NewReader(reader, size)
	checkError(err, "cannot create ZIP reader")
	return &StreamingZipFS{Reader: zipReader}
}

func (zipFS *StreamingZipFS) Open(name string) (fs.File, error) {
	verbose("open: %s", name)
	f, err := zipFS.Reader.Open(name)
	if err != nil || f == nil {
		verbose("error opening %s: %v", name, err)
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return f, nil
	}
	return NewBufferedZipEntry(f), nil
}

var _ fs.FS = &ZipReaderFS{}
