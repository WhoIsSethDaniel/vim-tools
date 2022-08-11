package main

import (
	"flag"
	"fmt"
	"os"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var listURL, showFlags bool
	flag.BoolVar(&listURL, "u", false, "List the repo URL along with the name")
	flag.BoolVar(
		&showFlags,
		"f",
		false,
		"List just sorted names of plugins. No categorization, no special flags.",
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
		if plugin.IsFrozen() {
			flags += "F"
		} else {
			flags += " "
		}
		if showFlags {
			fmt.Printf("%s  ", flags)
		}
		fmt.Print(name)
		if listURL {
			fmt.Printf(" [%s]", plugin.URL)
		}
		fmt.Println()
	}
}
