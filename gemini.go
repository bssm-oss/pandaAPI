package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func geminiInputSelectors() []string {
	return []string{
		"rich-textarea div[contenteditable='true']",
		"div.ql-editor[contenteditable='true']",
		"div[contenteditable='true'][role='textbox']",
		"div[contenteditable='true']",
	}
}

func geminiSubmitSelectors() []string {
	return []string{
		"button[aria-label*='Send message']",
		"button[aria-label*='Submit']",
		"button.send-button",
	}
}

func geminiStopSelectors() []string {
	return []string{
		"button[aria-label*='Stop']",
		"button[aria-label*='Cancel']",
	}
}

func AskGemini(ctx context.Context, query string) (string, error) {
	if err := FetchPageWithRetry(ctx, "https://gemini.google.com/", nil); err != nil {
		return "", fmt.Errorf("open Gemini: %w", err)
	}
	needsAuth, err := geminiNeedsAuth(ctx)
	if err != nil {
		return "", err
	}
	if needsAuth {
		return "", fmt.Errorf("Run 'pandaapi auth' first")
	}
	inputSelector, err := WaitForAnySelector(ctx, geminiInputSelectors(), 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("find Gemini input: %w", err)
	}

	beforeState, err := geminiMessageState(ctx)
	if err != nil {
		return "", err
	}
	if err := SubmitPrompt(ctx, inputSelector, query, geminiSubmitSelectors()); err != nil {
		return "", fmt.Errorf("submit Gemini prompt: %w", err)
	}
	return waitForGeminiAnswer(ctx, beforeState)
}

func geminiNeedsAuth(ctx context.Context) (bool, error) {
	expression := `(function() {
		const url = window.location.href.toLowerCase();
		if (url.includes('accounts.google.com') || url.includes('signin')) return true;
		const candidates = [...document.querySelectorAll('a,button')];
		return candidates.some((node) => {
			const text = (node.innerText || node.textContent || '').trim().toLowerCase();
			return text === 'sign in' || text === 'log in';
		});
	})()`
	return EvaluateBool(ctx, expression)
}

func geminiMessageState(ctx context.Context) (providerMessageState, error) {
	expression := `(function() {
		const selectorSets = [
			'model-response',
			'div[data-message-model-slug]',
			'div[data-response-id]',
			'div.response-content',
			'div.model-response-text'
		];
		const values = [];
		for (const selector of selectorSets) {
			for (const node of document.querySelectorAll(selector)) {
				const text = (node.innerText || node.textContent || '').trim();
				if (text) values.push(text);
			}
		}
		return {
			count: values.length,
			text: values.length ? values[values.length - 1] : ''
		};
	})()`
	var state providerMessageState
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &state)); err != nil {
		return providerMessageState{}, fmt.Errorf("read Gemini messages: %w", err)
	}
	return state, nil
}

func waitForGeminiAnswer(ctx context.Context, before providerMessageState) (string, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	stableText := ""
	stableCount := 0
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for Gemini answer: %w", ctx.Err())
		case <-ticker.C:
			state, err := geminiMessageState(ctx)
			if err != nil {
				continue
			}
			started := state.Count > before.Count || (state.Text != "" && state.Text != before.Text)
			if !started {
				continue
			}
			stopVisible, _ := AnySelectorPresent(ctx, geminiStopSelectors())
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
