package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	oldTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
)

// fixture describes one file to create under a test root.
type fixture struct {
	content string
	mode    os.FileMode // zero leaves the default
	mtime   time.Time   // zero means oldTime
}

func writeFile(t *testing.T, root, rel, content string) string {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// materialize creates every fixture under root. Parent directories are
// created implicitly, so an entry like "a/b/c.txt" yields directories a and a/b.
func materialize(t *testing.T, root string, files map[string]fixture) {
	t.Helper()
	for rel, f := range files {
		p := writeFile(t, root, rel, f.content)
		if f.mode != 0 {
			if err := os.Chmod(p, f.mode); err != nil {
				t.Fatal(err)
			}
		}
		mt := f.mtime
		if mt.IsZero() {
			mt = oldTime
		}
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
}
