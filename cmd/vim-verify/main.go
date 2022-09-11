package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	dataHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		fmt.Fprint(os.Stderr, "XDG_CONFIG_HOME does no seem to be set\n")
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine home directory: %s\n", err)
		os.Exit(1)
	}
	check := []string{
		filepath.Join(dataHome, "nvim"),
		filepath.Join(dataHome, "nvim", "lua"),
		filepath.Join(dataHome, "nvim", "lua", "plugins"),
		filepath.Join(dataHome, "nvim", "pack", "git-plugins", "opt"),
		filepath.Join(home, ".cache", "nvim"),
	}
	return check
}

func pluginsOnDisk() map[string]string {
	ent, err := os.ReadDir(tools.PluginDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read plugin directory: %s\n", err)
		os.Exit(1)
	}
	pluginsOnDisk := make(map[string]string)
	for _, dir := range ent {
		if dir.IsDir() {
			pluginsOnDisk[dir.Name()] = filepath.Join(tools.PluginDir(), dir.Name())
		}
	}
	return pluginsOnDisk
}

func main() {
	delUnknown := flag.Bool(
		"d",
		false,
		"Delete directories in the plugin directory that should not be there",
	)
	flag.Parse()
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}
	numPlugins := len(plugins)
	pluginsOnDisk := pluginsOnDisk()
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
		if plugin.IsFrozen() {
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
			fmt.Printf("    - INSTALLED %s [%s]", pluginName, pluginPath)
			if *delUnknown {
				fmt.Print("...REMOVING")
				os.RemoveAll(pluginPath)
			}
			fmt.Print("\n")
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
}
