package main

import (
	"errors"
	"io/fs"
	"maps"
	"path/filepath"
	"testing"
)

// kinds flattens a walk result into path -> kind so order doesn't matter.
func kinds(entries []Entry) map[string]EntryKind {
	m := make(map[string]EntryKind, len(entries))
	for _, e := range entries {
		m[e.CanonicalPath] = e.Kind
	}
	return m
}

func TestWalkTree(t *testing.T) {
	tests := []struct {
		name    string
		ignores []string
		files   map[string]fixture
		want    map[string]EntryKind
	}{
		{
			name:  "nested paths are relative to the root",
			files: map[string]fixture{"a.txt": {}, "d/b.txt": {}, "d/e/c.txt": {}},
			want: map[string]EntryKind{
				".":         KindDirectory,
				"a.txt":     KindFile,
				"d":         KindDirectory,
				"d/b.txt":   KindFile,
				"d/e":       KindDirectory,
				"d/e/c.txt": KindFile,
			},
		},
		{
			name:    "ignored directory is pruned entirely",
			ignores: DefaultIgnores,
			files:   map[string]fixture{"keep.txt": {}, ".git/HEAD": {}, ".git/hooks/pre-commit": {}},
			want: map[string]EntryKind{
				".":        KindDirectory,
				"keep.txt": KindFile,
			},
		},
		{
			name:    "ignore matches names at any depth",
			ignores: []string{"skip.txt"},
			files:   map[string]fixture{"skip.txt": {}, "d/skip.txt": {}, "d/keep.txt": {}},
			want: map[string]EntryKind{
				".":          KindDirectory,
				"d":          KindDirectory,
				"d/keep.txt": KindFile,
			},
		},
		{
			name:  "zero-value Walker ignores nothing",
			files: map[string]fixture{"keep.txt": {}, ".git/HEAD": {}},
			want: map[string]EntryKind{
				".":         KindDirectory,
				"keep.txt":  KindFile,
				".git":      KindDirectory,
				".git/HEAD": KindFile,
			},
		},
		{
			name: "empty directory",
			want: map[string]EntryKind{".": KindDirectory},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			materialize(t, root, tt.files)

			w := Walker{Ignores: tt.ignores}
			got, err := w.WalkTree(root)
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(kinds(got), tt.want) {
				t.Errorf("walk mismatch\n got:  %v\n want: %v", kinds(got), tt.want)
			}
		})
	}
}

func TestWalkTreeSingleFileRoot(t *testing.T) {
	// A file root yields exactly one entry keyed "." — the same key a
	// differently-named file on the other side produces, so they can be compared.
	root := t.TempDir()
	p := writeFile(t, root, "zshrc", "x")

	var w Walker
	got, err := w.WalkTree(p)
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{{CanonicalPath: ".", Kind: KindFile}}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWalkTreeMissingRoot(t *testing.T) {
	var w Walker
	_, err := w.WalkTree(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected an error for a missing root")
	}
	// Callers rely on this to treat "not deployed yet" as an empty tree.
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error %v does not wrap fs.ErrNotExist", err)
	}
}
