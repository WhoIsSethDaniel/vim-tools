package tools

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

// Filesys ....
var Filesys = afero.NewOsFs()

// Plugin ....
type Plugin struct {
	Name        string
	URL         string // not always a url
	Colorscheme bool
	Enabled     bool
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
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		panic("XDG_CONFIG_HOME must be set.")
	}
	return filepath.Join(configHome, "nvim/pack/git-plugins/opt") //nolint:gocritic // use of / is fine
}

// ConfigFileDir ....
func ConfigFileDir() string {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		panic("XDG_CONFIG_HOME must be set.")
	}
	return filepath.Join(configHome, "nvim/lua/plugins") //nolint:gocritic // use of / is fine
}

// DisabledFilePath ....
func DisabledFilePath() string {
	return filepath.Join(MetadataDir(), "disabled")
}

// PluginsFilePath ....
func PluginsFilePath() string {
	return filepath.Join(MetadataDir(), "all")
}

// AllPluginsPath ....
func AllPluginsPath() string {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		panic("XDG_CONFIG_HOME must be set.")
	}
	return filepath.Join(configHome, "nvim/lua/all.lua") //nolint:gocritic // use of / is fine
}

func readPluginsFile(file string, disabled []string) (Plugins, error) {
	p := Plugins{}
	f, err := Filesys.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load plugin file: %w", err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		elem := strings.Split(scanner.Text(), " ")
		if len(elem) < 3 {
			continue
		}
		p.Add(elem[2])
	}
	for _, name := range disabled {
		pt := p[name]
		pt.Enabled = false
		p[name] = pt
	}
	return p, nil
}

func readDisabledFile(file string) ([]string, error) {
	disabled := []string{}
	f, err := Filesys.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load disabled file: %w", err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		disabled = append(disabled, scanner.Text())
	}
	return disabled, nil
}

// RebuildConfig ....
func (p Plugins) RebuildConfig() error {
	names := p.SortedNames()

	allPluginsPath := AllPluginsPath()
	allLuaPlugins, _ := afero.TempFile(Filesys, filepath.Dir(allPluginsPath), filepath.Base(allPluginsPath))
	defer allLuaPlugins.Close()
	defer os.Remove(allLuaPlugins.Name())

	fmt.Fprint(allLuaPlugins, "-- load plugins\n")
	fmt.Fprint(allLuaPlugins, "vim.cmd[[\n")
	for _, name := range names {
		plugin := p[name]
		if plugin.IsDisabled() {
			fmt.Fprintf(allLuaPlugins, "\" packadd! %s\n", plugin.Name)
		} else {
			fmt.Fprintf(allLuaPlugins, "packadd! %s\n", plugin.Name)
		}
	}
	fmt.Fprint(allLuaPlugins, "]]\n")
	fmt.Fprint(allLuaPlugins, "\n-- colorscheme\n")
	fmt.Fprint(allLuaPlugins, "require'colorscheme.current'\n\n")

	fmt.Fprint(allLuaPlugins, "-- config files\n")
	for _, name := range names {
		plugin := p[name]
		fi, err := Filesys.Stat(plugin.ConfigFilePath())
		if err != nil {
			panic(fmt.Sprintf("Cannot stat config file for '%s': %s", name, err))
		}
		if fi.Size() == 0 || plugin.IsDisabled() {
			fmt.Fprintf(allLuaPlugins, "-- require'plugins.%s'\n", plugin.CleanName())
		} else {
			fmt.Fprintf(allLuaPlugins, "require'plugins.%s'\n", plugin.CleanName())
		}
	}

	fmt.Fprint(allLuaPlugins, "\nrequire'colorscheme.changes'\n")

	return Filesys.Rename(allLuaPlugins.Name(), allPluginsPath)
}

// SortedNames ....
func (p Plugins) SortedNames() []string {
	sortedPlugins := []string{}
	for name := range p {
		sortedPlugins = append(sortedPlugins, name)
	}
	sort.Strings(sortedPlugins)
	return sortedPlugins
}

// Read ....
func Read() (Plugins, error) {
	disabled, err := readDisabledFile(DisabledFilePath())
	if err != nil {
		return nil, err
	}

	p, err := readPluginsFile(PluginsFilePath(), disabled)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p Plugins) writePluginsFile(file string) error {
	tplugins, err := afero.TempFile(Filesys, filepath.Dir(file), filepath.Base(file))
	if err != nil {
		return fmt.Errorf("failed to create temp file for all plugins file: %w", err)
	}
	defer tplugins.Close()
	defer Filesys.Remove(tplugins.Name())

	sortedPlugins := p.SortedNames()
	for _, name := range sortedPlugins {
		plugin := p[name]
		fmt.Fprintf(tplugins, "git clone %s\n", plugin.URL)
	}

	if err := Filesys.Rename(tplugins.Name(), file); err != nil {
		return fmt.Errorf("failed to rename temp file for all plugins file: %w", err)
	}
	return nil
}

func (p Plugins) writeDisabledFile(file string) error {
	tdisabled, err := afero.TempFile(Filesys, filepath.Dir(file), filepath.Base(file))
	if err != nil {
		return fmt.Errorf("failed to create temp disabled file: %w", err)
	}
	defer tdisabled.Close()
	defer Filesys.Remove(tdisabled.Name())

	disabledPlugins := []string{}
	for name, plugin := range p {
		if !plugin.Enabled {
			disabledPlugins = append(disabledPlugins, name)
		}
	}

	sort.Strings(disabledPlugins)
	for _, name := range disabledPlugins {
		fmt.Fprint(tdisabled, name+"\n")
	}

	if err := Filesys.Rename(tdisabled.Name(), file); err != nil {
		return fmt.Errorf("failed to rename temp disabled file: %w", err)
	}
	return nil
}

// Add ....
func (p Plugins) Add(url string) Plugin {
	key := filepath.Base(strings.Split(url, ":")[1])
	_, err := os.Stat(filepath.Join(PluginDir(), key, "colors"))
	if errors.Is(err, fs.ErrNotExist) {
		p[key] = Plugin{Name: key, URL: url, Enabled: true, Colorscheme: false}
	} else {
		p[key] = Plugin{Name: key, URL: url, Enabled: true, Colorscheme: true}
	}
	return p[key]
}

// Remove ....
func (p Plugins) Remove(plugin Plugin) {
	delete(p, plugin.Name)
}

// Write ....
func (p Plugins) Write() error {
	if err := p.writePluginsFile(PluginsFilePath()); err != nil {
		return fmt.Errorf("failed to write plugins file: %w", err)
	}
	return p.writeDisabledFile(DisabledFilePath())
}

// ConfigFilePath ....
func (plugin Plugin) ConfigFilePath() string {
	return fmt.Sprintf("%s.lua", filepath.Join(ConfigFileDir(), plugin.CleanName()))
}

// CleanName ....
func (plugin Plugin) CleanName() string {
	return strings.Map(func(c rune) rune {
		if c == '.' {
			return '-'
		}
		return c
	}, plugin.Name)
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
