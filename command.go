package main

import (
	"os"
	"os/exec"
	"runtime"
)

func runCommand(binary string, args []string) error {
	// Check if you have curl command
	command := binary
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
