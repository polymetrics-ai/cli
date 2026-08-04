package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/agentcontract"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "check":
		root, ok := parseRoot(args[1:], stderr)
		if !ok {
			return 2
		}
		if err := agentcontract.CheckRoot(ctx, root); err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: check failed: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "agentcontractgen: canonical contract and registered projections are current")
		return 0

	case "render":
		flags := flag.NewFlagSet("render", flag.ContinueOnError)
		flags.SetOutput(stderr)
		rootFlag := flags.String("root", "", "repository root")
		role := flags.String("role", "", "canonical role name")
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *role == "" {
			if err == nil {
				_, _ = fmt.Fprintln(stderr, "agentcontractgen: render requires --role and no positional arguments")
			}
			return 2
		}
		root, err := resolveRoot(*rootFlag)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: %v\n", err)
			return 2
		}
		contract, err := agentcontract.Load(filepath.Join(root, agentcontract.SourcePath))
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: %v\n", err)
			return 1
		}
		block, err := agentcontract.RenderBlock(contract, *role)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: %v\n", err)
			return 1
		}
		if _, err := stdout.Write(block); err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: write rendered contract: %v\n", err)
			return 1
		}
		return 0

	case "sync":
		root, ok := parseRoot(args[1:], stderr)
		if !ok {
			return 2
		}
		contract, err := agentcontract.Load(filepath.Join(root, agentcontract.SourcePath))
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: %v\n", err)
			return 1
		}
		updated, err := agentcontract.SyncProjections(root, contract)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: sync failed: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "agentcontractgen: synchronized %d registered projection(s)\n", updated)
		return 0

	default:
		_, _ = fmt.Fprintf(stderr, "agentcontractgen: unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func parseRoot(args []string, stderr io.Writer) (string, bool) {
	flags := flag.NewFlagSet("root", flag.ContinueOnError)
	flags.SetOutput(stderr)
	rootFlag := flags.String("root", "", "repository root")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return "", false
	}
	root, err := resolveRoot(*rootFlag)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agentcontractgen: %v\n", err)
		return "", false
	}
	return root, true
}

func resolveRoot(value string) (string, error) {
	if value != "" {
		absolute, err := filepath.Abs(value)
		if err != nil {
			return "", fmt.Errorf("resolve root: %w", err)
		}
		if _, err := os.Stat(filepath.Join(absolute, agentcontract.SourcePath)); err != nil {
			return "", fmt.Errorf("root %s does not contain %s: %w", absolute, agentcontract.SourcePath, err)
		}
		return absolute, nil
	}

	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, agentcontract.SourcePath)); err == nil {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("could not find %s from working directory", agentcontract.SourcePath)
		}
		directory = parent
	}
}

func usage(output io.Writer) {
	_, _ = fmt.Fprintln(output, "usage: agentcontractgen <check|render|sync> [--root <path>] [--role <name>]")
}
