package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type entry struct {
	Name, Body string
}

func createZipFile(entries []entry) *bytes.Buffer {
	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)
	for _, e := range entries {
		f, err := w.Create(e.Name)
		checkError(err, "cannot create ZIP file")
		_, err = io.Copy(f, strings.NewReader(e.Body))
		checkError(err, "unable to write content")
	}
	checkError(w.Close(), "cannot close zip writer")
	return &buffer
}

type namedFS interface {
	fs.FS
	Name() string
}

func createBothTestZipReaderFS(entries []entry) []namedFS {
	zip := createZipFile(entries)
	return []namedFS{
		GetBufferingZipFS(bytes.NewReader(zip.Bytes()),
			int64(zip.Len()), largeBufferSize),
		GetStreamingZipFs(bytes.NewReader(zip.Bytes()), int64(zip.Len())),
	}
}

const largeBufferSize = 256 * 1024 * 1024 * 1024

// Test creating a zip file, opening a zip reader FS and reading some files
func TestReadZipEntries(t *testing.T) {
	entries := []entry{
		{"Hello", "World"},
		{"Tom", "Jerry"},
		{"Boom/Boom", "Boom/Boom"},
		{"Repeated", strings.Repeat("yep9a8hlnk", 1000)},
	}
	for _, fs := range createBothTestZipReaderFS(entries) {
		t.Run(fs.Name(), func(t *testing.T) {
			for _, e := range entries {
				f, err := fs.Open(e.Name)
				assert.Nil(t, err)
				read, err := io.ReadAll(f)
				assert.Nil(t, err)
				assert.Equal(t, e.Body, string(read))
			}
		})
	}
}

// Test determine HTTP content type from a reader
func TestHttpContentTypeWorks(t *testing.T) {
	repeatedP := strings.Repeat("<p>wguivlasbdswligbv,dsjfwalfdz.wotegul</p>", 1000)
	entries := []entry{
		{"Hello", "World"},
		{"html.txt", "<html><body></body></html>"},
		{"long_html.pdf", fmt.Sprintf("<html><body>%s</body></html>\n", repeatedP)},
		{"binary", strings.Repeat("\u0001\u0002\u0003", 1000)},
	}
	fses := createBothTestZipReaderFS(entries)
	expect := map[string]string{
		"Hello":         "text/plain",
		"html.txt":      "text/html",
		"long_html.pdf": "text/html",
		"binary":        "application/octet-stream",
	}
	for _, fs := range fses {
		t.Run(fs.Name(), func(t *testing.T) {
			for name, ctype := range expect {
				f, err := fs.Open(name)
				assert.Nil(t, err)
				bytes, err := io.ReadAll(f)
				assert.Nil(t, err)
				assert.Contains(t, http.DetectContentType(bytes), ctype,
					"Detect content type for "+name)
			}
		})
	}
}

// Test creating large HTML file and adding it into zip, then detecting content
// type
func TestLargeHtmlFile(t *testing.T) {
	repeatedP := strings.Repeat("<p>ABCDEFGHIJKLMNOPQRSTUVWXYZ</p>", 10000)
	entries := []entry{
		{"large.html", fmt.Sprintf("<html><body>%s</body></html>\n", repeatedP)},
	}
	fses := createBothTestZipReaderFS(entries)
	for _, fs := range fses {
		t.Run(fs.Name(), func(t *testing.T) {
			f, err := fs.Open("large.html")
			assert.Nil(t, err)
			bytes, err := io.ReadAll(f)
			assert.Nil(t, err)
			assert.Contains(t, http.DetectContentType(bytes), "text/html")
		})
	}
}
