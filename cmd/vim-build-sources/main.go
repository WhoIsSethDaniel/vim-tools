package main

import (
	"fmt"
	"os"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugin information: %s", err)
		os.Exit(1)
	}
	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s", err)
		os.Exit(1)
	}
}
