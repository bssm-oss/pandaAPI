package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAskRequiresExistingCookieFile(t *testing.T) {
	t.Setenv("PANDAAPI_COOKIE_DIR", t.TempDir())
	err := RunAsk("hello", ProviderChatGPT)
	if err == nil || !strings.Contains(err.Error(), "Run 'pandaapi auth' first") {
		t.Fatalf("expected missing-cookie error, got %v", err)
	}
}

func TestRunAskRejectsBlankQuery(t *testing.T) {
	err := RunAsk("   ", ProviderChatGPT)
	if err == nil || !strings.Contains(err.Error(), "--query is required") {
		t.Fatalf("expected blank query error, got %v", err)
	}
}

func TestRunAskRejectsUnsupportedProvider(t *testing.T) {
	t.Setenv("PANDAAPI_COOKIE_DIR", t.TempDir())
	err := RunAsk("hello", "unknown")
	if err == nil || !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("expected unsupported provider error, got %v", err)
	}
}

func TestRunAskWrapsCookieStatErrors(t *testing.T) {
	tmpDir := t.TempDir()
	blockedPath := filepath.Join(tmpDir, "file-not-dir")
	if err := os.WriteFile(blockedPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}
	t.Setenv("PANDAAPI_COOKIE_DIR", blockedPath)
	err := RunAsk("hello", ProviderChatGPT)
	if err == nil || !strings.Contains(err.Error(), "create cookie directory") {
		t.Fatalf("expected config creation error, got %v", err)
	}
}
