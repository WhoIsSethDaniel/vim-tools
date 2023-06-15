package main

import (
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	plugins, _ := tools.Read()
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s plugin [plugin ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	for _, arg := range os.Args[1:] {
		plugin, ok := plugins[arg]
		switch {
		case !ok:
			fmt.Fprintf(os.Stderr, "cannot find %s\n", arg)
		case plugin.Colorscheme:
			// only allow one colorscheme to be active at any one time
			plugins[arg] = plugin.Enable()
			for i := range plugins {
				if plugins[i].Colorscheme && plugin.Name != plugins[i].Name {
					plugins[i] = plugins[i].Disable()
				}
			}
		default:
			plugins[arg] = plugin.Enable()
		}
	}

	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s", err)
		os.Exit(1)
	}
}
