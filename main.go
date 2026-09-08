package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Need arguments")
		os.Exit(1)
	}

	args := os.Args[1:]

	configPath := args[0]

	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Config: %+v\n", config)
}
