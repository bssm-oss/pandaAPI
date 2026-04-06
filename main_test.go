package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParseAuthArgsRequiresProvider(t *testing.T) {
	_, err := parseAuthArgs(nil)
	if err == nil || !strings.Contains(err.Error(), "--provider is required") {
		t.Fatalf("expected provider requirement error, got %v", err)
	}
}

func TestParseAskArgsDefaultsProvider(t *testing.T) {
	query, provider, err := parseAskArgs([]string{"--query", "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "hello" {
		t.Fatalf("unexpected query: %q", query)
	}
	if provider != ProviderChatGPT {
		t.Fatalf("unexpected provider: %q", provider)
	}
}

func TestRunCLIHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runCLI([]string{"help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "pandaapi ask") {
		t.Fatalf("expected usage output, got %q", stdout.String())
	}
}

func TestRunCLIPassesAskToHandler(t *testing.T) {
	original := runAskFn
	defer func() { runAskFn = original }()

	called := false
	runAskFn = func(query string, provider string) error {
		called = true
		if query != "hello" || provider != ProviderGemini {
			return errors.New("unexpected arguments")
		}
		return nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runCLI([]string{"ask", "--query", "hello", "--provider", "gemini"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", exitCode, stderr.String())
	}
	if !called {
		t.Fatal("expected ask handler to be called")
	}
}

func TestRunCLIAuthRequiresProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runCLI([]string{"auth"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "--provider is required") {
		t.Fatalf("expected provider error in stderr, got %q", stderr.String())
	}
}
