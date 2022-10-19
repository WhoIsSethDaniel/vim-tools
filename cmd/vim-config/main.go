package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var create, edit bool
	flag.BoolVar(&create, "c", false, "Create the config file for the given plugin(s)")
	flag.BoolVar(&edit, "e", false, "Edit the config file(s) for the given plugin(s)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-ce] plugin [plugin ...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}

	var configs []string
	for _, arg := range flag.Args() {
		plugin, ok := plugins[arg]
		if !ok {
			fmt.Fprintf(os.Stderr, "cannot find %s\n", arg)
			continue
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
		configs = append(configs, plugin.ConfigFilePath())
	}

	if err := plugins.Write(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s\n", err)
		os.Exit(1)
	}

	if edit {
		for _, config := range configs {
			if _, err := os.Stat(config); err != nil {
				fmt.Fprintf(os.Stderr, "Config file '%s' does not exist\n", config)
				os.Exit(1)
			}
		}
		cmd, err := exec.LookPath("sensible-editor")
		if err != nil {
			fmt.Printf("Failed to find path for 'sensible-editor': %s\n", err)
			os.Exit(1)
		}
		if err := syscall.Exec(cmd, append([]string{cmd}, configs...), os.Environ()); err != nil {
			fmt.Printf("Failed to exec 'sensible-editor': %s\n", err)
			os.Exit(1)
		}
	} else {
		for _, config := range configs {
			fmt.Printf("%s\n", config)
		}
	}
}
