package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Filesys ....
var Filesys = afero.NewOsFs()

// Plugin ....
type Plugin struct {
	Name        string `json:"name"`
	URL         string `json:"url"` // not always a url
	CleanName   string `json:"clean_name"`
	ConfigFile  string `json:"config_file"`
	Colorscheme bool   `json:"colorscheme"`
	Enabled     bool   `json:"enabled"`
	Version     string `json:"version"`
}

// Plugins ....
type Plugins map[string]Plugin

// MetadataDir ....
func MetadataDir() string {
	dataHome, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok {
		panic("XDG_DATA_HOME must be set.")
	}
	return filepath.Join(dataHome, "nvim", "plugins")
}

// PluginDir ....
func PluginDir() string {
	configHome, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok {
		panic("XDG_DATA_HOME must be set.")
	}
	return filepath.Join(
		configHome,
		"nvim", "site", "pack", "core", "opt",
	)
}

// ConfigFileDir ....
func ConfigFileDir() string {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		panic("XDG_CONFIG_HOME must be set.")
	}
	return filepath.Join(configHome, "nvim", "lua", "plugins")
}

// PluginsFilePath ....
func PluginsFilePath() string {
	return filepath.Join(MetadataDir(), "all.json")
}

// AllPluginsPath ....
func AllPluginsPath() string {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		panic("XDG_CONFIG_HOME must be set.")
	}
	return filepath.Join(configHome, "nvim", "lua", "all.lua")
}

// PluginsOnDisk ....
func PluginsOnDisk() map[string]string {
	ent, err := os.ReadDir(PluginDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read plugin directory: %s\n", err)
		os.Exit(1)
	}
	pluginsOnDisk := make(map[string]string)
	for _, dir := range ent {
		if dir.IsDir() {
			pluginsOnDisk[dir.Name()] = filepath.Join(PluginDir(), dir.Name())
		}
	}
	return pluginsOnDisk
}

// RebuildConfig ....
func (p Plugins) RebuildConfig() error {
	names := p.SortedNames()

	allPluginsPath := AllPluginsPath()
	allLuaPlugins, _ := afero.TempFile(
		Filesys,
		filepath.Dir(allPluginsPath),
		filepath.Base(allPluginsPath),
	)
	defer allLuaPlugins.Close()
	defer Filesys.Remove(allLuaPlugins.Name())

	// fmt.Fprint(allLuaPlugins, "vim.pack.add({\n")
	fmt.Fprint(allLuaPlugins, "vim.cmd[[\n")
	for _, name := range names {
		plugin := p[name]
		if plugin.IsDisabled() {
			fmt.Fprintf(allLuaPlugins, "\" packadd! %s\n", plugin.Name)
			// fmt.Fprintf(allLuaPlugins, "\\%s\n", plugin.URL)
		} else {
			fmt.Fprintf(allLuaPlugins, "packadd! %s\n", plugin.Name)
			// fmt.Fprintf(allLuaPlugins, "  {\n    src = '%s',\n  },\n", plugin.URL)
		}
	}
	fmt.Fprint(allLuaPlugins, "]]\n\n")
	fmt.Fprint(allLuaPlugins, "-- colorscheme\n")
	// fmt.Fprint(allLuaPlugins, "})\n\n")
	// fmt.Fprint(allLuaPlugins, "-- colorscheme\n")
	for _, name := range names {
		plugin := p[name]
		if plugin.Colorscheme && plugin.IsEnabled() {
			if _, err := os.Stat(plugin.ConfigFilePath()); err == nil {
				fmt.Fprintf(allLuaPlugins, "require'plugins.%s'\n\n", plugin.CleanName)
				break // only allow the first enabled colorscheme
			}
		}
	}

	fmt.Fprint(allLuaPlugins, "-- config files\n")
	for _, name := range names {
		plugin := p[name]
		if _, err := os.Stat(plugin.ConfigFilePath()); err == nil {
			if plugin.IsDisabled() || plugin.Colorscheme {
				fmt.Fprintf(allLuaPlugins, "-- require'plugins.%s'\n", plugin.CleanName)
			} else {
				fmt.Fprintf(allLuaPlugins, "require'plugins.%s'\n", plugin.CleanName)
			}
		}
	}

	return Filesys.Rename(allLuaPlugins.Name(), allPluginsPath)
}

func ConfigsOnDisk() map[string]bool {
	ent, err := os.ReadDir(ConfigFileDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read plugin directory: %s\n", err)
		os.Exit(1)
	}
	configsOnDisk := make(map[string]bool)
	for _, f := range ent {
		if !f.IsDir() {
			configsOnDisk[f.Name()] = true
		}
	}
	return configsOnDisk
}

func (p Plugins) UnusedConfigFiles() []string {
	names := p.SortedNames()
	onDisk := ConfigsOnDisk()

	for _, name := range names {
		cf := p[name].ConfigFile
		delete(onDisk, cf)
	}
	unused := []string{}
	for cf := range onDisk {
		unused = append(unused, cf)
	}
	return unused
}

// SortedNames ....
func (p Plugins) SortedNames() []string {
	sortedPlugins := make([]string, 0, len(p))
	for name := range p {
		sortedPlugins = append(sortedPlugins, name)
	}
	sort.Strings(sortedPlugins)
	return sortedPlugins
}

// Add ....
func (p Plugins) Add(url, name, version string) Plugin {
	if name == "" {
		name = filepath.Base(strings.Split(url, ":")[1])
	}
	plugin := Plugin{
		Name:        name,
		URL:         url,
		Enabled:     true,
		Colorscheme: false,
		Version:     version,
	}
	plugin.CleanName = strings.Map(func(c rune) rune {
		if c == '.' {
			return '-'
		}
		return c
	}, name)
	plugin.ConfigFile = fmt.Sprintf("%s.lua", plugin.CleanName)

	_, err := Filesys.Stat(filepath.Join(PluginDir(), name, "colors"))
	if !errors.Is(err, fs.ErrNotExist) {
		plugin.Colorscheme = true
	}
	p[name] = plugin
	return p[name]
}

// Remove ....
func (p Plugins) Remove(plugin Plugin) {
	delete(p, plugin.Name)
}

// Read ....
func Read() (Plugins, error) {
	pf, err := Filesys.Open(PluginsFilePath())
	if err != nil {
		return nil, fmt.Errorf("failed to open plugins file: %w", err)
	}
	pfjson, err := afero.ReadAll(pf)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins file json: %w", err)
	}
	plugins := Plugins{}
	if err := json.Unmarshal(pfjson, &plugins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugins json: %w", err)
	}
	return plugins, nil
}

// Write ....
func (p Plugins) Write() error {
	pjson, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("conversion to JSON failed: %w", err)
	}
	pf, err := afero.TempFile(Filesys, MetadataDir(), filepath.Base(PluginsFilePath()))
	defer pf.Close()
	if err != nil {
		return fmt.Errorf("failed to create temp file for plugins file: %w", err)
	}
	fmt.Fprintf(pf, "%s", pjson)
	if err := Filesys.Rename(pf.Name(), PluginsFilePath()); err != nil {
		return fmt.Errorf("rename of plugins file failed: %w", err)
	}
	return nil
}

func (plugin *Plugin) CloneRepo() (string, error) {
	return plugin.runGitFromDir(PluginDir(), "clone", plugin.URL)
}

func (plugin *Plugin) RunGit(args ...string) (string, error) {
	return plugin.runGitFromDir(filepath.Join(PluginDir(), plugin.Name), args...)
}

func (plugin *Plugin) runGitFromDir(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: failed to run git: %w", plugin.Name, err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// ConfigFilePath ....
func (plugin Plugin) ConfigFilePath() string {
	return filepath.Join(ConfigFileDir(), plugin.ConfigFile)
}

// Freeze ....
func (plugin Plugin) Freeze(version string) Plugin {
	plugin.Version = version
	return plugin
}

// Thaw
func (plugin Plugin) Thaw() Plugin {
	plugin.Version = ""
	return plugin
}

// Disable ....
func (plugin Plugin) Disable() Plugin {
	plugin.Enabled = false
	return plugin
}

// Enable ....
func (plugin Plugin) Enable() Plugin {
	plugin.Enabled = true
	return plugin
}

// IsEnabled ....
func (plugin Plugin) IsEnabled() bool {
	return plugin.Enabled
}

// IsDisabled ....
func (plugin Plugin) IsDisabled() bool {
	return !plugin.Enabled
}

// IsColorscheme ....
func (plugin Plugin) IsColorscheme() bool {
	return plugin.Colorscheme
}

func (plugin Plugin) HasVersion() bool {
	return plugin.Version != ""
}
