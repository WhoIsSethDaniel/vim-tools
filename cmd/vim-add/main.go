package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	if len(os.Args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s url [url ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	plugins, _ := tools.Read()
	for _, arg := range os.Args[1:] {
		fmt.Print(" - cloning\n")
		git := exec.Command("git", "clone", arg)
		git.Dir = tools.PluginDir()
		if _, err := git.Output(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run git command: %s\n", err)
			os.Exit(1)
		}

		plugin := plugins.Add(arg)

		fmt.Printf(" - create config %s\n", plugin.ConfigFilePath())
		f, err := os.Create(plugin.ConfigFilePath())
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config file: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(" - rewrite files")
	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s", err)
		os.Exit(1)
	}
}
