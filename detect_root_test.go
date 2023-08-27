package main

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertDirHasEntries(t *testing.T, fsys fs.FS, names []string) {
	assert.NotNil(t, fsys, "Filesystem is nil")
	found := map[string]bool{}
	for _, name := range names {
		found[name] = false
	}
	count := 0
	entries, err := fs.ReadDir(fsys, ".")
	assert.Nil(t, err)
	for _, entry := range entries {
		entryName := entry.Name()
		if seen, ok := found[entryName]; ok {
			if !seen {
				found[entryName] = true
				count++
			}
		} else {
			t.Errorf("Unknown entry %s", entryName)
		}
	}
	if count != len(names) {
		t.Errorf("Not all names were found in the directory: %v", found)
	}
}

func assertDetectedRootWillHaveEntries(t *testing.T, zipEntries []entry, expected []string) {
	fses := createBothTestZipReaderFS(zipEntries)
	for _, fs := range fses {
		t.Run(fs.Name(), func(t *testing.T) {
			root, err := detectRoot(fs)
			assert.Nil(t, err)
			assertDirHasEntries(t, root, expected)
		})
	}
}

func TestDetectRootInNormalCase(t *testing.T) {
	entries := []entry{
		{"www.website.com/abc/1.html", ""},
		{"www.website.com/abc/2.html", ""},
		{"www.website.com/def/1.html", ""},
		{"www.website.com/def/2.html", ""},
		{"www.website.com/index.html", ""},
	}

	assertDetectedRootWillHaveEntries(t, entries,
		[]string{"abc", "def", "index.html"})
}

func TestDetectRootInBaseCase(t *testing.T) {
	entries := []entry{
		{"abc/1.html", ""},
		{"def/1.html", ""},
		{"def/2.html", ""},
		{"index.html", ""},
		{"abc/2.html", ""},
	}
	assertDetectedRootWillHaveEntries(t, entries,
		[]string{"abc", "def", "index.html"})
}

func TestDetectRootWithExcludedFiles(t *testing.T) {
	entries := []entry{
		{"def/1.html", ""},
		{"def/index.html", ""},
		{"wget.log", ""},
		{"nohup.out", ""},
	}
	assertDetectedRootWillHaveEntries(t, entries,
		[]string{"1.html", "index.html"})
}

// TODO: Test more edge cases
