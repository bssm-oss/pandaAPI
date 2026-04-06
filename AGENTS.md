# AGENTS.md

## Project Purpose

This repository contains `pandaapi`, a Go CLI that automates ChatGPT and Gemini through browser automation over CDP.

## Quick Start

```bash
go install github.com/bssm-oss/pandaAPI@latest
make deps
make test
make build
./pandaapi auth --provider chatgpt
./pandaapi ask --query "Hello" --provider chatgpt
```

## Install, Run, and Test Commands

- Install from GitHub: `go install github.com/bssm-oss/pandaAPI@latest`
- Install dependencies: `make deps`
- Run tests: `make test`
- Build binary: `make build`
- Clean binary: `make clean`

## Default Workflow

1. Read `README.md`, `AGENTS.md`, and `docs/changes/`.
2. Make the smallest change that satisfies the request.
3. Add or update tests for logic that can be exercised locally.
4. Update documentation when behavior, setup, or constraints change.
5. Run `make test` and `make build` before reporting completion.

## Definition of Done

- Requested behavior is implemented.
- `make test` passes.
- `make build` passes.
- Documentation matches the current CLI behavior.
- Manual CLI smoke checks were attempted and their outcomes were recorded.

## Code Style Principles

- Keep package structure flat unless a new subpackage is clearly necessary.
- Prefer small helpers over broad abstractions.
- Preserve explicit provider-specific logic when selectors differ materially.
- Return actionable errors.

## File Structure Principles

- `main.go` owns CLI parsing.
- `auth.go` and `ask.go` own command execution.
- `browser.go`, `cookies.go`, and `config.go` own shared infrastructure.
- `chatgpt.go` and `gemini.go` own provider-specific DOM automation.

## Documentation Principles

- Update `README.md` for user-visible behavior.
- Add a dated entry in `docs/changes/` for meaningful changes.
- Keep docs aligned with actual commands and environment variables.

## Testing Principles

- Add local tests for provider-independent logic and CLI parsing.
- Document gaps when live-provider or external-browser flows cannot be automated in CI.

## Branch, Commit, and PR Rules

- Do not work directly on the default branch.
- Use descriptive branch names like `feat/lightpanda-cli`.
- Keep commits atomic and separate code, tests, and docs when practical.

## Sensitive Paths

- Cookie files in the runtime cookie directory contain authentication state and must never be committed.
- Provider selector logic is sensitive to upstream UI changes.

## Pre-Work Checklist

- Confirm required environment variables and external dependencies.
- Confirm whether Lightpanda and a local Chrome installation are available.
- Confirm whether the repository is initialized and connected to a remote.

## Post-Work Checklist

- Run tests and build.
- Record manual verification results.
- Update documentation.
- Prepare branch, commits, and PR if git is available.

## Never Do

- Do not claim tests passed unless they were executed.
- Do not expose cookies or other secrets.
- Do not mark live browser automation as verified if Lightpanda or provider login was unavailable.
