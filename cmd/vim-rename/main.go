package main

import (
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	plugins, _ := tools.Read()
	if len(os.Args) <= 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s name new-name\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	fmt.Printf(" - rename plugin\n")
	name := os.Args[1]
	newName := os.Args[2]
	plugin, ok := plugins[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "Cannot find plugin %s", name)
	}
	oldConfig := plugin.ConfigFilePath()
	plugins.Add(plugin.URL, newName)
	delete(plugins, name)

	fmt.Print(" - rename plugin dir\n")
	if err := os.Rename(filepath.Join(tools.PluginDir(), name), filepath.Join(tools.PluginDir(), newName)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rename directory: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(" - rename config file\n")
	// for now don't worry about this failing
	os.Rename(oldConfig, plugins[newName].ConfigFilePath())

	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s", err)
		os.Exit(1)
	}
}
