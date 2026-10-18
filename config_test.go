package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	return writeFile(t, t.TempDir(), "garlic_pear.toml", body)
}

func TestLoadConfig(t *testing.T) {
	path := writeConfig(t, `
[global]
config_dir = "configs"
default_strategy = "symlink"

[[file]]
name = "zshrc"
source = "zshrc"
target = "~/.zshrc"

[[file]]
name = "hypr"
source = "hypr"
target = "~/.config/hypr"
strategy = "copy"
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Global.ConfigDir != "configs" {
		t.Errorf("ConfigDir = %q, want %q", cfg.Global.ConfigDir, "configs")
	}
	if cfg.Global.DefaultStrategy != StrategySymlink {
		t.Errorf("DefaultStrategy = %q, want %q", cfg.Global.DefaultStrategy, StrategySymlink)
	}
	if len(cfg.File) != 2 {
		t.Fatalf("got %d file entries, want 2", len(cfg.File))
	}

	zsh := cfg.File[0]
	if zsh.Name != "zshrc" || zsh.Source != "zshrc" || zsh.Target != "~/.zshrc" {
		t.Errorf("first entry decoded wrong: %+v", zsh)
	}
	// A nil pointer is how "not specified, use the global default" is spelled.
	if zsh.Strategy != nil {
		t.Errorf("unspecified strategy = %q, want nil", *zsh.Strategy)
	}

	hypr := cfg.File[1]
	if hypr.Strategy == nil {
		t.Fatal("hypr strategy is nil, want copy")
	}
	if *hypr.Strategy != StrategyCopy {
		t.Errorf("hypr strategy = %q, want %q", *hypr.Strategy, StrategyCopy)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name: "unknown strategy is rejected, naming the value",
			body: `
[global]
config_dir = "c"
default_strategy = "hardlink"
`,
			wantErr: "hardlink",
		},
		{
			name: "unknown per-file strategy is rejected",
			body: `
[global]
config_dir = "c"
default_strategy = "symlink"

[[file]]
name = "x"
source = "x"
target = "/t"
strategy = "teleport"
`,
			wantErr: "teleport",
		},
		{
			name:    "malformed toml",
			body:    "[global\nconfig_dir = ",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(writeConfig(t, tt.body))
			if err == nil {
				t.Fatalf("expected an error, got config %+v", cfg)
			}
			if cfg != nil {
				t.Errorf("on error got non-nil config %+v", cfg)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "nope.toml"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error %v does not wrap fs.ErrNotExist", err)
	}
}

func TestLoadConfigIgnoresUnknownSections(t *testing.T) {
	// Real configs still contain [[git]] blocks from the old schema. The
	// decoder skips keys it has no field for, so they must not break loading.
	path := writeConfig(t, `
[global]
config_dir = "c"
default_strategy = "symlink"

[[git]]
name = "oh-my-zsh"
repo = "https://example.invalid/x.git"
target = "~/.oh-my-zsh"

[[file]]
name = "a"
source = "a"
target = "/t/a"
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.File) != 1 || cfg.File[0].Name != "a" {
		t.Errorf("file entries decoded wrong: %+v", cfg.File)
	}
}

func TestStrategyUnmarshalText(t *testing.T) {
	tests := []struct {
		in      string
		want    Strategy
		wantErr bool
	}{
		{in: "symlink", want: StrategySymlink},
		{in: "copy", want: StrategyCopy},
		{in: "", wantErr: true},
		{in: "Symlink", wantErr: true}, // case-sensitive
		{in: "link", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var s Strategy
			err := s.UnmarshalText([]byte(tt.in))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalText(%q) = %q, want error", tt.in, s)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if s != tt.want {
				t.Errorf("got %q, want %q", s, tt.want)
			}
		})
	}
}

func TestLoadConfigIgnore(t *testing.T) {
	path := writeConfig(t, `
[global]
config_dir = "c"
default_strategy = "symlink"

[[file]]
name = "with"
source = "a"
target = "/t/a"
ignore = ["themes", "*.swp"]

[[file]]
name = "without"
source = "b"
target = "/t/b"
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.File) != 2 {
		t.Fatalf("got %d file entries, want 2", len(cfg.File))
	}

	if got, want := cfg.File[0].Ignore, []string{"themes", "*.swp"}; !slices.Equal(got, want) {
		t.Errorf("entry with ignore = %v, want %v", got, want)
	}
	// Absent key is a nil slice, which mergeIgnores treats as "no extras".
	if cfg.File[1].Ignore != nil {
		t.Errorf("entry without ignore = %v, want nil", cfg.File[1].Ignore)
	}
}
