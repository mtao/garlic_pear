package main

import (
	"maps"
	"strings"
	"testing"
)

// outcome is the comparable part of a DiffEntry once the path is the map key.
type outcome struct {
	Comparison ComparisonType
	Direction  SyncDirection
}

// buildAndIndex walks both trees and returns the report keyed by path.
func buildAndIndex(t *testing.T, srcRoot, tgtRoot string) map[string]DiffEntry {
	t.Helper()
	w := Walker{Ignores: DefaultIgnores}

	src, err := w.WalkTree(srcRoot)
	if err != nil {
		t.Fatal(err)
	}
	tgt, err := w.WalkTree(tgtRoot)
	if err != nil {
		t.Fatal(err)
	}
	report, err := BuildDiff(src, tgt, srcRoot, tgtRoot)
	if err != nil {
		t.Fatal(err)
	}

	idx := make(map[string]DiffEntry, len(report))
	for _, e := range report {
		idx[e.CanonicalPath] = e
	}
	return idx
}

func TestBuildDiff(t *testing.T) {
	tests := []struct {
		name     string
		src, tgt map[string]fixture
		// want lists every entry expected to be anything other than Same.
		want map[string]outcome
		// wantSame names paths that must be present and Same.
		wantSame []string
	}{
		{
			name:     "identical files",
			src:      map[string]fixture{"a.txt": {content: "hello"}},
			tgt:      map[string]fixture{"a.txt": {content: "hello"}},
			wantSame: []string{"a.txt", "."},
		},
		{
			name: "source only is deployed to target",
			src:  map[string]fixture{"a.txt": {content: "x"}},
			tgt:  map[string]fixture{},
			want: map[string]outcome{"a.txt": {SourceOnly, ToTarget}},
		},
		{
			name: "target only is left undecided",
			src:  map[string]fixture{},
			tgt:  map[string]fixture{"a.txt": {content: "x"}},
			want: map[string]outcome{"a.txt": {TargetOnly, DirNone}},
		},
		{
			name: "differs, source newer",
			src:  map[string]fixture{"a.txt": {content: "new contents", mtime: newTime}},
			tgt:  map[string]fixture{"a.txt": {content: "old", mtime: oldTime}},
			want: map[string]outcome{"a.txt": {Differs, ToTarget}},
		},
		{
			name: "differs, target newer",
			src:  map[string]fixture{"a.txt": {content: "old", mtime: oldTime}},
			tgt:  map[string]fixture{"a.txt": {content: "new contents", mtime: newTime}},
			want: map[string]outcome{"a.txt": {Differs, ToSource}},
		},
		{
			// The reason contents, not mtime, decide whether files differ: a
			// copy-deployed file gets a fresh mtime but is still identical.
			name:     "same contents with different mtime is Same",
			src:      map[string]fixture{"a.txt": {content: "identical", mtime: oldTime}},
			tgt:      map[string]fixture{"a.txt": {content: "identical", mtime: newTime}},
			wantSame: []string{"a.txt"},
		},
		{
			// Equal size forces the content read; the contents differ.
			name: "same size, different contents",
			src:  map[string]fixture{"a.txt": {content: "aaaa", mtime: newTime}},
			tgt:  map[string]fixture{"a.txt": {content: "bbbb", mtime: oldTime}},
			want: map[string]outcome{"a.txt": {Differs, ToTarget}},
		},
		{
			// No evidence which side is right, so no direction is offered.
			name: "differs with equal mtime has no direction",
			src:  map[string]fixture{"a.txt": {content: "left", mtime: oldTime}},
			tgt:  map[string]fixture{"a.txt": {content: "right", mtime: oldTime}},
			want: map[string]outcome{"a.txt": {Differs, DirNone}},
		},
		{
			// chmod changes ctime but not mtime, so a mode-only difference
			// cannot be directed either.
			name: "mode-only difference",
			src:  map[string]fixture{"run.sh": {content: "#!/bin/sh", mode: 0o755, mtime: oldTime}},
			tgt:  map[string]fixture{"run.sh": {content: "#!/bin/sh", mode: 0o644, mtime: oldTime}},
			want: map[string]outcome{"run.sh": {Differs, DirNone}},
		},
		{
			name: "file on one side, directory on the other",
			src:  map[string]fixture{"thing": {content: "i am a file"}},
			tgt:  map[string]fixture{"thing/inner.txt": {content: "inside"}},
			want: map[string]outcome{
				"thing":           {Uncomparable, DirNone},
				"thing/inner.txt": {TargetOnly, DirNone},
			},
		},
		{
			name:     "directories present on both sides are Same",
			src:      map[string]fixture{"d/a.txt": {content: "x"}},
			tgt:      map[string]fixture{"d/a.txt": {content: "x"}},
			wantSame: []string{".", "d", "d/a.txt"},
		},
		{
			name:     "ignored directory in target produces no entries",
			src:      map[string]fixture{"a.txt": {content: "x"}},
			tgt:      map[string]fixture{"a.txt": {content: "x"}, ".git/HEAD": {content: "junk"}},
			wantSame: []string{"a.txt"},
		},
		{
			name: "mixed tree",
			src: map[string]fixture{
				"same.txt":     {content: "s"},
				"src_only.txt": {content: "s"},
				"new.txt":      {content: "fresh", mtime: newTime},
			},
			tgt: map[string]fixture{
				"same.txt":     {content: "s"},
				"tgt_only.txt": {content: "t"},
				"new.txt":      {content: "stale", mtime: oldTime},
			},
			want: map[string]outcome{
				"src_only.txt": {SourceOnly, ToTarget},
				"tgt_only.txt": {TargetOnly, DirNone},
				"new.txt":      {Differs, ToTarget},
			},
			wantSame: []string{"same.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcRoot, tgtRoot := t.TempDir(), t.TempDir()
			materialize(t, srcRoot, tt.src)
			materialize(t, tgtRoot, tt.tgt)

			idx := buildAndIndex(t, srcRoot, tgtRoot)

			got := make(map[string]outcome)
			for p, e := range idx {
				if e.Comparison != Same {
					got[p] = outcome{e.Comparison, e.Direction}
				}
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("non-Same entries mismatch\n got:  %v\n want: %v", got, tt.want)
			}

			for _, p := range tt.wantSame {
				e, ok := idx[p]
				switch {
				case !ok:
					t.Errorf("%q missing from report, want Same", p)
				case e.Comparison != Same:
					t.Errorf("%q = %v, want Same", p, e.Comparison)
				}
			}
		})
	}
}

func TestBuildDiffSingleFileRoots(t *testing.T) {
	// Source "zshrc" deploys to target ".zshrc": different base names, so the
	// walker keys both as "." and they compare directly.
	tests := []struct {
		name     string
		src, tgt fixture
		want     outcome
	}{
		{
			name: "identical",
			src:  fixture{content: "same"},
			tgt:  fixture{content: "same"},
			want: outcome{Same, DirNone},
		},
		{
			name: "source newer",
			src:  fixture{content: "left", mtime: newTime},
			tgt:  fixture{content: "right", mtime: oldTime},
			want: outcome{Differs, ToTarget},
		},
		{
			name: "target newer",
			src:  fixture{content: "left", mtime: oldTime},
			tgt:  fixture{content: "right", mtime: newTime},
			want: outcome{Differs, ToSource},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			materialize(t, dir, map[string]fixture{"zshrc": tt.src, ".zshrc": tt.tgt})

			idx := buildAndIndex(t, dir+"/zshrc", dir+"/.zshrc")
			if len(idx) != 1 {
				t.Fatalf("got %d entries, want exactly 1: %v", len(idx), idx)
			}
			e, ok := idx["."]
			if !ok {
				t.Fatalf("no \".\" entry in %v", idx)
			}
			if got := (outcome{e.Comparison, e.Direction}); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDiffIsSorted(t *testing.T) {
	// Map iteration order is randomized, so sorted output is a real property.
	srcRoot, tgtRoot := t.TempDir(), t.TempDir()
	files := map[string]fixture{}
	for _, n := range []string{"m.txt", "a.txt", "z.txt", "b/c.txt", "b/a.txt"} {
		files[n] = fixture{content: "x"}
	}
	materialize(t, srcRoot, files)

	w := Walker{}
	src, err := w.WalkTree(srcRoot)
	if err != nil {
		t.Fatal(err)
	}
	report, err := BuildDiff(src, nil, srcRoot, tgtRoot)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(report); i++ {
		if report[i-1].CanonicalPath > report[i].CanonicalPath {
			t.Fatalf("report not sorted at %d: %q before %q",
				i, report[i-1].CanonicalPath, report[i].CanonicalPath)
		}
	}
}

func TestBuildDiffMissingFileIsError(t *testing.T) {
	// An entry that vanishes between walking and comparing must surface as an
	// error, not be silently reported as Same.
	srcRoot, tgtRoot := t.TempDir(), t.TempDir()
	materialize(t, srcRoot, map[string]fixture{"a.txt": {content: "x"}})
	materialize(t, tgtRoot, map[string]fixture{"a.txt": {content: "x"}})

	w := Walker{}
	src, _ := w.WalkTree(srcRoot)
	tgt, _ := w.WalkTree(tgtRoot)

	if _, err := BuildDiff(src, tgt, srcRoot, t.TempDir()); err == nil {
		t.Error("expected an error when the target file does not exist on disk")
	}
}

func TestStringers(t *testing.T) {
	t.Run("ComparisonType values are distinct and non-empty", func(t *testing.T) {
		seen := map[string]ComparisonType{}
		for _, c := range []ComparisonType{Same, SourceOnly, Differs, TargetOnly, Uncomparable} {
			s := c.String()
			if s == "" {
				t.Errorf("%d has an empty String()", int(c))
			}
			if prev, dup := seen[s]; dup {
				t.Errorf("%d and %d both print as %q", int(prev), int(c), s)
			}
			seen[s] = c
		}
	})

	t.Run("SyncDirection values are distinct and non-empty", func(t *testing.T) {
		seen := map[string]SyncDirection{}
		for _, d := range []SyncDirection{DirNone, ToTarget, ToSource} {
			s := d.String()
			if s == "" {
				t.Errorf("%d has an empty String()", int(d))
			}
			if prev, dup := seen[s]; dup {
				t.Errorf("%d and %d both print as %q", int(prev), int(d), s)
			}
			seen[s] = d
		}
	})

	t.Run("out-of-range values report their number instead of panicking", func(t *testing.T) {
		if got := ComparisonType(99).String(); !strings.Contains(got, "99") {
			t.Errorf("ComparisonType(99) = %q, want it to contain 99", got)
		}
		if got := SyncDirection(99).String(); !strings.Contains(got, "99") {
			t.Errorf("SyncDirection(99) = %q, want it to contain 99", got)
		}
	})

	t.Run("DiffReport prints one line per entry", func(t *testing.T) {
		r := DiffReport{
			{CanonicalPath: "a.txt", Comparison: SourceOnly, Direction: ToTarget},
			{CanonicalPath: "b/c.txt", Comparison: TargetOnly, Direction: DirNone},
		}
		lines := strings.Split(strings.TrimSuffix(r.String(), "\n"), "\n")
		if len(lines) != len(r) {
			t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(r), r)
		}
		for i, e := range r {
			if !strings.HasSuffix(lines[i], e.CanonicalPath) {
				t.Errorf("line %d = %q, want it to end with %q", i, lines[i], e.CanonicalPath)
			}
		}
	})
}
