package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	TokenURI = "https://www.googleapis.com/oauth2/v4/token"

	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	IAPClientID                  = "IAP_CLIENT_ID"
	IAPCurlBinary                = "IAP_CURL_BIN"
)

const helpText string = `Usage: curl

Extended options:
  --list-urls    List service URLs
  --edit-config  Edit config file
`

var (
	credentials string
	clientID    string
	binary      string

	cfg Config
)

func main() {
	dir, _ := configDir()
	json := filepath.Join(dir, "config.json")

	err := cfg.LoadFile(json)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	credentials = os.Getenv(GoogleApplicationCredentials)
	clientID = os.Getenv(IAPClientID)
	binary = os.Getenv(IAPCurlBinary)

	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			fmt.Fprint(os.Stderr, helpText)
			return 1
		case "--list-urls":
			fmt.Println(strings.Join(cfg.GetURLs(), "\n"))
			return 0
		case "--edit-config":
			err := cfg.Edit()
			if err == nil {
				return 0
			}
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return 1
		}

		env, err := cfg.GetEnv(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return 1
		}
		if credentials == "" {
			credentials = env.Credentials
		}
		if clientID == "" {
			clientID = env.ClientID
		}
		if binary == "" {
			binary = env.Binary
		}
	}

	if credentials == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is missing\n", GoogleApplicationCredentials)
		return 1
	}
	if clientID == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is missing\n", IAPClientID)
		return 1
	}
	if binary == "" {
		binary = "curl"
	}

	token, err := getToken(credentials, clientID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		return 1
	}

	authHeader := fmt.Sprintf("'Authorization: Bearer %s'", token)
	curlArgs := append(
		// For IAP header
		[]string{"-H", authHeader},
		// Original args
		args...,
	)

	if err := doCurl(binary, curlArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		return 1
	}

	return 0
}
