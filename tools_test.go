package tools_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tools "github.com/WhoIsSethDaniel/vim-tools"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func createInstalledPlugin(t *testing.T, f afero.Fs, name, clean string, colorscheme bool) {
	t.Helper()

	f.MkdirAll(filepath.Join(tools.PluginDir(), name), 0o755)
	if colorscheme {
		f.MkdirAll(filepath.Join(tools.PluginDir(), name, "colors"), 0o755)
	}
	f.Create(filepath.Join(tools.ConfigFileDir(), fmt.Sprintf("%s.lua", clean)))
}

func prepareEnv(t *testing.T) {
	t.Helper()

	t.Setenv("XDG_DATA_HOME", "_TEST_")
	t.Setenv("XDG_CONFIG_HOME", "_TEST_")

	tools.Filesys = afero.NewMemMapFs()
	f := tools.Filesys
	f.MkdirAll(tools.MetadataDir(), 0o755)
	f.MkdirAll(tools.PluginDir(), 0o755)
	f.MkdirAll(tools.ConfigFileDir(), 0o755)
}

func createPluginsFile(t *testing.T) {
	t.Helper()

	plugins := tools.Plugins{
		"plugin1.nvim": {
			Name:        "plugin1.nvim",
			URL:         "https://github.com/user/plugin1.nvim",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin1-nvim.lua",
			CleanName:   "plugin1-nvim",
		},
		"plugin-a": {
			Name:        "plugin-a",
			URL:         "git@github.com:SomeUser/plugin-a",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin-a.lua",
			CleanName:   "plugin-a",
		},
		"someotherplugin.nvim": {
			Name:        "someotherplugin.nvim",
			URL:         "git@github.com:SomeOtherUser/someotherplugin.nvim",
			Colorscheme: false,
			Enabled:     false,
			ConfigFile:  "someotherplugin-nvim.lua",
			CleanName:   "someotherplugin-nvim",
		},
		"colorscheme.nvim": {
			Name:        "colorscheme.nvim",
			URL:         "https://gitlab.com/user/colorscheme.nvim",
			Colorscheme: true,
			Enabled:     true,
			ConfigFile:  "colorscheme-nvim.lua",
			CleanName:   "colorscheme-nvim",
		},
	}

	if err := plugins.Write(); err != nil {
		t.Fatal(err)
	}
}

func TestAdd(t *testing.T) {
	prepareEnv(t)

	createInstalledPlugin(t, tools.Filesys, "plugin1.nvim", "plugin1-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "colorscheme.nvim", "colorscheme-nvim", true)

	plugins := tools.Plugins{}
	tests := []struct {
		url  string
		want tools.Plugin
	}{
		{
			"https://github.com/user/plugin1.nvim",
			tools.Plugin{
				Name:        "plugin1.nvim",
				URL:         "https://github.com/user/plugin1.nvim",
				Colorscheme: false,
				Enabled:     true,
				ConfigFile:  "plugin1-nvim.lua",
				CleanName:   "plugin1-nvim",
			},
		},
		{
			"git@github.com:SomeUser/colorscheme.nvim",
			tools.Plugin{
				Name:        "colorscheme.nvim",
				URL:         "git@github.com:SomeUser/colorscheme.nvim",
				Colorscheme: true,
				Enabled:     true,
				ConfigFile:  "colorscheme-nvim.lua",
				CleanName:   "colorscheme-nvim",
			},
		},
	}
	for _, tt := range tests {
		if got := plugins.Add(tt.url); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Plugins.Add() = %v, want %v", got, tt.want)
		}
	}
}

func TestWrite(t *testing.T) {
	prepareEnv(t)

	createInstalledPlugin(t, tools.Filesys, "plugin1.nvim", "plugin1-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "plugin-a", "plugin-a", false)
	createInstalledPlugin(t, tools.Filesys, "plugin-b", "plugin-b", false)
	createInstalledPlugin(t, tools.Filesys, "someotherplugin.nvim", "someotherplugin-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "colorscheme.nvim", "colorscheme-nvim", true)

	plugins := tools.Plugins{
		"plugin1.nvim": {
			Name:        "plugin1.nvim",
			URL:         "https://github.com/user/plugin1.nvim",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin1-nvim.lua",
			CleanName:   "plugin1-nvim",
		},
		"plugin-a": {
			Name:        "plugin-a",
			URL:         "git@github.com:SomeUser/plugin-a",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin-a.lua",
			CleanName:   "plugin-a",
		},
		"someotherplugin.nvim": {
			Name:        "someotherplugin.nvim",
			URL:         "git@github.com:SomeOtherUser/someotherplugin.nvim",
			Colorscheme: false,
			Enabled:     false,
			ConfigFile:  "someotherplugin-nvim.lua",
			CleanName:   "someotherplugin-nvim",
		},
		"colorscheme.nvim": {
			Name:        "colorscheme.nvim",
			URL:         "https://gitlab.com/user/colorscheme.nvim",
			Colorscheme: true,
			Enabled:     true,
			ConfigFile:  "colorscheme-nvim.lua",
			CleanName:   "colorscheme-nvim",
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
		want := `{
    "colorscheme.nvim": {
    "name": "colorscheme.nvim",
    "url": "https://gitlab.com/user/colorscheme.nvim",
    "colorscheme": true,
    "enabled": true,
	"config_file":  "colorscheme-nvim.lua",
	"clean_name":   "colorscheme-nvim"
  },
  "plugin-a": {
    "name": "plugin-a",
    "url": "git@github.com:SomeUser/plugin-a",
    "colorscheme": false,
    "enabled": true,
	"config_file":  "plugin-a.lua",
	"clean_name":   "plugin-a"
  },
  "plugin1.nvim": {
    "name": "plugin1.nvim",
    "url": "https://github.com/user/plugin1.nvim",
    "colorscheme": false,
    "enabled": true,
	"config_file":  "plugin1-nvim.lua",
	"clean_name":   "plugin1-nvim"
  },
  "someotherplugin.nvim": {
    "name": "someotherplugin.nvim",
    "url": "git@github.com:SomeOtherUser/someotherplugin.nvim",
    "colorscheme": false,
    "enabled": false,
	"config_file":  "someotherplugin-nvim.lua",
	"clean_name":   "someotherplugin-nvim"
  }
}`
		require.JSONEqf(t, want, string(data), "plugins file")
	})
}

func TestRead(t *testing.T) {
	prepareEnv(t)

	createInstalledPlugin(t, tools.Filesys, "plugin1.nvim", "plugin1-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "plugin-a", "plugin-a", false)

	t.Run("all.json file missing", func(t *testing.T) {
		_, err := tools.Read()
		if err == nil {
			t.Error("wanted an error but didn't get one")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("got %q, wanted %s", err.Error(), os.ErrNotExist)
		}
	})

	createPluginsFile(t)
	t.Run("opens plugins file successfully", func(t *testing.T) {
		want := tools.Plugins{
			"plugin1.nvim": {
				Name:        "plugin1.nvim",
				URL:         "https://github.com/user/plugin1.nvim",
				Colorscheme: false,
				Enabled:     true,
				ConfigFile:  "plugin1-nvim.lua",
				CleanName:   "plugin1-nvim",
			},
			"plugin-a": {
				Name:        "plugin-a",
				URL:         "git@github.com:SomeUser/plugin-a",
				Colorscheme: false,
				Enabled:     true,
				ConfigFile:  "plugin-a.lua",
				CleanName:   "plugin-a",
			},
			"someotherplugin.nvim": {
				Name:        "someotherplugin.nvim",
				URL:         "git@github.com:SomeOtherUser/someotherplugin.nvim",
				Colorscheme: false,
				Enabled:     false,
				ConfigFile:  "someotherplugin-nvim.lua",
				CleanName:   "someotherplugin-nvim",
			},
			"colorscheme.nvim": {
				Name:        "colorscheme.nvim",
				URL:         "https://gitlab.com/user/colorscheme.nvim",
				Colorscheme: true,
				Enabled:     true,
				ConfigFile:  "colorscheme-nvim.lua",
				CleanName:   "colorscheme-nvim",
			},
		}
		data, err := tools.Read()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(data, want) {
			t.Errorf("got %#v, want %#v", data, want)
		}
	})
}

func TestRebuildConfig(t *testing.T) {
	prepareEnv(t)

	createInstalledPlugin(t, tools.Filesys, "plugin1.nvim", "plugin1-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "plugin-a", "plugin-a", false)
	createInstalledPlugin(t, tools.Filesys, "plugin-b", "plugin-b", false)
	createInstalledPlugin(t, tools.Filesys, "someotherplugin.nvim", "someotherplugin-nvim", false)
	createInstalledPlugin(t, tools.Filesys, "colorscheme.nvim", "colorscheme-nvim", true)

	plugins := tools.Plugins{
		"plugin1.nvim": {
			Name:        "plugin1.nvim",
			URL:         "https://github.com/user/plugin1.nvim",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin1-nvim.lua",
			CleanName:   "plugin1-nvim",
		},
		"plugin-a": {
			Name:        "plugin-a",
			URL:         "git@github.com:SomeUser/plugin-a",
			Colorscheme: false,
			Enabled:     true,
			ConfigFile:  "plugin-a.lua",
			CleanName:   "plugin-a",
		},
		"someotherplugin.nvim": {
			Name:        "someotherplugin.nvim",
			URL:         "git@github.com:SomeOtherUser/someotherplugin.nvim",
			Colorscheme: false,
			Enabled:     false,
			ConfigFile:  "someotherplugin-nvim.lua",
			CleanName:   "someotherplugin-nvim",
		},
		"colorscheme.nvim": {
			Name:        "colorscheme.nvim",
			URL:         "https://gitlab.com/user/colorscheme.nvim",
			Colorscheme: true,
			Enabled:     true,
			ConfigFile:  "colorscheme-nvim.lua",
			CleanName:   "colorscheme-nvim",
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
		want := "-- load plugins\nvim.cmd[[\npackadd! colorscheme.nvim\npackadd! plugin-a\npackadd! plugin1.nvim\n\" packadd! someotherplugin.nvim\n]]\n\n-- colorscheme\nrequire'colorscheme'\n\n-- config files\nrequire'plugins.colorscheme-nvim'\nrequire'plugins.plugin-a'\nrequire'plugins.plugin1-nvim'\n-- require'plugins.someotherplugin-nvim'\n"
		if !reflect.DeepEqual(string(data), want) {
			t.Errorf("got %#v, want %#v", string(data), want)
		}
	})
}
