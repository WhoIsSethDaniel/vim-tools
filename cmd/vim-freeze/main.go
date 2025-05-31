package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var version string
	flag.StringVar(&version, "v", "", "Freeze to a particular branch/tag")
	flag.Parse()

	plugins, _ := tools.Read()
	if version == "" || flag.NArg() == 0 {
		if version == "" {
			fmt.Fprint(os.Stderr, "-v is required\n")
		}
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s -v <branch/tag> plugin [plugin ...]\n",
			filepath.Base(os.Args[0]),
		)
		os.Exit(1)
	}
	for _, arg := range flag.Args() {
		plugin, ok := plugins[arg]
		if !ok {
			fmt.Fprintf(os.Stderr, "cannot find %s\n", arg)
		} else {
			plugins[arg] = plugin.Freeze(version)
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
