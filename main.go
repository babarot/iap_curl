package main

import (
	"fmt"
	"os"
)

const (
	TokenURI = "https://www.googleapis.com/oauth2/v4/token"

	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	IAPClientID                  = "IAP_CLIENT_ID"
)

var helpText string = `Usage: curl
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	var (
		creds    = os.Getenv(GoogleApplicationCredentials)
		clientID = os.Getenv(IAPClientID)
	)

	if len(args) > 0 {
		if args[0] == "-h" || args[0] == "--help" {
			fmt.Fprint(os.Stderr, helpText)
			return 1
		}
	}

	if creds == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is missing\n", GoogleApplicationCredentials)
		return 1
	}

	if clientID == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is missing\n", IAPClientID)
		return 1
	}

	token, err := getToken(creds, clientID)
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

	if err := doCurl(curlArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		return 1
	}

	return 0
}
