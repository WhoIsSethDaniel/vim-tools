package main

import (
	"flag"
	"fmt"
	"os"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var listURL, listVersion, showFlags bool
	flag.BoolVar(&listURL, "u", false, "List the repo URL along with the name")
	flag.BoolVar(&listVersion, "v", false, "List the version the repo is frozen to, if any")
	flag.BoolVar(
		&showFlags,
		"f",
		false,
		"Show flags for each module.",
	)
	flag.Parse()

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read all plugin file: %s\n", err)
		os.Exit(1)
	}

	pluginNames := plugins.SortedNames()
	for _, name := range pluginNames {
		plugin := plugins[name]
		flags := ""
		if plugin.IsColorscheme() {
			flags += "C"
		} else {
			flags += " "
		}
		if plugin.IsDisabled() {
			flags += "D"
		} else {
			flags += " "
		}
		if plugin.HasVersion() {
			flags += "F"
		} else {
			flags += " "
		}
		if showFlags {
			fmt.Printf("%s  ", flags)
		}
		fmt.Print(name)
		if listVersion && plugin.HasVersion() {
			fmt.Printf(" [%s]", plugin.Version)
		}
		if listURL {
			fmt.Printf(" [%s]", plugin.URL)
		}
		fmt.Println()
	}
}
