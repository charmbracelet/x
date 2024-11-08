//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func listenForResize(fn func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	for range sig {
		fn()
	}
}
