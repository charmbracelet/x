//go:build !windows
// +build !windows

// Package main demonstrates usage.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

var (
	rows int
	cols int
)

func init() {
	flag.IntVar(&rows, "rows", 24, "number of rows")
	flag.IntVar(&cols, "cols", 80, "number of columns")
	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [command] [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s -cols=80 -rows=24 [command] [args...]\n", os.Args[0])

		os.Exit(1)
	}

	newStdin := os.Stdin
	newStderr := os.Stderr

	winsize := pty.Winsize{
		Rows: uint16(rows), //nolint:gosec
		Cols: uint16(cols), //nolint:gosec
	}

	ptm1, pts1, err := pty.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating pty: %v\n", err)
		os.Exit(1)
	}
	if err := pty.Setsize(ptm1, &winsize); err != nil {
		fmt.Fprintf(os.Stderr, "error setting pty size: %v\n", err)
		os.Exit(1)
	}

	go io.Copy(os.Stdout, ptm1) //nolint:errcheck

	newStdout := os.Stdout

	ptm2, pts2, err := pty.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating pty: %v\n", err)
		os.Exit(1)
	}

	if err := pty.Setsize(ptm2, &winsize); err != nil {
		fmt.Fprintf(os.Stderr, "error setting pty size: %v\n", err)
		os.Exit(1)
	}

	go io.Copy(newStderr, ptm2) //nolint:errcheck

	if err := syscall.Dup2(int(newStdin.Fd()), int(os.Stdin.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "error duplicating stdin file descriptor: %v\n", err)
		os.Exit(1)
	}
	if err := syscall.Dup2(int(newStdout.Fd()), int(os.Stdout.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "error duplicating stdout file descriptor: %v\n", err)
		os.Exit(1)
	}

	n := flag.NFlag()
	c := exec.Command(os.Args[n+1], os.Args[n+2:]...) //nolint:gosec,noctx
	c.Stdout = pts1
	c.Stderr = pts2
	c.Stdin = newStdin

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running command: %v\n", err)
		os.Exit(1)
	}
}
