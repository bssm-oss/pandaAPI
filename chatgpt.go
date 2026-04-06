package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

type providerMessageState struct {
	Count int    `json:"count"`
	Text  string `json:"text"`
}

func chatGPTInputSelectors() []string {
	return []string{
		"#prompt-textarea",
		"textarea[placeholder*='Send a message']",
		"textarea[data-id='root']",
	}
}

func chatGPTSubmitSelectors() []string {
	return []string{
		"button[data-testid='send-button']",
		"button[aria-label*='Send prompt']",
		"button[aria-label*='Send message']",
	}
}

func chatGPTStopSelectors() []string {
	return []string{
		"button[aria-label*='Stop']",
		"button[data-testid='stop-button']",
	}
}

func AskChatGPT(ctx context.Context, query string) (string, error) {
	if err := FetchPageWithRetry(ctx, "https://chat.openai.com/", nil); err != nil {
		return "", fmt.Errorf("open ChatGPT: %w", err)
	}
	needsAuth, err := chatGPTNeedsAuth(ctx)
	if err != nil {
		return "", err
	}
	if needsAuth {
		return "", fmt.Errorf("Run 'pandaapi auth' first")
	}

	inputSelector, err := WaitForAnySelector(ctx, chatGPTInputSelectors(), 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("find ChatGPT input: %w", err)
	}

	beforeState, err := chatGPTMessageState(ctx)
	if err != nil {
		return "", err
	}
	if err := SubmitPrompt(ctx, inputSelector, query, chatGPTSubmitSelectors()); err != nil {
		return "", fmt.Errorf("submit ChatGPT prompt: %w", err)
	}
	return waitForChatGPTAnswer(ctx, beforeState)
}

func chatGPTNeedsAuth(ctx context.Context) (bool, error) {
	expression := `(function() {
		const url = window.location.href.toLowerCase();
		if (url.includes('/auth') || url.includes('/login')) return true;
		const candidates = [...document.querySelectorAll('a,button')];
		return candidates.some((node) => {
			const text = (node.innerText || node.textContent || '').trim().toLowerCase();
			return text === 'log in' || text === 'sign up';
		});
	})()`
	return EvaluateBool(ctx, expression)
}

func chatGPTMessageState(ctx context.Context) (providerMessageState, error) {
	expression := `(function() {
		const nodes = [...document.querySelectorAll("[data-message-author-role='assistant']")]
			.map((node) => (node.innerText || node.textContent || '').trim())
			.filter(Boolean);
		return {
			count: nodes.length,
			text: nodes.length ? nodes[nodes.length - 1] : ''
		};
	})()`
	var state providerMessageState
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &state)); err != nil {
		return providerMessageState{}, fmt.Errorf("read ChatGPT messages: %w", err)
	}
	return state, nil
}

func waitForChatGPTAnswer(ctx context.Context, before providerMessageState) (string, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	stableText := ""
	stableCount := 0
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for ChatGPT answer: %w", ctx.Err())
		case <-ticker.C:
			state, err := chatGPTMessageState(ctx)
			if err != nil {
				continue
			}
			started := state.Count > before.Count || (state.Text != "" && state.Text != before.Text)
			if !started {
				continue
			}
			stopVisible, _ := AnySelectorPresent(ctx, chatGPTStopSelectors())
			if state.Text == stableText && state.Text != "" {
				stableCount++
			} else {
				stableText = state.Text
				stableCount = 1
			}
			if !stopVisible && stableCount >= 2 {
				return state.Text, nil
			}
		}
	}
}
