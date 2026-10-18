package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// sharedDefaults returns a defaults slice with spare capacity. That is the
// situation where append(defaults, ...) writes into memory that other slices
// built from the same defaults can also see.
func sharedDefaults() []string {
	d := make([]string, 1, 8)
	d[0] = ".git"
	return d
}

func TestMergeIgnores(t *testing.T) {
	tests := []struct {
		name            string
		defaults, extra []string
		want            []string
	}{
		{"defaults only", []string{".git"}, nil, []string{".git"}},
		{"extras follow defaults in order", []string{".git"}, []string{"a", "b"}, []string{".git", "a", "b"}},
		{"no defaults", nil, []string{"a"}, []string{"a"}},
		{"both empty", nil, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeIgnores(tt.defaults, tt.extra)
			if !slices.Equal(got, tt.want) {
				t.Errorf("mergeIgnores(%v, %v) = %v, want %v", tt.defaults, tt.extra, got, tt.want)
			}
		})
	}
}

func TestMergeIgnoresLeavesDefaultsAlone(t *testing.T) {
	d := sharedDefaults()
	_ = mergeIgnores(d, []string{"themes"})

	if !slices.Equal(d, []string{".git"}) {
		t.Errorf("defaults changed to %v", d)
	}
	// An in-place append would leave len(d) alone but write into the spare
	// capacity, where any other slice over the same array would see it.
	for i, s := range d[len(d):cap(d)] {
		if s != "" {
			t.Errorf("spare slot %d of the defaults array was written: %q", i, s)
		}
	}
}

func TestMergeIgnoresDoesNotLeakBetweenEntries(t *testing.T) {
	d := sharedDefaults()
	a := mergeIgnores(d, []string{"themes"})
	b := mergeIgnores(d, []string{"lua"})

	if want := []string{".git", "themes"}; !slices.Equal(a, want) {
		t.Errorf("entry A = %v, want %v (changed after entry B was built)", a, want)
	}
	if want := []string{".git", "lua"}; !slices.Equal(b, want) {
		t.Errorf("entry B = %v, want %v", b, want)
	}
}

func TestMergeIgnoresLeavesPackageDefaultsAlone(t *testing.T) {
	before := slices.Clone(DefaultIgnores)
	_ = mergeIgnores(DefaultIgnores, []string{"x", "y"})

	if !slices.Equal(DefaultIgnores, before) {
		t.Errorf("DefaultIgnores changed from %v to %v", before, DefaultIgnores)
	}
}

func TestResolveConfigDir(t *testing.T) {
	const home = "/h"

	tests := []struct {
		name    string
		file    string // path of the config file
		dir     string // the config_dir value, as written
		env     map[string]string
		unset   []string
		want    string
		wantErr string // substring of the expected error; empty means no error
	}{
		{name: "relative joins the config file's directory", file: "/r/c.toml", dir: "configs", want: "/r/configs"},
		{name: "relative with .. is cleaned", file: "/r/c.toml", dir: "../shared", want: "/shared"},
		{name: "config file with no directory part", file: "garlic_pear.toml", dir: "configs", want: "configs"},
		{name: "omitted config_dir means the config file's own directory", file: "/r/c.toml", dir: "", want: "/r"},

		{name: "absolute is used as written", file: "/r/c.toml", dir: "/abs/x", want: "/abs/x"},
		{name: "tilde expands", file: "/r/c.toml", dir: "~/x", want: home + "/x"},
		{name: "bare tilde", file: "/r/c.toml", dir: "~", want: home},
		{name: "HOME variable", file: "/r/c.toml", dir: "${HOME}/x", want: home + "/x"},

		{
			// Expanding after joining would see "/r/${SUB}" and fail to find a
			// relative path to join; joining the raw field would drop the expansion.
			name: "variable expanding to a relative path is joined",
			file: "/r/c.toml", dir: "${SUB}",
			env:  map[string]string{"SUB": "mysub"},
			want: "/r/mysub",
		},
		{
			name: "variable expanding to an absolute path is not joined",
			file: "/r/c.toml", dir: "${SUB}/x",
			env:  map[string]string{"SUB": "/abs"},
			want: "/abs/x",
		},

		{
			name: "unset variable is an error",
			file: "/r/c.toml", dir: "${NOPE}/x",
			unset:   []string{"NOPE"},
			wantErr: "NOPE",
		},
		{
			name: "empty variable is an error",
			file: "/r/c.toml", dir: "${EMPTYVAR}/x",
			env:     map[string]string{"EMPTYVAR": ""},
			wantErr: "EMPTYVAR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", home)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			for _, k := range tt.unset {
				t.Setenv(k, "")
				if err := os.Unsetenv(k); err != nil {
					t.Fatal(err)
				}
			}

			got, err := resolveConfigDir(tt.file, tt.dir)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("resolveConfigDir(%q, %q) = %q, want error containing %q", tt.file, tt.dir, got, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not mention %q", err, tt.wantErr)
				}
				if got != "" {
					t.Errorf("on error got %q, want empty string", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveConfigDir(%q, %q) unexpected error: %v", tt.file, tt.dir, err)
			}
			if got != tt.want {
				t.Errorf("resolveConfigDir(%q, %q) = %q, want %q", tt.file, tt.dir, got, tt.want)
			}
		})
	}
}
