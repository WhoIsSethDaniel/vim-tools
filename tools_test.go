package tools_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tools "github.com/WhoIsSethDaniel/vim-tools"
	"github.com/spf13/afero"
)

func createDisabledFile() {
	const disabledContent = `auto-session`
	afero.WriteFile(tools.Filesys, tools.DisabledFilePath(), []byte(disabledContent), 0644)
}

func createPluginsFile() {
	const pluginsContent = `git clone git@github.com:WhoIsSethDaniel/goldsmith.nvim
git clone https://github.com/rmagatti/auto-session`
	afero.WriteFile(tools.Filesys, tools.PluginsFilePath(), []byte(pluginsContent), 0644)
}

func prepareEnv(t *testing.T) {
	t.Helper()

	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	tools.Filesys = afero.NewMemMapFs()
	f := tools.Filesys
	f.MkdirAll(tools.MetadataDir(), 0755)
	f.MkdirAll(tools.PluginDir(), 0755)
	f.Create(filepath.Join(tools.ConfigFileDir(), "goldsmith-nvim.lua"))
	f.Create(filepath.Join(tools.ConfigFileDir(), "auto-session.lua"))
	f.Create(filepath.Join(tools.ConfigFileDir(), "nvim-lspconfig.lua"))
	f.Create(filepath.Join(tools.ConfigFileDir(), "someotherplugin-nvim.lua"))
	f.MkdirAll(tools.ConfigFileDir(), 0755)
}

func TestRebuildConfig(t *testing.T) {
	prepareEnv(t)

	plugins := tools.Plugins{
		"nvim-lspconfig": {
			Name: "nvim-lspconfig", URL: "https://github.com/neovim/nvim-lspconfig", Colorscheme: false, Enabled: true,
		},
		"goldsmith.nvim": {
			Name: "goldsmith.nvim", URL: "git@github.com:WhoIsSethDaniel/goldsmith.nvim", Colorscheme: false, Enabled: true,
		},
		"someotherplugin.nvim": {
			Name: "someotherplugin.nvim", URL: "git@github.com:WhoIsSethDaniel/someotherplugin.nvim", Colorscheme: false, Enabled: false,
		},
	}

	if err := plugins.RebuildConfig(); err != nil {
		t.Fatal(err)
	}

	t.Run("all.lua is ok", func(t *testing.T) {
		data, err := afero.ReadFile(tools.Filesys, tools.AllPluginsPath())
		if err != nil {
			t.Fatal(err)
		}
		want := "-- load plugins\nvim.cmd[[\npackadd! goldsmith.nvim\npackadd! nvim-lspconfig\n\" packadd! someotherplugin.nvim\n]]\n\n-- colorscheme\nrequire'colorscheme.current'\n\n-- config files\n-- require'plugins.goldsmith-nvim'\n-- require'plugins.nvim-lspconfig'\n-- require'plugins.someotherplugin-nvim'\n\nrequire'colorscheme.changes'\n"
		if !reflect.DeepEqual(string(data), want) {
			t.Errorf("got %#v, want %#v", string(data), want)
		}
	})
}

func TestWrite(t *testing.T) {
	prepareEnv(t)

	plugins := tools.Plugins{
		"nvim-lspconfig": {
			Name: "nvim-lspconfig", URL: "https://github.com/neovim/nvim-lspconfig", Colorscheme: false, Enabled: true,
		},
		"goldsmith.nvim": {
			Name: "goldsmith.nvim", URL: "git@github.com:WhoIsSethDaniel/goldsmith.nvim", Colorscheme: false, Enabled: true,
		},
		"someotherplugin.nvim": {
			Name: "someotherplugin.nvim", URL: "git@github.com:WhoIsSethDaniel/someotherplugin.nvim", Colorscheme: false, Enabled: false,
		},
	}

	if err := plugins.Write(); err != nil {
		t.Fatal(err)
	}

	t.Run("plugins file ok", func(t *testing.T) {
		data, err := afero.ReadFile(tools.Filesys, tools.PluginsFilePath())
		if err != nil {
			t.Fatal(err)
		}
		want := `git clone git@github.com:WhoIsSethDaniel/goldsmith.nvim
git clone https://github.com/neovim/nvim-lspconfig
git clone git@github.com:WhoIsSethDaniel/someotherplugin.nvim
`
		if !reflect.DeepEqual(string(data), want) {
			t.Errorf("got %#v, want %#v", string(data), want)
		}
	})

	t.Run("disabled file ok", func(t *testing.T) {
		data, err := afero.ReadFile(tools.Filesys, tools.DisabledFilePath())
		if err != nil {
			t.Fatal(err)
		}
		want := "someotherplugin.nvim\n"
		if !reflect.DeepEqual(string(data), want) {
			t.Errorf("got %#v, want %#v", string(data), want)
		}
	})
}

func TestRead(t *testing.T) {
	prepareEnv(t)

	t.Run("disabled file missing", func(t *testing.T) {
		_, err := tools.Read()
		if err == nil {
			t.Error("wanted an error but didn't get one")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("got %q, wanted %s", err.Error(), os.ErrNotExist)
		}
	})

	createDisabledFile()
	t.Run("all file missing", func(t *testing.T) {
		_, err := tools.Read()
		if err == nil {
			t.Error("wanted an error but didn't get one")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("got %q, wanted %s", err.Error(), os.ErrNotExist)
		}
	})

	createPluginsFile()
	t.Run("opens all files successfully", func(t *testing.T) {
		data, err := tools.Read()
		if err != nil {
			t.Fatal(err)
		}
		want := tools.Plugins{
			"auto-session": {
				Name:        "auto-session",
				URL:         "https://github.com/rmagatti/auto-session",
				Colorscheme: false,
				Enabled:     false,
			},
			"goldsmith.nvim": {
				Name:        "goldsmith.nvim",
				URL:         "git@github.com:WhoIsSethDaniel/goldsmith.nvim",
				Colorscheme: false,
				Enabled:     true,
			},
		}

		if !reflect.DeepEqual(data, want) {
			t.Errorf("got %#v, want %#v", data, want)
		}
	})
}

func TestCleanName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"plugin", "plugin"},
		{"plugin.nvim", "plugin-nvim"},
	}
	for _, tt := range tests {
		plugin := tools.Plugin{
			Name: tt.name,
		}
		if got := plugin.CleanName(); got != tt.want {
			t.Errorf("Plugin.CleanName() = %v, want %v", got, tt.want)
		}
	}
}

func TestAdd(t *testing.T) {
	plugins := tools.Plugins{}
	tests := []struct {
		url  string
		want tools.Plugin
	}{
		{
			"https://github.com/neovim/nvim-lspconfig",
			tools.Plugin{Name: "nvim-lspconfig", URL: "https://github.com/neovim/nvim-lspconfig", Colorscheme: false, Enabled: true},
		},
		{
			"git@github.com:WhoIsSethDaniel/goldsmith.nvim",
			tools.Plugin{Name: "goldsmith.nvim", URL: "git@github.com:WhoIsSethDaniel/goldsmith.nvim", Colorscheme: false, Enabled: true},
		},
	}
	for _, tt := range tests {
		if got := plugins.Add(tt.url); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Plugins.Add() = %v, want %v", got, tt.want)
		}
	}
}
