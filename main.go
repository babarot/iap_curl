package main

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"io"
	"log"
	"os"
)

const (
	// AppName is this tool name
	AppName = "iap_token"
	// Version is the version information of this tool
	Version = "0.1.0"
)

const help = `iap_token

Usage:


Flags:
  --help                  show help message
`

// CLI represents the attributes for command-line interface
type CLI struct {
	opt    option
	stdout io.Writer
	stderr io.Writer
}

type option struct {
	help bool
}

func main() {
	os.Exit(newCLI(os.Args[1:]).run())
}

func newCLI(args []string) CLI {
	logWriter, err := LogOutput()
	if err != nil {
		panic(err)
	}
	log.SetOutput(logWriter)

	var c CLI

	c.stdout = os.Stdout
	c.stderr = os.Stderr

	for _, arg := range args {
		switch arg {
		case "--help":
			c.opt.help = true
		}
	}

	return c
}

func (c CLI) exit(msg interface{}) int {
	switch m := msg.(type) {
	case int:
		return m
	case nil:
		return 0
	case string:
		fmt.Fprintf(c.stdout, "%s\n", m)
		return 0
	case error:
		fmt.Fprintf(c.stderr, "[ERROR] %s: %s\n", AppName, m.Error())
		return 1
	default:
		panic(msg)
	}
}

func (c CLI) run() int {
	if c.opt.help {
		return c.exit(help)
	}

	env, err := GetEnv()
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
	fmt.Fprintf(c.stdout, "%s", token)
	return 0
}

type Env struct {
	Credentials string `json:"GOOGLE_APPLICATION_CREDENTIALS"`
	ClientID    string `json:"IAP_CLIENT_ID"`
}

const (
	envCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	envClientID    = "IAP_CLIENT_ID"
	envCurlCommand = "IAP_CURL_BIN"
)

// GetEnv returns Env includes url
func GetEnv() (env Env, err error) {
	credentials := os.Getenv(envCredentials)
	clientID := os.Getenv(envClientID)
	if credentials == "" {
		credentials, _ = homedir.Expand(env.Credentials)
	}
	if clientID == "" {
		clientID = env.ClientID
	}
	if credentials == "" {
		return env, fmt.Errorf("%s is missing", envCredentials)
	}
	if clientID == "" {
		return env, fmt.Errorf("%s is missing", envClientID)
	}
	return Env{
		Credentials: credentials,
		ClientID:    clientID,
	}, nil
}
