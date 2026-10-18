package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

func mergeIgnores(defaults, extra []string) []string {
	return slices.Concat(defaults, extra)
}

func resolveConfigDir(configFile, configDir string) (string, error) {
	configDir, err := ExpandPath(configDir)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(configDir) {
		configDir = filepath.Join(filepath.Dir(configFile), configDir)
	}
	return configDir, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "garlic_pear.toml", "path to config file")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	configPath, err := ExpandPath(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var failed bool

	configDir, err := resolveConfigDir(configPath, config.Global.ConfigDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch args[0] {
	case "status":
		for _, entry := range config.File {

			walker := Walker{Ignores: mergeIgnores(DefaultIgnores, entry.Ignore)}
			srcRoot := filepath.Join(configDir, entry.Source)
			tgtRoot, err := ExpandPath(entry.Target)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				failed = true
				continue
			}
			srcTree, err := walker.WalkTree(srcRoot)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				failed = true
				continue
			}
			tgtTree, err := walker.WalkTree(tgtRoot)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					fmt.Fprintln(os.Stderr, err)
					failed = true
					continue
				}
				tgtTree = nil
			}
			fmt.Printf("Entry %s:\n", entry.Name)
			diff, err := BuildDiff(srcTree, tgtTree, srcRoot, tgtRoot)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				failed = true
				continue
			}

			for _, d := range diff {
				if d.Comparison == Same {
					continue
				}
				fmt.Println(d)
			}

		}
	}
	if failed {
		os.Exit(1)
	}
}
