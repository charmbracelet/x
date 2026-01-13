package lsp

import (
	"context"
	"sync"
	"testing"
)

func TestProcessCloser_ConcurrentClose(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	stream, err := startServerProcess(t.Context(), config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stream.Close()
		}()
	}
	wg.Wait()
}

func TestProcessCloser_CloseAfterContextCancel(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	ctx, cancel := context.WithCancel(t.Context())
	stream, err := startServerProcess(ctx, config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	cancel()
	stream.Close()
}

func TestProcessCloser_ConcurrentCancelAndClose(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stream, err := startServerProcess(ctx, config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		cancel()
	}()
	go func() {
		defer wg.Done()
		stream.Close()
	}()
	wg.Wait()
}
