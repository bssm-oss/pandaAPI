package main

import (
	"path/filepath"
	"testing"
)

func TestNormalizeProvider(t *testing.T) {
	provider, err := NormalizeProvider("Gemini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != ProviderGemini {
		t.Fatalf("unexpected provider: %q", provider)
	}
}

func TestCookieFilePath(t *testing.T) {
	path := CookieFilePath(Config{CookieDir: "/tmp/pandaapi"}, ProviderChatGPT)
	expected := filepath.Join("/tmp/pandaapi", "cookies_chatgpt.json")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}
