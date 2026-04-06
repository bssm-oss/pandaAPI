package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type StoredCookie struct {
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	Domain   string   `json:"domain,omitempty"`
	Path     string   `json:"path,omitempty"`
	Expires  *float64 `json:"expires,omitempty"`
	HTTPOnly bool     `json:"httpOnly,omitempty"`
	Secure   bool     `json:"secure,omitempty"`
	SameSite string   `json:"sameSite,omitempty"`
}

func SaveCookies(ctx context.Context, filePath string) error {
	var browserCookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(runCtx context.Context) error {
		if err := network.Enable().Do(runCtx); err != nil {
			return err
		}
		cookies, err := network.GetCookies().Do(runCtx)
		if err != nil {
			return err
		}
		browserCookies = cookies
		return nil
	})); err != nil {
		return fmt.Errorf("read browser cookies: %w", err)
	}

	persisted := make([]StoredCookie, 0, len(browserCookies))
	for _, cookie := range browserCookies {
		stored := StoredCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			HTTPOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
			SameSite: string(cookie.SameSite),
		}
		if cookie.Expires != 0 {
			expires := float64(cookie.Expires)
			stored.Expires = &expires
		}
		persisted = append(persisted, stored)
	}
	return writeStoredCookies(filePath, persisted)
}

func writeStoredCookies(filePath string, cookies []StoredCookie) error {

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("create cookie file directory: %w", err)
	}
	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cookies: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("write cookies: %w", err)
	}
	return nil
}

func LoadCookies(ctx context.Context, filePath string) error {
	cookies, err := readStoredCookies(filePath)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(runCtx context.Context) error {
		if err := network.Enable().Do(runCtx); err != nil {
			return err
		}
		for _, cookie := range cookies {
			setCookie := network.SetCookie(cookie.Name, cookie.Value)
			if cookie.Domain != "" {
				setCookie = setCookie.WithDomain(cookie.Domain)
			}
			if cookie.Path != "" {
				setCookie = setCookie.WithPath(cookie.Path)
			}
			setCookie = setCookie.WithHTTPOnly(cookie.HTTPOnly)
			setCookie = setCookie.WithSecure(cookie.Secure)
			if cookie.SameSite != "" {
				setCookie = setCookie.WithSameSite(network.CookieSameSite(cookie.SameSite))
			}
			if cookie.Expires != nil {
				expires := cdp.TimeSinceEpoch(time.Unix(int64(*cookie.Expires), 0).UTC())
				setCookie = setCookie.WithExpires(&expires)
			}
			if err := setCookie.Do(runCtx); err != nil {
				return err
			}
		}
		return nil
	}))
}

func readStoredCookies(filePath string) ([]StoredCookie, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read cookies: %w", err)
	}

	var cookies []StoredCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("unmarshal cookies: %w", err)
	}
	return cookies, nil
}
