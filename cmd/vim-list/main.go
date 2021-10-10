package main

import (
	"fmt"
	"os"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read all plugin file: %s\n", err)
		os.Exit(1)
	}
	pluginNames := plugins.SortedNames()

	filterAndPrint := func(f func(p tools.Plugin) bool) {
		for _, name := range pluginNames {
			if f(plugins[name]) {
				fmt.Println(name)
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
