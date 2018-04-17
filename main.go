package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CLI represents the attributes for command-line interface
type CLI struct {
	opt  option
	args []string
	urls []url.URL
	cfg  Config

	stdout io.Writer
	stderr io.Writer
}

type option struct {
	list bool
	edit bool
}

func main() {
	cli, err := newCLI(os.Args[1:])
	if err != nil {
		panic(err)
	}
	os.Exit(cli.run())
}

func newCLI(args []string) (CLI, error) {
	var c CLI

	// TODO: make it customizable
	c.stdout = os.Stdout
	c.stderr = os.Stderr

	for _, arg := range args {
		switch arg {
		case "--list", "--list-urls":
			c.opt.list = true
		case "--edit", "--edit-config":
			c.opt.edit = true
		default:
			u, err := url.Parse(arg)
			if err == nil {
				c.urls = append(c.urls, *u)
			} else {
				c.args = append(c.args, arg)
			}
		}
	}

	dir, _ := configDir()
	json := filepath.Join(dir, "config.json")
	if err := c.cfg.LoadFile(json); err != nil {
		return c, err
	}

	return c, nil
}

func (c CLI) exit(msg interface{}) int {
	switch m := msg.(type) {
	case string:
		fmt.Fprintf(c.stdout, "%s\n", m)
		return 0
	case error:
		fmt.Fprintf(c.stderr, "Error: %s\n", m.Error())
		return 1
	case int:
		return m
	case nil:
		return 0
	default:
		panic(msg)
	}
}

func (c CLI) run() int {
	if c.opt.list {
		return c.exit(strings.Join(c.cfg.GetURLs(), "\n"))
	}

	if c.opt.edit {
		return c.exit(c.cfg.Edit())
	}

	url := c.getURL()
	if url == "" {
		return c.exit(errors.New("invalid url or url not given"))
	}

	env, err := c.cfg.GetEnv(url)
	if err != nil {
		return c.exit(err)
	}

	iap, err := newIAP(env.Credentials, env.ClientID)
	if err != nil {
		return c.exit(err)
	}
	token, err := iap.GetToken()
	if err != nil {
		return c.exit(err)
	}

	authHeader := fmt.Sprintf("'Authorization: Bearer %s'", token)
	args := append(
		[]string{"-H", authHeader}, // For IAP header
		c.args..., // Original args
	)
	args = append(args, url)

	return c.exit(runCommand(env.Binary, args))
}

func (c CLI) getURL() string {
	if len(c.urls) == 0 {
		return ""
	}
	return c.urls[0].String()
}

func runCommand(command string, args []string) error {
	// Check if you have curl command
	if _, err := exec.LookPath(command); err != nil {
		return err
	}
	for _, arg := range args {
		command += " " + arg
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
