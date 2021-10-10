package main

import (
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	if len(os.Args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s plugin [plugin ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	plugins, _ := tools.Read()
	for _, arg := range os.Args[1:] {
		plugin, ok := plugins[arg]
		if !ok {
			fmt.Fprintf(os.Stderr, "No such plugin: %s\n", arg)
			os.Exit(1)
		}
		pluginDir := filepath.Join(tools.PluginDir(), arg)
		fi, err := os.Stat(pluginDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "plugin '%s' directory does not exist: %s\n", arg, err)
			os.Exit(1)
		}
		if !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "Plugin '%s' has no directory: %s\n", arg, err)
			os.Exit(1)
		}

		fmt.Print(" - removing directory\n")
		if err := os.RemoveAll(pluginDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove plugin %s dir: %s\n", arg, err)
			os.Exit(1)
		}

		fmt.Print(" - removing config file\n")
		if err := os.Remove(plugin.ConfigFilePath()); err != nil {
			fmt.Printf("Failed to remove plugin %s config file: %s\n", arg, err)
			os.Exit(1)
		}
		plugins.Remove(plugin)
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
