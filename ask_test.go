package main

import (
	"errors"
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

func TestDetermineAskRuntime(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		cookieExists  bool
		headless      bool
		requireCookie bool
	}{
		{name: "chatgpt without cookie", provider: ProviderChatGPT, cookieExists: false, headless: true, requireCookie: true},
		{name: "gemini without cookie", provider: ProviderGemini, cookieExists: false, headless: false, requireCookie: false},
		{name: "gemini with cookie", provider: ProviderGemini, cookieExists: true, headless: true, requireCookie: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := determineAskRuntime(tc.provider, tc.cookieExists)
			if plan.headless != tc.headless || plan.requireCookie != tc.requireCookie {
				t.Fatalf("unexpected plan: %+v", plan)
			}
		})
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

func TestShouldFallbackGeminiSignedOut(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "auth required", err: errors.New("Run 'pandaapi auth' first"), want: true},
		{name: "unknown method", err: errors.New("submit Gemini prompt: UnknownMethod (-31998)"), want: true},
		{name: "other error", err: errors.New("wait for Gemini answer: context deadline exceeded"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldFallbackGeminiSignedOut(tc.err)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
