package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	newServerMux(func(query, provider string) (string, error) {
		return "", nil
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServerAskDefaultsProvider(t *testing.T) {
	body, _ := json.Marshal(askRequest{Query: "hello"})
	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	newServerMux(func(query, provider string) (string, error) {
		if query != "hello" || provider != ProviderChatGPT {
			return "", errors.New("unexpected request")
		}
		return "world", nil
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServerAskRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	newServerMux(func(query, provider string) (string, error) {
		return "", nil
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestServerAskReturnsUnauthorizedForAuthError(t *testing.T) {
	body, _ := json.Marshal(askRequest{Query: "hello", Provider: ProviderGemini})
	req := httptest.NewRequest(http.MethodPost, "/ask", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	newServerMux(func(query, provider string) (string, error) {
		return "", errors.New("Run 'pandaapi auth' first")
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
