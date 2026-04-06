package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoredCookiesRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "cookies.json")
	expires := 1712345678.0
	input := []StoredCookie{
		{
			Name:     "session",
			Value:    "token",
			Domain:   ".example.com",
			Path:     "/",
			Expires:  &expires,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		},
	}

	if err := writeStoredCookies(filePath, input); err != nil {
		t.Fatalf("writeStoredCookies() error = %v", err)
	}

	output, err := readStoredCookies(filePath)
	if err != nil {
		t.Fatalf("readStoredCookies() error = %v", err)
	}
	if len(output) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(output))
	}
	if output[0].Name != input[0].Name || output[0].Value != input[0].Value || output[0].SameSite != input[0].SameSite {
		t.Fatalf("unexpected cookie round-trip output: %+v", output[0])
	}
	if output[0].Expires == nil || *output[0].Expires != expires {
		t.Fatalf("unexpected expires round-trip output: %+v", output[0].Expires)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat cookie file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 permissions, got %o", info.Mode().Perm())
	}
}

func TestReadStoredCookiesRejectsInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "cookies.json")
	if err := os.WriteFile(filePath, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write invalid cookie file: %v", err)
	}

	if _, err := readStoredCookies(filePath); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
