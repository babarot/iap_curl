package main

import (
	"io"
	"os"
	"os/exec"
	"runtime"

	colorable "github.com/mattn/go-colorable"
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
	// cmd.Stderr = os.Stderr
	cmd.Stderr = &colorWriter{colorable.NewColorable(os.Stderr)}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type newlineWriter struct {
	w io.Writer
}

type colorWriter struct {
	w io.Writer
}

func (w *colorWriter) Write(p []byte) (int, error) {
	n := len(p)
	var b []byte
	red := []byte("\033[31m")
	clear := []byte("\033[m")
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '>':
			b = append(b, red...)
			b = append(b, p[i])
		case '\n':
			b = append(b, clear...)
			b = append(b, p[i])
		default:
			b = append(b, p[i])
		}
	}
	p = b

	_, err := w.w.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}
