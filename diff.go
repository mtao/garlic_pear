package main

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type ComparisonType int

const (
	Same ComparisonType = iota
	SourceOnly
	Differs
	TargetOnly
	Uncomparable
)

func (c ComparisonType) String() string {
	switch c {
	case Same:
		return "Same"
	case SourceOnly:
		return "SourceOnly"
	case Differs:
		return "Differs"
	case TargetOnly:
		return "TargetOnly"
	case Uncomparable:
		return "Uncomparable"
	default:
		return fmt.Sprintf("ComparisonType(%d)", int(c))
	}
}

type SyncDirection int

const (
	DirNone SyncDirection = iota
	ToTarget
	ToSource
)

func (c SyncDirection) String() string {
	switch c {
	case DirNone:
		return "—"
	case ToTarget:
		return "⇒ Target"
	case ToSource:
		return "⇒ Source"
	default:
		return fmt.Sprintf("SyncDirection(%d)", int(c))
	}
}

type DiffEntry struct {
	CanonicalPath string
	Comparison    ComparisonType
	Direction     SyncDirection
}

func (e DiffEntry) String() string {
	return fmt.Sprintf("%-12v %-8v %s", e.Comparison, e.Direction, e.CanonicalPath)
}

type DiffReport []DiffEntry

func (r DiffReport) String() string {
	var b strings.Builder
	for _, e := range r {
		fmt.Fprintln(&b, e)
	}
	return b.String()
}

func sameContents(pathA, pathB string) (bool, error) {
	a, err := os.ReadFile(pathA)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(pathB)
	if err != nil {
		return false, err
	}
	return bytes.Equal(a, b), nil
}

func compareFiles(sourceRoot, targetRoot, rel string) (ComparisonType, SyncDirection, error) {
	srcPath := filepath.Join(sourceRoot, rel)
	tgtPath := filepath.Join(targetRoot, rel)

	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return Uncomparable, DirNone, err
	}
	tgtStat, err := os.Stat(tgtPath)
	if err != nil {
		return Uncomparable, DirNone, err
	}

	differs := srcStat.Size() != tgtStat.Size() ||
		srcStat.Mode().Perm() != tgtStat.Mode().Perm()

	if !differs {
		eq, err := sameContents(srcPath, tgtPath)
		if err != nil {
			return Uncomparable, DirNone, err
		}
		differs = !eq
	}

	if !differs {
		return Same, DirNone, nil
	}

	srcTime := srcStat.ModTime()
	tgtTime := tgtStat.ModTime()

	switch {
	case srcTime.After(tgtTime):
		return Differs, ToTarget, nil
	case tgtTime.After(srcTime):
		return Differs, ToSource, nil
		// same modtime but different contents?
	default:
		return Differs, DirNone, nil
	}
}

func BuildDiff(sourceEntries, targetEntries []Entry, sourceRoot, targetRoot string) (DiffReport, error) {
	srcIdx := make(map[string]Entry, len(sourceEntries))
	for _, e := range sourceEntries {
		srcIdx[e.CanonicalPath] = e
	}
	tgtIdx := make(map[string]Entry, len(targetEntries))
	for _, e := range targetEntries {
		tgtIdx[e.CanonicalPath] = e
	}

	paths := make(map[string]struct{}, len(srcIdx)+len(tgtIdx))
	for path := range srcIdx {
		paths[path] = struct{}{}
	}
	for path := range tgtIdx {
		paths[path] = struct{}{}
	}

	var report []DiffEntry

	for _, p := range slices.Sorted(maps.Keys(paths)) {

		s, inSrc := srcIdx[p]
		t, inTgt := tgtIdx[p]

		switch {
		case inSrc && !inTgt:
			report = append(report,
				DiffEntry{p, SourceOnly, ToTarget})
		case !inSrc && inTgt:
			report = append(report,
				DiffEntry{p, TargetOnly, DirNone})

		// Cases where both exist
		// ---------
		case s.Kind != t.Kind:
			report = append(report,
				DiffEntry{p, Uncomparable, DirNone})

		// both exist, same kind, but is a directory
		case s.Kind == KindDirectory:
			report = append(report, DiffEntry{p, Same, DirNone})

			// compare mtime of two files
		default:
			comp, dir, err := compareFiles(sourceRoot, targetRoot, p)
			if err != nil {
				return nil, err
			}
			report = append(report, DiffEntry{p, comp, dir})

		}

	}
	return report, nil
}
