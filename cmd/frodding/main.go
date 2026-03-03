package main

import (
	"fmt"
	"os"
	"path/filepath"

	frodding "github.com/cubexteam/Frodding"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	configPath := filepath.Join(wd, "server.yml")

	srv, err := frodding.NewServer(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}
}
