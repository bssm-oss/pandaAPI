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
	if provider == ProviderGemini {
		return askRuntimePlan{headless: false, requireCookie: false}
	}
	return askRuntimePlan{headless: true, requireCookie: true}
}

func determineServerAskRuntime(_ string, _ bool) askRuntimePlan {
	return askRuntimePlan{headless: true, requireCookie: true}
}

func RunAsk(query string, provider string) error {
	answer, err := runAsk(query, provider, determineAskRuntime, true)
	if err != nil {
		return err
	}

	fmt.Println(answer)
	return nil
}

func RunServerAsk(query string, provider string) (string, error) {
	return runAsk(query, provider, determineServerAskRuntime, false)
}

func runAsk(query string, provider string, planner func(string, bool) askRuntimePlan, allowGeminiFallback bool) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("--query is required")
	}

	providerConfig, err := GetProviderConfig(provider)
	if err != nil {
		return "", err
	}
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	cookieFile := CookieFilePath(cfg, providerConfig.Name)
	_, statErr := os.Stat(cookieFile)
	cookieExists := statErr == nil
	if statErr != nil {
		if os.IsNotExist(statErr) {
			cookieExists = false
		} else {
			return "", fmt.Errorf("stat cookie file: %w", statErr)
		}
	}

	plan := planner(providerConfig.Name, cookieExists)
	if plan.requireCookie && !cookieExists {
		return "", errors.New("Run 'pandaapi auth' first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AskTimeout)
	defer cancel()

	browserCtx, browserCancel, err := NewBrowserContext(ctx, cfg.LightpandaCDP, plan.headless)
	if err != nil {
		return "", err
	}
	defer browserCancel()

	if plan.requireCookie {
		if err := LoadCookies(browserCtx, cookieFile); err != nil {
			if providerConfig.Name == ProviderGemini && allowGeminiFallback {
				return runAsk(query, provider, func(string, bool) askRuntimePlan {
					return askRuntimePlan{headless: false, requireCookie: false}
				}, false)
			}
			return "", fmt.Errorf("load cookies: %w", err)
		}
	} else if providerConfig.Name == ProviderGemini && cookieExists {
		_ = LoadCookies(browserCtx, cookieFile)
	}

	var answer string
	switch providerConfig.Name {
	case ProviderChatGPT:
		answer, err = AskChatGPT(browserCtx, query)
	case ProviderGemini:
		answer, err = AskGemini(browserCtx, query)
	default:
		return "", fmt.Errorf("unsupported provider %q", providerConfig.Name)
	}
	if err != nil {
		if providerConfig.Name == ProviderGemini && allowGeminiFallback && shouldFallbackGeminiSignedOut(err) && plan.requireCookie {
			return runAsk(query, provider, func(string, bool) askRuntimePlan {
				return askRuntimePlan{headless: false, requireCookie: false}
			}, false)
		}
		return "", err
	}

	return strings.TrimSpace(answer), nil
}

func shouldFallbackGeminiSignedOut(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Run 'pandaapi auth' first") ||
		strings.Contains(msg, "submit Gemini prompt: UnknownMethod") ||
		strings.Contains(msg, "find Gemini input:") ||
		strings.Contains(msg, "wait for Gemini answer:")
}
