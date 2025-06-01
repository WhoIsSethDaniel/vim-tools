package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var name, version string
	flag.StringVar(&name, "n", "", "Name for given URL (only one URL may be specified)")
	flag.StringVar(&version, "v", "", "Version to freeze on")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-n name <url> | <url> ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	if name != "" && flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "When -n is provided only one URL may be given")
		os.Exit(1)
	}
	if version != "" && flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "When -v is provided only one URL may be given")
		os.Exit(1)
	}

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}

	for _, arg := range flag.Args() {
		plugins.Add(arg, name, version)
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
