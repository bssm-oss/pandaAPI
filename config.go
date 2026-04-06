package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ProviderChatGPT    = "chatgpt"
	ProviderGemini     = "gemini"
	defaultCDPURL      = "http://127.0.0.1:9222"
	defaultAskTimeout  = 90 * time.Second
	defaultAuthTimeout = 15 * time.Minute
)

type Config struct {
	LightpandaCDP string
	CookieDir     string
	AskTimeout    time.Duration
	AuthTimeout   time.Duration
}

type ProviderConfig struct {
	Name            string
	AuthURL         string
	ChatURL         string
	InputSelectors  []string
	SubmitSelectors []string
	StopSelectors   []string
}

func LoadConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}

	cookieDir := strings.TrimSpace(os.Getenv("PANDAAPI_COOKIE_DIR"))
	if cookieDir == "" {
		cookieDir = filepath.Join(home, ".pandaapi")
	}
	if err := os.MkdirAll(cookieDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("create cookie directory: %w", err)
	}

	cdpURL := strings.TrimSpace(os.Getenv("LIGHTPANDA_CDP"))
	if cdpURL == "" {
		cdpURL = defaultCDPURL
	}

	return Config{
		LightpandaCDP: cdpURL,
		CookieDir:     cookieDir,
		AskTimeout:    defaultAskTimeout,
		AuthTimeout:   defaultAuthTimeout,
	}, nil
}

func NormalizeProvider(provider string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderChatGPT:
		return ProviderChatGPT, nil
	case ProviderGemini:
		return ProviderGemini, nil
	default:
		return "", fmt.Errorf("unsupported provider %q", provider)
	}
}

func CookieFilePath(cfg Config, provider string) string {
	return filepath.Join(cfg.CookieDir, fmt.Sprintf("cookies_%s.json", provider))
}

func GetProviderConfig(provider string) (ProviderConfig, error) {
	providerValue, err := NormalizeProvider(provider)
	if err != nil {
		return ProviderConfig{}, err
	}

	switch providerValue {
	case ProviderChatGPT:
		return ProviderConfig{
			Name:            ProviderChatGPT,
			AuthURL:         "https://chat.openai.com/auth/login",
			ChatURL:         "https://chat.openai.com/",
			InputSelectors:  chatGPTInputSelectors(),
			SubmitSelectors: chatGPTSubmitSelectors(),
			StopSelectors:   chatGPTStopSelectors(),
		}, nil
	case ProviderGemini:
		return ProviderConfig{
			Name:            ProviderGemini,
			AuthURL:         "https://gemini.google.com/",
			ChatURL:         "https://gemini.google.com/",
			InputSelectors:  geminiInputSelectors(),
			SubmitSelectors: geminiSubmitSelectors(),
			StopSelectors:   geminiStopSelectors(),
		}, nil
	default:
		return ProviderConfig{}, fmt.Errorf("unsupported provider %q", provider)
	}
}
