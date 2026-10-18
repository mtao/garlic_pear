package main

import (
	"os"
	"strings"
	"testing"
)

func TestExpandPath(t *testing.T) {
	const home = "/home/test"

	tests := []struct {
		name    string
		in      string
		env     map[string]string // set for the duration of the case
		unset   []string          // explicitly removed from the environment
		want    string
		wantErr string // substring of the expected error; empty means no error
	}{
		{name: "empty string", in: "", want: ""},
		{name: "plain path untouched", in: "/etc/hosts", want: "/etc/hosts"},
		{name: "relative path untouched", in: "a/b", want: "a/b"},

		{name: "bare tilde", in: "~", want: home},
		{name: "tilde slash", in: "~/", want: home},
		{name: "tilde path", in: "~/.zshrc", want: home + "/.zshrc"},
		{name: "tilde user is literal", in: "~bob/x", want: "~bob/x"},
		{name: "tilde mid-path is literal", in: "/a/~/b", want: "/a/~/b"},

		{
			name: "braced variable",
			in:   "${FOO}/y",
			env:  map[string]string{"FOO": "bar"},
			want: "bar/y",
		},
		{
			name: "bare variable form",
			in:   "$FOO/y",
			env:  map[string]string{"FOO": "bar"},
			want: "bar/y",
		},
		{name: "HOME variable", in: "${HOME}/y", want: home + "/y"},
		{
			name: "tilde and variable together",
			in:   "~/${FOO}",
			env:  map[string]string{"FOO": "bar"},
			want: home + "/bar",
		},
		{
			// Tilde expansion runs first, so a ~ arriving from a variable's
			// value must not be expanded a second time.
			name: "tilde from variable value is not re-expanded",
			in:   "${T}/y",
			env:  map[string]string{"T": "~/x"},
			want: "~/x/y",
		},

		{
			name: "XDG_CONFIG_HOME set",
			in:   "${XDG_CONFIG_HOME}/mpv",
			env:  map[string]string{"XDG_CONFIG_HOME": "/custom"},
			want: "/custom/mpv",
		},
		{
			name:  "XDG_CONFIG_HOME unset falls back to ~/.config",
			in:    "${XDG_CONFIG_HOME}/mpv",
			unset: []string{"XDG_CONFIG_HOME"},
			want:  home + "/.config/mpv",
		},
		{
			// An empty value must behave like unset, or the path collapses to "/mpv".
			name: "XDG_CONFIG_HOME empty falls back to ~/.config",
			in:   "${XDG_CONFIG_HOME}/mpv",
			env:  map[string]string{"XDG_CONFIG_HOME": ""},
			want: home + "/.config/mpv",
		},

		{
			name:    "unset variable is an error",
			in:      "${NOPE}/z",
			unset:   []string{"NOPE"},
			wantErr: "NOPE",
		},
		{
			// Same collapse-to-root hazard as the XDG case, with no fallback.
			name:    "empty variable is an error",
			in:      "${EMPTYVAR}/etc",
			env:     map[string]string{"EMPTYVAR": ""},
			wantErr: "EMPTYVAR",
		},
		{
			name:    "first of several failures is reported",
			in:      "${AAA}/${BBB}",
			unset:   []string{"AAA", "BBB"},
			wantErr: "AAA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", home)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			for _, k := range tt.unset {
				// Setenv registers the cleanup that restores the original
				// value; Unsetenv then makes the variable genuinely absent.
				t.Setenv(k, "")
				if err := os.Unsetenv(k); err != nil {
					t.Fatal(err)
				}
			}

			got, err := ExpandPath(tt.in)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ExpandPath(%q) = %q, want error containing %q", tt.in, got, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not mention %q", err, tt.wantErr)
				}
				if got != "" {
					t.Errorf("on error got %q, want empty string (no half-expanded path)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandPath(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
