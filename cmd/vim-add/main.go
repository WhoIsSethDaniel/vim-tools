package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var name string
	flag.StringVar(&name, "n", "", "Name for given URL (only one URL may be specified)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-n name <url> | <url> ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	if name != "" && flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "When -n is provided only one URL may be given")
		os.Exit(1)
	}

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}

	baseCtx := context.Background()
	for _, arg := range flag.Args() {
		fmt.Print(" - cloning\n")

		ctx, cancel := context.WithTimeout(baseCtx, 30*time.Second)
		var git *exec.Cmd
		if name == "" {
			git = exec.CommandContext(ctx, "git", "clone", arg)
		} else {
			git = exec.CommandContext(ctx, "git", "clone", arg, name)
		}
		git.Dir = tools.PluginDir()
		if _, err := git.Output(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run git command: %s\n", err)
			os.Exit(1)
		}
		cancel()
		plugins.Add(arg, name)
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
