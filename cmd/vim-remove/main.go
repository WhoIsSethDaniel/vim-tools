package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var keep, noRemove bool
	flag.BoolVar(&keep, "k", false, "Remove the directory but keep the plugin registered")
	flag.BoolVar(&noRemove, "u", false, "Do not remove the directory but unregister the plugin")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-ku] plugin [plugin ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}

	for _, arg := range flag.Args() {
		plugin, ok := plugins[arg]
		if !ok {
			fmt.Fprintf(os.Stderr, "No such plugin: %s\n", arg)
			os.Exit(1)
		}
		if !noRemove {
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
		}

		if !keep {
			plugins.Remove(plugin)
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
