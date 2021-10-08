package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	tools "github.com/WhoIsSethDaniel/vim-tools"
)

func main() {
	plugins, err := tools.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read plugins file: %s", err)
		os.Exit(1)
	}

	// probably should use flags if this gets any more complicated
	versionCheck := false
	args := []string{}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			if arg != "-h" {
				fmt.Fprintf(os.Stderr, "Invalid option: %s\n", arg)
				fmt.Fprintf(os.Stderr, "usage: %s [-h] [plugin ...]\n", filepath.Base(os.Args[0]))
				os.Exit(1)
			}
			versionCheck = true
			continue
		}
		_, ok := plugins[arg]
		if !ok {
			fmt.Fprintf(os.Stderr, "No such plugin %s\n", arg)
			os.Exit(1)
		}
		args = append(args, arg)
	}

	if len(args) == 0 {
		args = plugins.SortedNames()
	}

	runGit := func(pluginName string, args ...string) (string, error) {
		cmd := exec.Command("git", args...)
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
				symref, err := runGit(pluginName, "symbolic-ref", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				branch := path.Base(symref)
				remote, err := runGit(pluginName, "config", fmt.Sprintf("branch.%s.remote", branch))
				if err != nil {
					errPrint <- err
					return
				}
				remoteURL, err := runGit(pluginName, "config", fmt.Sprintf("remote.%s.url", remote))
				if err != nil {
					errPrint <- err
					return
				}
				lhead, err := runGit(pluginName, "rev-parse", "HEAD")
				if err != nil {
					errPrint <- err
					return
				}
				rheadRefs, err := runGit(pluginName, "ls-remote", "--heads", remoteURL, branch)
				if err != nil {
					errPrint <- err
					return
				}
				var rhead string
				rheads := strings.Split(rheadRefs, "\n")
				for _, ref := range rheads {
					f := strings.Fields(ref)
					if f[1] == symref {
						rhead = f[0]
					}
				}
				if rhead == "" {
					errPrint <- fmt.Errorf("failed to find remote head for %s", pluginName)
					return
				}
				if lhead == rhead {
					toPrint <- fmt.Sprintf("OK %s", pluginName)
				} else {
					if _, err := runGit(pluginName, "pull", "--rebase", remoteURL, branch); err != nil {
						errPrint <- fmt.Errorf("ERROR %s", pluginName)
					} else {
						toPrint <- fmt.Sprintf("UPDATED %s", pluginName)
					}
				}
			}(pluginName)
		}
	}

	done := make(chan struct{})
	go func() {
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
}
