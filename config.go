package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Global GlobalConfig `toml:"global"`
	File   []FileEntry  `toml:"file"`
}

type GlobalConfig struct {
	ConfigDir       string   `toml:"config_dir"`
	DefaultStrategy Strategy `toml:"default_strategy"`
}

type FileEntry struct {
	Name     string    `toml:"name"`
	Source   string    `toml:"source"`
	Target   string    `toml:"target"`
	Strategy *Strategy `toml:"strategy"`
	Ignore   []string  `toml:"ignore"`
}

func LoadGlobalConfig(path string) (*GlobalConfig, error) {
	var config GlobalConfig
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type Strategy string

const (
	StrategySymlink Strategy = "symlink"
	StrategyCopy    Strategy = "copy"
)

func (s *Strategy) UnmarshalText(text []byte) error {
	asStr := string(text)

	switch asStr {
	case "symlink":
		*s = StrategySymlink
	case "copy":
		*s = StrategyCopy
	default:
		return fmt.Errorf("unknown strategy %q", asStr)
	}
	return nil
}
