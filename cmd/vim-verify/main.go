package main

import (
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
		fmt.Fprint(os.Stderr, "XDG_CONFIG_HOME does no seem to be set")
		os.Exit(1)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine home directory: %s", err)
		os.Exit(1)
	}
	check := []string{}
	check = append(check,
		filepath.Join(dataHome, "nvim"),
		filepath.Join(dataHome, "nvim/lua"),
		filepath.Join(dataHome, "nvim/lua/plugins"),
		filepath.Join(dataHome, "nvim/pack/git-plugins/opt"),
		filepath.Join(home, ".cache/nvim"))
	return check
}

func filesToCheck(extra []string) []string {
	check := []string{
		tools.PluginsFilePath(),
	}
	check = append(check, extra...)
	return check
}

func pluginsOnDisk() int {
	cnt := 0
	ent, err := os.ReadDir(tools.PluginDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read plugin directory: %s", err)
		os.Exit(1)
	}
	for _, dir := range ent {
		if dir.IsDir() {
			cnt++
		}
	}
	return cnt
}

func main() {
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s", err)
		os.Exit(1)
	}
	numPlugins := len(plugins)
	numPluginsOnDisk := pluginsOnDisk()
	disabledPlugins := 0
	csPlugins := 0
	pluginConfigFiles := []string{}
	for _, plugin := range plugins {
		if plugin.IsDisabled() {
			disabledPlugins++
		} else if plugin.IsColorscheme() {
			csPlugins++
		}
		pluginConfigFiles = append(pluginConfigFiles, plugin.ConfigFilePath())
	}

	fmt.Print("# of modules:\n")
	fmt.Printf("  total: %d\n", numPlugins)
	fmt.Printf("  on-disk: %d\n", numPluginsOnDisk)
	if numPluginsOnDisk != numPlugins {
		fmt.Print("    ERROR total should equal on-disk\n")
	}
	fmt.Printf("  colorscheme: %d\n", csPlugins)
	fmt.Printf("  disabled: %d\n", disabledPlugins)

	// check important dirs/files
	missingFiles := check(filesToCheck(pluginConfigFiles))
	missingDirs := check(dirsToCheck())

	fmt.Print("\nsanity check:\n")
	fmt.Print("  checking files: ")
	if len(missingFiles) > 0 {
		for _, miss := range missingFiles {
			fmt.Printf("\n    %s is MISSING", miss)
		}
	} else {
		fmt.Print("all ok")
	}

	fmt.Print("\n  checking dirs: ")
	if len(missingDirs) > 0 {
		for _, miss := range missingDirs {
			fmt.Printf("\n   %s is MISSING", miss)
		}
	} else {
		fmt.Print("all ok\n")
	}
}
