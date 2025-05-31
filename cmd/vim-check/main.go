package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var hashCheck, showBranch bool
	flag.BoolVar(&hashCheck, "hash", false, "Check hash of each installed plugin")
	flag.BoolVar(&showBranch, "b", false, "Show the branch name that is being inspected")
	flag.Parse()

	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s\n", err)
		os.Exit(1)
	}

	var args []string
	if flag.NArg() == 0 {
		args = plugins.SortedNames()
	} else {
		args = flag.Args()
		for _, arg := range args {
			_, ok := plugins[arg]
			if !ok {
				fmt.Fprintf(os.Stderr, "No such plugin %s\n", arg)
				os.Exit(1)
			}
		}
	}

	var wg sync.WaitGroup
	toPrint := make(chan string)
	errPrint := make(chan error)
	defer close(toPrint)
	defer close(errPrint)

	for _, pluginName := range args {
		plugin := plugins[pluginName]
		wg.Add(1)
		if hashCheck {
			go func(plugin tools.Plugin) {
				defer wg.Done()
				out, err := plugin.RunGit("rev-parse", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				toPrint <- fmt.Sprintf("%s %s", plugin.Name, out)
			}(plugin)
		} else {
			go func(plugin tools.Plugin) {
				defer wg.Done()
				if _, err := os.Stat(filepath.Join(tools.PluginDir(), plugin.Name)); err != nil {
					out, err := plugin.CloneRepo()
					if err != nil {
						errPrint <- fmt.Errorf("%s: failed to clone repo: %s: %w", plugin.Name, strings.TrimRight(out, "\n"), err)
						return
					}
					if plugin.HasVersion() {
						if _, err := plugin.RunGit("reset", "--hard", plugin.Version); err != nil {
							errPrint <- fmt.Errorf("%s: failed to reset repo: %s: %w", plugin.Name, strings.TrimRight(out, "\n"), err)
							return
						}
					}
					toPrint <- fmt.Sprintf("CLONED %s", plugin.Name)
					return
				}
				var branch, symref string
				switch plugin.Version {
				case "":
					symref, err = plugin.RunGit("symbolic-ref", "HEAD")
					if err != nil {
						errPrint <- err
						return
					}
					branch = path.Base(symref)
				default:
					branch = plugin.Version
				}
				lhead, err := plugin.RunGit("rev-parse", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				rheadRefs, err := plugin.RunGit("ls-remote", "--refs", plugin.URL, branch)
				if err != nil {
					errPrint <- err
					return
				}
				// this logic is not always correct.
				var rhead string
				rheads := strings.Split(rheadRefs, "\n")
				for _, ref := range rheads {
					f := strings.Fields(ref)
					if len(f) == 0 {
						errPrint <- fmt.Errorf("ERROR %s: no remote heads found (possible change of primary branch?)", plugin.Name)
						return
					}
					if symref != "" {
						if f[1] == symref {
							rhead = f[0]
							break
						}
					} else {
						rhead = f[0]
					}
				}
				if rhead == "" {
					errPrint <- fmt.Errorf("failed to find remote head for %s", plugin.Name)
					return
				}
				outputString := plugin.Name
				if showBranch {
					outputString = fmt.Sprintf("%s [%s]", outputString, branch)
				}
				if lhead == rhead {
					toPrint <- fmt.Sprintf("OK %s", outputString)
				} else {
					if _, err := plugin.RunGit("pull", "--rebase", plugin.URL, branch); err != nil {
						errPrint <- fmt.Errorf("ERROR %s", outputString)
					}
					if plugin.HasVersion() {
						if _, err := plugin.RunGit("reset", "--hard", branch); err != nil {
							errPrint <- fmt.Errorf("ERROR %s", outputString)
						}
					}
					toPrint <- fmt.Sprintf("UPDATED %s", outputString)
				}
			}(plugin)
		}
	}

	done := make(chan struct{})
	go func() {
		// remove plugins that are no longer being used
		for pluginName, pluginPath := range tools.PluginsOnDisk() {
			if _, ok := plugins[pluginName]; !ok {
				fmt.Printf("DELETE %s\n", pluginName)
				os.RemoveAll(pluginPath)
			}
		}

		defer close(done)
		wg.Wait()
	}()

done:
	for {
		select {
		case <-done:
			break done
		case txt := <-toPrint:
			fmt.Print(txt + "\n")
		case e := <-errPrint:
			fmt.Fprintf(os.Stderr, "%s\n", e)
		}
	}

	if err := plugins.RebuildConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rebuild configuration: %s\n", err)
	}
}
