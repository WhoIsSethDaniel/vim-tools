package main

import (
	"flag"
	"fmt"
	"os"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var listURL bool
	flag.BoolVar(&listURL, "u", false, "List the repo URL along with the name")
	flag.Parse()

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read all plugin file: %s\n", err)
		os.Exit(1)
	}
	pluginNames := plugins.SortedNames()

	filterAndPrint := func(f func(p tools.Plugin) bool) {
		for _, name := range pluginNames {
			plugin := plugins[name]
			if f(plugin) {
				fmt.Print(name)
				if plugin.Frozen {
					fmt.Print(" [*]")
				}
				if listURL {
					fmt.Printf(" [%s]", plugin.URL)
				}
				fmt.Println()
			}
		}
	}

	filterAndPrint(func(plugin tools.Plugin) bool {
		return !plugin.IsColorscheme() && !plugin.IsDisabled()
	})

	fmt.Println("\n" + "colorschemes:")
	filterAndPrint(func(plugin tools.Plugin) bool {
		return plugin.IsColorscheme()
	})

	fmt.Println("\n" + "disabled:")
	filterAndPrint(func(plugin tools.Plugin) bool {
		return plugin.IsDisabled()
	})
}
