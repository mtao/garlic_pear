package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type EntryKind int

const (
	KindFile EntryKind = iota
	KindDirectory
)

type Entry struct {
	CanonicalPath string // relative to the root
	Kind          EntryKind
}

type Walker struct {
	Ignores []string
}

func (w *Walker) doesIgnore(str string) bool {
	return slices.Contains(w.Ignores, str)
}

func (w *Walker) walkFrom(root string, cur string) ([]Entry, error) {
	info, err := os.Stat(cur)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", cur, err)
	}
	var entries []Entry
	var kind EntryKind
	if info.IsDir() {
		kind = KindDirectory

		dirEntries, err := os.ReadDir(cur)
		if err != nil {
			return nil, fmt.Errorf("read dir %s: %w", cur, err)
		}
		for _, dirEntry := range dirEntries {
			if w.doesIgnore(dirEntry.Name()) {
				continue
			}

			childPath := filepath.Join(cur, dirEntry.Name())
			childEntries, err := w.walkFrom(root, childPath)
			if err != nil {
				return nil, err
			}
			entries = append(entries, childEntries...)
		}
	} else {
		kind = KindFile
	}

	relPath, err := filepath.Rel(root, cur)
	if err != nil {
		return nil, fmt.Errorf("relative path from %s to %s: %w", root, cur, err)
	}
	entries = append(entries, Entry{relPath, kind})
	return entries, nil
}

func (w *Walker) WalkTree(root string) ([]Entry, error) {
	return w.walkFrom(root, root)
}
