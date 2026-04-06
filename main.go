package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	runAuthFn = RunAuth
	runAskFn  = RunAsk
)

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeUsage(stderr)
		return 1
	}

	switch args[0] {
	case "help", "-h", "--help":
		writeUsage(stdout)
		return 0
	case "auth":
		provider, err := parseAuthArgs(args[1:])
		if err != nil {
			fmt.Fprintln(stderr, err)
			writeAuthUsage(stderr)
			return 1
		}
		if err := runAuthFn(provider); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "ask":
		query, provider, err := parseAskArgs(args[1:])
		if err != nil {
			fmt.Fprintln(stderr, err)
			writeAskUsage(stderr)
			return 1
		}
		if err := runAskFn(query, provider); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		writeUsage(stderr)
		return 1
	}
}

func parseAuthArgs(args []string) (string, error) {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	provider := fs.String("provider", "", "Provider to authenticate: chatgpt or gemini")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 0 {
		return "", fmt.Errorf("unexpected auth arguments: %s", strings.Join(fs.Args(), " "))
	}
	if strings.TrimSpace(*provider) == "" {
		return "", errors.New("--provider is required")
	}
	providerValue, err := NormalizeProvider(*provider)
	if err != nil {
		return "", err
	}
	return providerValue, nil
}

func parseAskArgs(args []string) (string, string, error) {
	fs := flag.NewFlagSet("ask", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	query := fs.String("query", "", "Question to send")
	provider := fs.String("provider", ProviderChatGPT, "Provider to use: chatgpt or gemini")
	if err := fs.Parse(args); err != nil {
		return "", "", err
	}
	if fs.NArg() != 0 {
		return "", "", fmt.Errorf("unexpected ask arguments: %s", strings.Join(fs.Args(), " "))
	}
	if strings.TrimSpace(*query) == "" {
		return "", "", errors.New("--query is required")
	}
	providerValue, err := NormalizeProvider(*provider)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(*query), providerValue, nil
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  pandaapi auth --provider chatgpt|gemini")
	fmt.Fprintln(w, "  pandaapi ask --query \"question\" [--provider chatgpt|gemini]")
}

func writeAuthUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: pandaapi auth --provider chatgpt|gemini")
}

func writeAskUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: pandaapi ask --query \"question\" [--provider chatgpt|gemini]")
}
