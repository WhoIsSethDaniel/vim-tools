package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func check(check []string) []string {
	missing := []string{}
	for _, item := range check {
		_, err := os.Stat(item)
		if err != nil {
			missing = append(missing, item)
		}
	}
	return missing
}

func dirsToCheck() []string {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		fmt.Fprint(os.Stderr, "XDG_CONFIG_HOME does not seem to be set\n")
		os.Exit(1)
	}
	dataHome, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok {
		fmt.Fprint(os.Stderr, "XDG_DATA_HOME does not seem to be set\n")
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine home directory: %s\n", err)
		os.Exit(1)
	}
	stateHome, ok := os.LookupEnv("XDG_STATE_HOME")
	if !ok {
		stateHome = filepath.Join(home, ".local", "state")
	}
	check := []string{
		filepath.Join(dataHome, "nvim"),
		filepath.Join(stateHome, "nvim"),
		filepath.Join(configHome, "nvim"),
		filepath.Join(home, ".cache", "nvim"),
		filepath.Join(dataHome, "nvim", "plugins"),
		filepath.Join(dataHome, "nvim", "sessions"),
		filepath.Join(stateHome, "nvim", "shada"),
		filepath.Join(configHome, "nvim", "lua"),
		filepath.Join(configHome, "nvim", "lua", "plugins"),
		filepath.Join(configHome, "nvim", "pack", "git-plugins", "opt"),
	}
	return check
}

func main() {
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}
	numPlugins := len(plugins)
	pluginsOnDisk := tools.PluginsOnDisk()
	numPluginsOnDisk := len(pluginsOnDisk)
	disabledPlugins := 0
	frozenPlugins := 0
	csPlugins := 0
	for _, plugin := range plugins {
		if plugin.IsDisabled() {
			disabledPlugins++
		}
		if plugin.IsColorscheme() {
			csPlugins++
		}
		if plugin.HasVersion() {
			frozenPlugins++
		}
	}

	fmt.Print("# of modules:\n")
	fmt.Printf("  total: %d\n", numPlugins)
	fmt.Printf("  on-disk: %d\n", numPluginsOnDisk)
	if numPluginsOnDisk != numPlugins {
		fmt.Print("    - ERROR total should equal on-disk\n")
	}
	for pluginName, pluginPath := range pluginsOnDisk {
		if _, ok := plugins[pluginName]; !ok {
			fmt.Printf("    - INSTALLED %s [%s]\n", pluginName, pluginPath)
		}
	}
	for _, plugin := range plugins {
		if _, ok := pluginsOnDisk[plugin.Name]; !ok {
			fmt.Printf("    - UNINSTALLED %s\n", plugin.Name)
		}
	}
	fmt.Printf("  colorscheme: %d\n", csPlugins)
	fmt.Printf("  disabled: %d\n", disabledPlugins)
	fmt.Printf("  frozen: %d\n", frozenPlugins)

	// check important dirs/files
	missingDirs := check(dirsToCheck())

	fmt.Print("\nsanity check:\n")
	fmt.Print("  checking dirs: ")
	if len(missingDirs) > 0 {
		for _, miss := range missingDirs {
			fmt.Printf("\n   %s is MISSING", miss)
		}
	} else {
		fmt.Print("all ok\n")
	}

	// print out unused config files
	unused := plugins.UnusedConfigFiles()
	if len(unused) > 0 {
		sort.Strings(unused)
		fmt.Print("\nunused config files:\n")
		for _, cf := range unused {
			fmt.Printf("  %s\n", cf)
		}
	}
}
