package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	if len(os.Args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s url [url ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}
	baseCtx := context.Background()
	for _, arg := range os.Args[1:] {
		fmt.Print(" - cloning\n")

		ctx, cancel := context.WithTimeout(baseCtx, 30*time.Second)
		git := exec.CommandContext(ctx, "git", "clone", arg)
		git.Dir = tools.PluginDir()
		if _, err := git.Output(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run git command: %s\n", err)
			os.Exit(1)
		}
		cancel()

		plugin := plugins.Add(arg)

		fmt.Printf(" - create config %s\n", plugin.ConfigFilePath())
		f, err := os.Create(plugin.ConfigFilePath())
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config file: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(" - rewrite files\n")
	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s\n", err)
		os.Exit(1)
	}
}
