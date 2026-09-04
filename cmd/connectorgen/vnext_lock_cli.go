package main

import (
	"context"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func runLockRenderContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	connector, defsRoot, check, code := parseVNextLockArgs("lock-render", args, stderr)
	if code != 0 {
		return code
	}
	publisher, err := newVNextGenerationPublisher(defsRoot, connector, vNextPublicationHooks{})
	if err != nil {
		logf(stderr, "connectorgen lock-render: initialize publisher: %v\n", err)
		return 1
	}
	raw, err := publisher.readSourceLock()
	if err != nil {
		logf(stderr, "connectorgen lock-render: read source lock: %v\n", err)
		return 1
	}
	lock, err := decodeVNextSourceLock(raw)
	if err != nil {
		logf(stderr, "connectorgen lock-render: parse source lock: %v\n", err)
		return 1
	}
	canonical, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		logf(stderr, "connectorgen lock-render: %v\n", err)
		return 1
	}
	if canonical.Connector != connector {
		logf(stderr, "connectorgen lock-render: lock connector %q does not match target %q\n", canonical.Connector, connector)
		return 1
	}
	artifacts, err := vNextPublicationArtifactsForStage(raw, connector, canonical.Staged)
	if err != nil {
		logf(stderr, "connectorgen lock-render: stage publication artifacts: %v\n", err)
		return 1
	}
	if check {
		if err := publisher.CheckContext(ctx, artifacts); err != nil {
			logf(stderr, "connectorgen lock-render: check published generation: %v\n", err)
			return 1
		}
		logf(stdout, "vNext execution generation is current: %s\n", connector)
		return 0
	}
	pointer, err := publisher.PublishContext(ctx, artifacts)
	if err != nil {
		logf(stderr, "connectorgen lock-render: publish generation: %v\n", err)
		return 1
	}
	logf(stdout, "published vNext execution generation: %s (%d files, %s)\n", connector, len(artifacts.Files), pointer.Generation)
	return 0
}

func parseVNextLockArgs(command string, args []string, stderr io.Writer) (connector, defsRoot string, check bool, code int) {
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--defs":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logf(stderr, "connectorgen %s: --defs requires a path\n", command)
				return "", "", false, 2
			}
			index++
			defsRoot = args[index]
		case "--check":
			check = true
		default:
			if strings.HasPrefix(args[index], "-") || connector != "" {
				logf(stderr, "connectorgen %s: unexpected argument %q\n", command, args[index])
				return "", "", false, 2
			}
			connector = args[index]
		}
	}
	if !namePattern.MatchString(connector) {
		logf(stderr, "connectorgen %s: connector %q is invalid\n", command, connector)
		return "", "", false, 2
	}
	if defsRoot == "" {
		root, err := repoRoot()
		if err != nil {
			logf(stderr, "connectorgen %s: resolve repository root: %v\n", command, err)
			return "", "", false, 1
		}
		defsRoot = filepath.Join(root, "internal", "connectors", "defs")
	}
	abs, err := filepath.Abs(defsRoot)
	if err != nil {
		logf(stderr, "connectorgen %s: resolve defs root: %v\n", command, err)
		return "", "", false, 1
	}
	return connector, abs, check, 0
}

func sortedOutputNames(outputs map[string][]byte) []string {
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
