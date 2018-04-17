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

func (c CLI) exit(msg interface{}) int {
	switch m := msg.(type) {
	case string:
		fmt.Fprintf(c.stdout, "%s\n", m)
		return 0
	case error:
		fmt.Fprintf(c.stderr, "Error: %s\n", m.Error())
		return 1
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

	env, _ := c.cfg.GetEnv(url)
	i, err := getInfo(env)
	if err != nil {
		return c.exit(err)
	}

	iap, err := newIAP(i.credentials, i.clientID)
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

	return c.exit(runCommand(i.binary, args))
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
