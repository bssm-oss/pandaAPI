package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
)

func NewBrowserContext(ctx context.Context, cdpURL string, headless bool) (context.Context, context.CancelFunc, error) {
	quietErrf := func(string, ...interface{}) {}
	if headless {
		allocatorCtx, allocatorCancel := chromedp.NewRemoteAllocator(ctx, cdpURL)
		browserCtx, browserCancel := chromedp.NewContext(allocatorCtx, chromedp.WithErrorf(quietErrf))
		cancel := func() {
			browserCancel()
			allocatorCancel()
		}
		return browserCtx, cancel, nil
	}

	userDataDir, err := os.MkdirTemp("", "pandaapi-auth-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create auth user data dir: %w", err)
	}

	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("start-maximized", true),
		chromedp.UserDataDir(userDataDir),
	)

	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(ctx, allocatorOptions...)
	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx, chromedp.WithErrorf(quietErrf))
	cancel := func() {
		browserCancel()
		allocatorCancel()
		_ = os.RemoveAll(userDataDir)
	}
	return browserCtx, cancel, nil
}

func Retry(ctx context.Context, attempts int, initialDelay time.Duration, fn func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}

	delay := initialDelay
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt == attempts {
			break
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay *= 2
	}
	return lastErr
}

func FetchPageWithRetry(ctx context.Context, url string, waitSelectors []string) error {
	return Retry(ctx, 3, time.Second, func(runCtx context.Context) error {
		if err := chromedp.Run(runCtx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
		); err != nil {
			return err
		}
		if len(waitSelectors) == 0 {
			return nil
		}
		_, err := WaitForAnySelector(runCtx, waitSelectors, 20*time.Second)
		return err
	})
}

func WaitForAnySelector(ctx context.Context, selectors []string, timeout time.Duration) (string, error) {
	if len(selectors) == 0 {
		return "", fmt.Errorf("no selectors provided")
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	encoded, err := json.Marshal(selectors)
	if err != nil {
		return "", fmt.Errorf("marshal selectors: %w", err)
	}
	expr := fmt.Sprintf(`(function() {
		const selectors = %s;
		for (const selector of selectors) {
			const node = document.querySelector(selector);
			if (!node) continue;
			const style = window.getComputedStyle(node);
			if (style && style.visibility === 'hidden') continue;
			return selector;
		}
		return "";
	})()`, string(encoded))

	for {
		var matched string
		err := chromedp.Run(pollCtx, chromedp.Evaluate(expr, &matched))
		if err == nil && matched != "" {
			return matched, nil
		}
		if pollCtx.Err() != nil {
			if err != nil {
				return "", err
			}
			return "", pollCtx.Err()
		}
		select {
		case <-time.After(500 * time.Millisecond):
		case <-pollCtx.Done():
			if err != nil {
				return "", err
			}
			return "", pollCtx.Err()
		}
	}
}

func AnySelectorPresent(ctx context.Context, selectors []string) (bool, error) {
	if len(selectors) == 0 {
		return false, nil
	}
	encoded, err := json.Marshal(selectors)
	if err != nil {
		return false, fmt.Errorf("marshal selectors: %w", err)
	}
	expr := fmt.Sprintf(`(function() {
		const selectors = %s;
		return selectors.some((selector) => document.querySelector(selector));
	})()`, string(encoded))
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &ok)); err != nil {
		return false, err
	}
	return ok, nil
}

func EvaluateString(ctx context.Context, expression string) (string, error) {
	var value string
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &value)); err != nil {
		return "", err
	}
	return value, nil
}

func EvaluateBool(ctx context.Context, expression string) (bool, error) {
	var value bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &value)); err != nil {
		return false, err
	}
	return value, nil
}

func SubmitPrompt(ctx context.Context, selector string, query string, submitSelectors []string) error {
	if err := Retry(ctx, 3, 500*time.Millisecond, func(runCtx context.Context) error {
		return setPromptText(runCtx, selector, query)
	}); err != nil {
		return err
	}

	if matched, err := WaitForAnySelector(ctx, submitSelectors, 3*time.Second); err == nil && matched != "" {
		if err := Retry(ctx, 3, 400*time.Millisecond, func(runCtx context.Context) error {
			return chromedp.Run(runCtx, chromedp.Click(matched, chromedp.ByQuery))
		}); err == nil {
			return nil
		}
	}

	return submitPromptWithEnterKey(ctx, selector)
}

func setPromptText(ctx context.Context, selector string, query string) error {
	selectorLiteral := strconv.Quote(selector)
	queryLiteral := strconv.Quote(query)
	expr := fmt.Sprintf(`(() => {
		const el = document.querySelector(%s);
		if (!el) return "missing";
		const text = %s;
		if (el.matches('textarea, input')) {
			el.value = text;
		} else {
			el.textContent = text;
		}
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return "ok";
	})()`, selectorLiteral, queryLiteral)
	var result string
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &result)); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("set prompt text failed: %s", result)
	}
	return nil
}

func submitPromptWithEnterKey(ctx context.Context, selector string) error {
	selectorLiteral := strconv.Quote(selector)
	expr := fmt.Sprintf(`(() => {
		const el = document.querySelector(%s);
		if (!el) return "missing";
		const eventInit = { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter', keyCode: 13, which: 13 };
		el.dispatchEvent(new KeyboardEvent('keydown', eventInit));
		el.dispatchEvent(new KeyboardEvent('keypress', eventInit));
		el.dispatchEvent(new KeyboardEvent('keyup', eventInit));
		return "ok";
	})()`, selectorLiteral)
	var result string
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &result)); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("submit prompt with enter failed: %s", result)
	}
	return nil
}
