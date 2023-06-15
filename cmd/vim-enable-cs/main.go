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
		fmt.Fprintf(os.Stderr, "Usage: %s <colorscheme>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	cs := os.Args[1]
	for i := range plugins {
		if plugins[i].Colorscheme {
			if plugins[i].Name == cs {
				fmt.Printf("enabling %s\n", cs)
				plugins[i] = plugins[i].Enable()
			} else {
				fmt.Printf("disabling %s\n", plugins[i].Name)
				plugins[i] = plugins[i].Disable()
			}
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
