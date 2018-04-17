package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	envCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	envClientID    = "IAP_CLIENT_ID"
	envCurlCommand = "IAP_CURL_BIN"
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
	var cli CLI

	// TODO: make it customizable
	cli.stdout = os.Stdout
	cli.stderr = os.Stderr

	for _, arg := range args {
		switch arg {
		case "--list", "--list-urls":
			cli.opt.list = true
		case "--edit", "--edit-config":
			cli.opt.edit = true
		default:
			u, err := url.Parse(arg)
			if err == nil {
				cli.urls = append(cli.urls, *u)
			} else {
				cli.args = append(cli.args, arg)
			}
		}
	}

	dir, _ := configDir()
	json := filepath.Join(dir, "config.json")
	if err := cli.cfg.LoadFile(json); err != nil {
		return cli, err
	}

	return cli, nil
}

func (c CLI) run() int {
	if c.opt.list {
		fmt.Fprintln(c.stdout, strings.Join(c.cfg.GetURLs(), "\n"))
		return 0
	}

	if c.opt.edit {
		if err := c.cfg.Edit(); err != nil {
			fmt.Fprintf(c.stderr, "Error: %v\n", err.Error())
			return 1
		}
		return 0
	}

	url := c.getURL()
	if url == "" {
		fmt.Fprintf(c.stderr, "invalid url or url not given\n")
		return 1
	}

	env, _ := c.cfg.GetEnv(url)
	i, err := getInfo(env)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %s\n", err.Error())
		return 1
	}

	iap, err := newIAP(i.credentials, i.clientID)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err.Error())
		return 1
	}
	token, err := iap.GetToken()
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err.Error())
		return 1
	}

	authHeader := fmt.Sprintf("'Authorization: Bearer %s'", token)
	args := append(
		[]string{"-H", authHeader}, // For IAP header
		c.args..., // Original args
	)
	args = append(args, url)

	if err := runCommand(i.binary, args); err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err.Error())
		return 1
	}

	return 0
}

func (c CLI) getURL() string {
	if len(c.urls) == 0 {
		return ""
	}
	return c.urls[0].String()
}

type info struct {
	credentials string
	clientID    string
	binary      string
}

func getInfo(env Env) (info, error) {
	credentials := os.Getenv(envCredentials)
	clientID := os.Getenv(envClientID)
	binary := os.Getenv(envCurlCommand)
	if credentials == "" {
		credentials, _ = homedir.Expand(env.Credentials)
	}
	if clientID == "" {
		clientID = env.ClientID
	}
	if binary == "" {
		binary = env.Binary
	}
	if credentials == "" {
		return info{}, fmt.Errorf("%s is missing", envCredentials)
	}
	if clientID == "" {
		return info{}, fmt.Errorf("%s is missing", envClientID)
	}
	if binary == "" {
		binary = "curl"
	}
	return info{
		credentials: credentials,
		clientID:    clientID,
		binary:      binary,
	}, nil
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
