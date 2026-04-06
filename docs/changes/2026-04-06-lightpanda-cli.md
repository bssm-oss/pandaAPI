# 2026-04-06 lightpanda CLI bootstrap

## Background

The repository had no implementation files, tests, docs, or workflows. The requested outcome was a working Go CLI for ChatGPT and Gemini browser automation using Lightpanda for headless execution and a visible browser for authentication.

## Problem or Goal

Create the initial CLI with reproducible local verification, cookie persistence, provider-specific automation, and documentation that reflects the actual behavior.

## Changes

- Added the initial Go module and CLI entrypoint.
- Implemented provider authentication and prompt execution flows.
- Added shared browser, config, and cookie helpers.
- Added local tests for provider-independent behavior and CLI parsing.
- Added README, AGENTS guidance, and CI.
- Added explicit GitHub installation guidance for `go install` consumers.
- Added focused tests for cookie persistence helpers and `ask` precondition handling.

## Design Rationale

- Authentication uses a visible local browser because login flows are less reliable in a headless remote browser.
- Prompt execution uses a Lightpanda remote allocator so the CLI can target a running CDP endpoint.
- Provider-specific files isolate selector drift and DOM differences.

## Impact Scope

- New CLI commands: `auth`, `ask`
- New runtime dependency on a reachable Lightpanda CDP endpoint for `ask`
- New runtime dependency on local Chrome or Chromium for `auth`

## Verification Method

- `make deps`
- `go vet ./...`
- `make test`
- `make build`
- `GOBIN=$PWD/.bin go install .`
- Manual CLI smoke checks for help output and unauthenticated failure path
- Lightpanda endpoint probe at `http://127.0.0.1:9222/json/version`

## Remaining Limitations

- End-to-end provider verification depends on live ChatGPT and Gemini UI availability.
- Selectors may need updates as upstream DOM changes.
- CI cannot complete real provider login or prompt flows.
- The current local environment returned `connection refused` for the default Lightpanda endpoint, so authenticated `ask` verification remains blocked until Lightpanda is running.

## Follow-Up Tasks

- Add fixture-backed tests for cookie serialization edge cases.
- Revalidate selectors against live provider UIs after initial deployment.
