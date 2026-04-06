package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type askRuntimePlan struct {
	headless      bool
	requireCookie bool
}

func determineAskRuntime(provider string, cookieExists bool) askRuntimePlan {
	if provider == ProviderGemini && !cookieExists {
		return askRuntimePlan{headless: false, requireCookie: false}
	}
	return askRuntimePlan{headless: true, requireCookie: true}
}

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
	_, statErr := os.Stat(cookieFile)
	cookieExists := statErr == nil
	if statErr != nil {
		if os.IsNotExist(statErr) {
			cookieExists = false
		} else {
			return fmt.Errorf("stat cookie file: %w", statErr)
		}
	}

	plan := determineAskRuntime(providerConfig.Name, cookieExists)
	if plan.requireCookie && !cookieExists {
		return errors.New("Run 'pandaapi auth' first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AskTimeout)
	defer cancel()

	browserCtx, browserCancel, err := NewBrowserContext(ctx, cfg.LightpandaCDP, plan.headless)
	if err != nil {
		return err
	}
	defer browserCancel()

	if plan.requireCookie {
		if err := LoadCookies(browserCtx, cookieFile); err != nil {
			return fmt.Errorf("load cookies: %w", err)
		}
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
