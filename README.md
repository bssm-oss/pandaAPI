# pandaapi

`pandaapi` is a Go CLI that automates the ChatGPT and Google Gemini web interfaces through browser automation instead of official APIs. It uses a visible Chrome session for manual login and a Lightpanda CDP endpoint for headless prompt execution.

## Overview

- `pandaapi auth --provider chatgpt|gemini` opens a visible browser and saves authenticated cookies.
- `pandaapi ask --query "..." [--provider chatgpt|gemini]` loads cookies, connects to Lightpanda, sends a prompt, waits for completion, and prints the answer.
- Cookies are stored on disk under `~/.pandaapi` by default.

## Requirements

- Go 1.21+
- A running Lightpanda CDP server, default `http://127.0.0.1:9222`
- A locally installed Chrome or Chromium browser for `auth`

## Environment Variables

- `LIGHTPANDA_CDP`: overrides the Lightpanda CDP endpoint
- `PANDAAPI_COOKIE_DIR`: overrides the cookie storage directory

## Installation

Install directly from GitHub:

```bash
go install github.com/bssm-oss/pandaAPI@latest
```

Or build locally from a checkout:

```bash
make deps
make build
```

## Usage

Authenticate with a provider:

```bash
./pandaapi auth --provider chatgpt
./pandaapi auth --provider gemini
```

Ask a question:

```bash
./pandaapi ask --query "Hello" --provider chatgpt
./pandaapi ask --query "Summarize this page" --provider gemini
```

If `--provider` is omitted for `ask`, `chatgpt` is used.

## Project Layout

```text
pandaapi/
├── .github/workflows/ci.yml
├── AGENTS.md
├── docs/changes/2026-04-06-lightpanda-cli.md
├── go.mod
├── go.sum
├── main.go
├── auth.go
├── ask.go
├── browser.go
├── chatgpt.go
├── gemini.go
├── cookies.go
├── config.go
├── config_test.go
├── main_test.go
├── Makefile
└── README.md
```

## Behavior Notes

- `auth` uses a visible local browser session because Lightpanda is headless-oriented and login flows are more reliable in a standard browser.
- `ask` expects valid cookies to exist already and returns `Run 'pandaapi auth' first` when they do not.
- The CLI retries navigation and DOM interactions, uses fallback selectors, and waits for generation to finish by combining answer stability checks with stop-button disappearance.

## Verification

```bash
make deps
go vet ./...
make test
make build
GOBIN=$PWD/.bin go install .
```

Manual smoke checks should also include:

- `./pandaapi`
- `./pandaapi auth --provider chatgpt`
- `./pandaapi auth --provider gemini`
- `./pandaapi ask --query "Hello" --provider chatgpt`
- `./pandaapi ask --query "Hello" --provider gemini`

## Limitations

- ChatGPT and Gemini DOM structures change over time; selector fallbacks reduce but do not remove that risk.
- Login and answer extraction depend on the provider’s live web UI.
- End-to-end automation requires external services and a reachable Lightpanda CDP instance.
- If `http://127.0.0.1:9222` is not serving CDP endpoints, `ask` cannot be verified end to end until Lightpanda is started.
- In local verification, ChatGPT presented a browser-verification page under Lightpanda instead of the chat UI, so ChatGPT `ask` may require a different supported browser runtime even after authentication.

## Development

- `make deps`: resolve module dependencies
- `make test`: run tests
- `make build`: build the CLI
- `make clean`: remove the local binary

## GitHub Install Notes

- `go install github.com/bssm-oss/pandaAPI@latest` requires the repository to be published at that module path.
- After installation, the binary is available as `pandaapi` in your Go bin directory.
- Runtime dependencies still apply: Lightpanda for `ask`, and Chrome or Chromium for `auth`.
