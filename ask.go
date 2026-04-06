package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func RunAsk(query string, provider string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("--query is required")
	}

	providerConfig, err := GetProviderConfig(provider)
	if err != nil {
		return err
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	cookieFile := CookieFilePath(cfg, providerConfig.Name)
	if _, err := os.Stat(cookieFile); err != nil {
		if os.IsNotExist(err) {
			return errors.New("Run 'pandaapi auth' first")
		}
		return fmt.Errorf("stat cookie file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AskTimeout)
	defer cancel()

	browserCtx, browserCancel, err := NewBrowserContext(ctx, cfg.LightpandaCDP, true)
	if err != nil {
		return err
	}
	defer browserCancel()

	if err := LoadCookies(browserCtx, cookieFile); err != nil {
		return fmt.Errorf("load cookies: %w", err)
	}

	var answer string
	switch providerConfig.Name {
	case ProviderChatGPT:
		answer, err = AskChatGPT(browserCtx, query)
	case ProviderGemini:
		answer, err = AskGemini(browserCtx, query)
	default:
		return fmt.Errorf("unsupported provider %q", providerConfig.Name)
	}
	if err != nil {
		return err
	}

	fmt.Println(strings.TrimSpace(answer))
	return nil
}
