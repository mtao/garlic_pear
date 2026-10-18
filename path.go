package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(s string) (string, error) {
	if rest, ok := strings.CutPrefix(s, "~"); ok && (rest == "" || rest[0] == '/') {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		s = filepath.Join(homeDir, rest)
	}

	var expandErr error
	expanded := os.Expand(s, func(name string) string {
		v, ok := os.LookupEnv(name)
		if ok && v != "" {
			return v
		}
		if name == "XDG_CONFIG_HOME" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				expandErr = err
				return ""
			}
			return filepath.Join(homeDir, ".config")
		}
		if expandErr == nil {
			expandErr = fmt.Errorf("environment variable %s is not set", name)
		}
		return ""
	})
	if expandErr != nil {
		return "", expandErr
	}

	return expanded, nil
}
