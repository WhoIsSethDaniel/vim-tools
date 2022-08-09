package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var create bool
	flag.BoolVar(&create, "c", false, "Create the config file for the given plugin(s)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s plugin [plugin ...]\n", filepath.Base(os.Args[0]))
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
			fmt.Fprintf(os.Stderr, "cannot find %s\n", arg)
		}

		if create {
			if _, err := os.Stat(plugin.ConfigFilePath()); err != nil {
				f, err := os.Create(plugin.ConfigFilePath())
				f.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to create config file: %s\n", err)
					os.Exit(1)
				}
			}
		}
		fmt.Printf("%s\n", plugin.ConfigFilePath())
	}

	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s\n", err)
		os.Exit(1)
	}
}
