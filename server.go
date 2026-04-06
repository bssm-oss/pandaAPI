package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultServerAddr = "127.0.0.1:8080"

type serverConfig struct {
	Addr string
}

type askRequest struct {
	Query    string `json:"query"`
	Provider string `json:"provider,omitempty"`
}

type askResponse struct {
	Provider string `json:"provider,omitempty"`
	Answer   string `json:"answer,omitempty"`
	Error    string `json:"error,omitempty"`
}

func parseServerArgs(args []string) (serverConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	addr := fs.String("addr", defaultServerAddr, "Address to bind the server to")
	if err := fs.Parse(args); err != nil {
		return serverConfig{}, err
	}
	if fs.NArg() != 0 {
		return serverConfig{}, fmt.Errorf("unexpected server arguments: %s", strings.Join(fs.Args(), " "))
	}
	if strings.TrimSpace(*addr) == "" {
		return serverConfig{}, errors.New("--addr is required")
	}
	return serverConfig{Addr: strings.TrimSpace(*addr)}, nil
}

func RunServer(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: newServerMux(RunServerAsk),
	}
	return server.ListenAndServe()
}

func newServerMux(askRunner func(query, provider string) (string, error)) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, askResponse{Error: "method not allowed"})
			return
		}

		defer r.Body.Close()
		var req askRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, askResponse{Error: "invalid JSON body"})
			return
		}

		provider := strings.TrimSpace(req.Provider)
		if provider == "" {
			provider = ProviderChatGPT
		}
		providerValue, err := NormalizeProvider(provider)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, askResponse{Error: err.Error()})
			return
		}

		answer, err := askRunner(req.Query, providerValue)
		if err != nil {
			writeJSON(w, statusForAskError(err), askResponse{Provider: providerValue, Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, askResponse{Provider: providerValue, Answer: answer})
	})
	return mux
}

func statusForAskError(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "--query is required"), strings.Contains(msg, "unsupported provider"), strings.Contains(msg, "invalid JSON body"):
		return http.StatusBadRequest
	case strings.Contains(msg, "Run 'pandaapi auth' first"):
		return http.StatusUnauthorized
	default:
		return http.StatusBadGateway
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
