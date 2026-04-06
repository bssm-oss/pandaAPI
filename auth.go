package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func RunAuth(provider string) error {
	providerConfig, err := GetProviderConfig(provider)
	if err != nil {
		return err
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AuthTimeout)
	defer cancel()

	browserCtx, browserCancel, err := NewBrowserContext(ctx, cfg.LightpandaCDP, false)
	if err != nil {
		return err
	}
	defer browserCancel()

	if err := FetchPageWithRetry(browserCtx, providerConfig.AuthURL, nil); err != nil {
		return fmt.Errorf("open auth page: %w", err)
	}

	fmt.Printf("Opened %s login in a visible browser. Complete login, then press Enter.\n", providerConfig.Name)
	enterCh := startManualConfirmReader(os.Stdin)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("authentication timed out: %w", ctx.Err())
		case <-ticker.C:
			ready, err := ProviderReady(browserCtx, providerConfig.Name)
			if err != nil {
				continue
			}
			if ready {
				return SaveCookies(browserCtx, CookieFilePath(cfg, providerConfig.Name))
			}
		case _, ok := <-enterCh:
			if !ok {
				enterCh = nil
				continue
			}
			ready, err := ProviderReady(browserCtx, providerConfig.Name)
			if err != nil {
				return fmt.Errorf("check authentication state: %w", err)
			}
			if !ready {
				fmt.Println("Login was not detected yet. Continue in the browser, then press Enter again.")
				enterCh = startManualConfirmReader(os.Stdin)
				continue
			}
			if err := SaveCookies(browserCtx, CookieFilePath(cfg, providerConfig.Name)); err != nil {
				return err
			}
			fmt.Printf("Saved cookies to %s\n", CookieFilePath(cfg, providerConfig.Name))
			return nil
		}
	}
}

func startManualConfirmReader(input *os.File) <-chan struct{} {
	ch := make(chan struct{}, 1)
	if !term.IsTerminal(int(input.Fd())) {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		reader := bufio.NewReader(input)
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
		ch <- struct{}{}
	}()

	return ch
}

func ProviderReady(ctx context.Context, provider string) (bool, error) {
	switch strings.ToLower(provider) {
	case ProviderChatGPT:
		visible, err := AnySelectorPresent(ctx, chatGPTInputSelectors())
		if err != nil {
			return false, err
		}
		return visible, nil
	case ProviderGemini:
		visible, err := AnySelectorPresent(ctx, geminiInputSelectors())
		if err != nil {
			return false, err
		}
		return visible, nil
	default:
		return false, errors.New("unsupported provider")
	}
}
