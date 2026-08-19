package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	case "certification-gate":
		return runCertificationGate(args[1:], stdout, stderr)

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
		catalogUpdated, err := agentcontract.SyncCertificationFlowKindCatalog(root, contract)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: sync failed: %v\n", err)
			return 1
		}
		updated, err := agentcontract.SyncProjections(root, contract)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: sync failed: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "agentcontractgen: synchronized %d registered projection(s) and %d certification flow-kind catalog(s)\n", updated, catalogUpdated)
		return 0

	default:
		_, _ = fmt.Fprintf(stderr, "agentcontractgen: unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func runCertificationGate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("certification-gate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	rootFlag := flags.String("root", "", "repository root")
	connector := flags.String("connector", "", "connector name")
	transition := flags.String("transition", "", "protected transition")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeCertificationGateHalt(stdout, stderr, agentcontract.CertificationGateRequest{
				Connector:  *connector,
				Transition: *transition,
			}, "request/help", "request", "certification-gate help does not authorize a protected transition")
		}
		return writeCertificationGateHalt(stdout, stderr, agentcontract.CertificationGateRequest{
			Connector:  *connector,
			Transition: *transition,
		}, "request/decode", "request", err.Error())
	}
	if flags.NArg() != 0 {
		return writeCertificationGateHalt(stdout, stderr, agentcontract.CertificationGateRequest{
			Connector:  *connector,
			Transition: *transition,
		}, "request/decode", "request", "certification-gate accepts no positional arguments")
	}

	root, err := resolveCertificationGateRoot(*rootFlag)
	if err != nil {
		return writeCertificationGateHalt(stdout, stderr, agentcontract.CertificationGateRequest{
			Connector:  *connector,
			Transition: *transition,
		}, "request/root", "request", err.Error())
	}
	contract, err := agentcontract.Load(filepath.Join(root, agentcontract.SourcePath))
	if err != nil {
		return writeCertificationGateHalt(stdout, stderr, agentcontract.CertificationGateRequest{
			Connector:  *connector,
			Transition: *transition,
		}, "contract/invalid", "contract", err.Error())
	}

	request := agentcontract.CertificationGateRequest{
		SchemaVersion: contract.CertificationGate.InputSchemaVersion,
		Connector:     *connector,
		Transition:    *transition,
		Inputs:        contract.CertificationGate.Inputs,
	}
	verdict, err := agentcontract.EnforceCertificationGate(root, contract, request)
	if err != nil {
		var blocked *agentcontract.CertificationGateBlockedError
		if errors.As(err, &blocked) {
			if !writeCertificationGateVerdict(stdout, stderr, verdict) {
				return 1
			}
			_, _ = fmt.Fprintf(stderr, "agentcontractgen: certification gate %s blocked %s for %s\n", verdict.Decision, verdict.Transition, verdict.Connector)
			return 1
		}
		return writeCertificationGateHalt(stdout, stderr, request, "gate/evaluate", "gate", err.Error())
	}
	if !writeCertificationGateVerdict(stdout, stderr, verdict) {
		return 1
	}
	return 0
}

func resolveCertificationGateRoot(value string) (string, error) {
	if value == "" {
		return "", errors.New("certification-gate requires an explicit --root")
	}
	if !filepath.IsAbs(value) {
		return "", errors.New("certification-gate --root must be absolute")
	}
	if filepath.Clean(value) != value {
		return "", errors.New("certification-gate --root must be canonical without traversal")
	}
	info, err := os.Lstat(value)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("certification-gate --root must be an existing non-symlink directory")
	}
	resolved, err := filepath.EvalSymlinks(value)
	if err != nil || resolved != value {
		return "", errors.New("certification-gate --root must not contain symlink components")
	}
	if err := validateCertificationGateContractPath(value); err != nil {
		return "", err
	}
	return value, nil
}

func validateCertificationGateContractPath(root string) error {
	path := root
	components := strings.Split(filepath.ToSlash(agentcontract.SourcePath), "/")
	for index, component := range components {
		path = filepath.Join(path, component)
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("certification-gate contract must be a non-symlink regular file below --root")
		}
		if index < len(components)-1 && !info.IsDir() {
			return errors.New("certification-gate contract path has a non-directory ancestor")
		}
		if index == len(components)-1 && !info.Mode().IsRegular() {
			return errors.New("certification-gate contract must be a non-symlink regular file below --root")
		}
	}
	return nil
}

func writeCertificationGateHalt(stdout, stderr io.Writer, request agentcontract.CertificationGateRequest, id, class, message string) int {
	verdict := agentcontract.CertificationGateVerdict{
		SchemaVersion: 1,
		Connector:     request.Connector,
		Transition:    request.Transition,
		Decision:      agentcontract.CertificationGateHalt,
		Failures: []agentcontract.CertificationGateFailure{{
			ID:      id,
			Class:   class,
			Message: message,
		}},
	}
	if !writeCertificationGateVerdict(stdout, stderr, verdict) {
		return 1
	}
	_, _ = fmt.Fprintf(stderr, "agentcontractgen: certification gate HALT: %s\n", message)
	return 1
}

func writeCertificationGateVerdict(stdout, stderr io.Writer, verdict agentcontract.CertificationGateVerdict) bool {
	if err := json.NewEncoder(stdout).Encode(verdict); err != nil {
		_, _ = fmt.Fprintf(stderr, "agentcontractgen: write certification gate verdict: %v\n", err)
		return false
	}
	return true
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
	_, _ = fmt.Fprintln(output, "usage: agentcontractgen <check|certification-gate|render|sync> [--root <path>] [--connector <name>] [--transition <name>] [--role <name>]")
}
