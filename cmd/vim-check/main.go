package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	var versionCheck, showBranch bool
	flag.BoolVar(&versionCheck, "v", false, "Check version of each installed plugin")
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

	runGit := func(pluginName string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = filepath.Join(tools.PluginDir(), pluginName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s: failed to run git: %w", pluginName, err)
		}
		return strings.TrimRight(string(out), "\n"), nil
	}

	var wg sync.WaitGroup
	toPrint := make(chan string)
	errPrint := make(chan error)
	defer close(toPrint)
	defer close(errPrint)

	for _, pluginName := range args {
		wg.Add(1)
		if versionCheck {
			go func(pluginName string) {
				defer wg.Done()
				out, err := runGit(pluginName, "rev-parse", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				toPrint <- fmt.Sprintf("%s %s", pluginName, out)
			}(pluginName)
		} else {
			go func(pluginName string) {
				defer wg.Done()
				plugin := plugins[pluginName]
				if _, err := os.Stat(filepath.Join(tools.PluginDir(), pluginName)); err != nil {
					var cmd *exec.Cmd
					if plugin.HasVersion() {
						cmd = exec.Command("git", "clone", "--branch", plugin.Version, plugin.URL, plugin.Name) //nolint:gosec // not a function
					} else {
						cmd = exec.Command("git", "clone", plugin.URL, plugin.Name) //nolint:gosec // not a function
					}
					cmd.Dir = tools.PluginDir()
					out, err := cmd.CombinedOutput()
					if err != nil {
						errPrint <- fmt.Errorf("%s: failed to run git: %s: %w", pluginName, strings.TrimRight(string(out), "\n"), err)
						return
					}
					toPrint <- fmt.Sprintf("CLONED %s", pluginName)
					return
				}
				var branch, symref string
				switch plugin.Version {
				case "":
					symref, err = runGit(pluginName, "symbolic-ref", "HEAD")
					if err != nil {
						errPrint <- err
						return
					}
					branch = path.Base(symref)
				default:
					branch = plugin.Version
				}
				lhead, err := runGit(pluginName, "rev-parse", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				rheadRefs, err := runGit(pluginName, "ls-remote", "--refs", plugin.URL, branch)
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
						errPrint <- fmt.Errorf("ERROR %s: no remote heads found (possible change of primary branch?)", pluginName)
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
					errPrint <- fmt.Errorf("failed to find remote head for %s", pluginName)
					return
				}
				outputString := pluginName
				if showBranch {
					outputString = fmt.Sprintf("%s [%s]", outputString, branch)
				}
				if lhead == rhead {
					toPrint <- fmt.Sprintf("OK %s", outputString)
				} else {
					if _, err := runGit(pluginName, "pull", "--rebase", plugin.URL, branch); err != nil {
						errPrint <- fmt.Errorf("ERROR %s", outputString)
					}
					if plugin.HasVersion() {
						if _, err := runGit(pluginName, "reset", "--hard", branch); err != nil {
							errPrint <- fmt.Errorf("ERROR %s", outputString)
						}
					}
					toPrint <- fmt.Sprintf("UPDATED %s", outputString)
				}
			}(pluginName)
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
