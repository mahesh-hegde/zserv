package main

import (
	"io/fs"
)

var excludedFiles = map[string]bool{
	"wget.log":  true,
	"nohup.out": true,
}

const traverseLimit = 5

// detectRoot attempts to find possible website root by traversing down the
// filesystem till there is single directory entry.
//
// It's a very conservative heuristic, it may not be able to determine the root
// successfully, in which case it returns nil (if there's an error), or the
// parent filesystem itself.
func detectRoot(fsys fs.FS) (fs.FS, error) {
	depth := 0
	for depth < traverseLimit {
		sub, err := getSingleChild(fsys)
		if err != nil {
			return nil, err
		}
		if sub == nil {
			// cannot go further down the filesystem
			return fsys, nil
		}
		fsys = sub
		depth++
	}
	return fsys, nil
}

// getSingleChild returns the single directory child of dir if applicable, or
// nil if there is no single directory child.
//
// both fs and error can be nil
func getSingleChild(fsys fs.FS) (fs.FS, error) {
	children, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	count := 0
	var last fs.DirEntry = nil
	for _, child := range children {
		if !excludedFiles[child.Name()] {
			count++
			last = child
		}
	}
	if count == 1 && last != nil && last.IsDir() {
		return fs.Sub(fsys, last.Name())
	}
	return nil, nil
}
