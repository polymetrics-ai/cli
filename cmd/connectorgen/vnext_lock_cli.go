package main

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

func runLockRenderContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return runLockRenderContextWithHooks(ctx, args, stdout, stderr, vNextPublicationHooks{})
}

func runLockRenderContextWithHooks(ctx context.Context, args []string, stdout, stderr io.Writer, hooks vNextPublicationHooks) int {
	connector, defsRoot, check, code := parseVNextLockArgs("lock-render", args, stderr)
	if code != 0 {
		return code
	}
	publisher, err := newVNextGenerationPublisher(defsRoot, connector, hooks)
	if err != nil {
		logf(stderr, "connectorgen lock-render: initialize publisher: %v\n", err)
		return 1
	}
	mode, create := syscall.LOCK_EX, true
	if check {
		mode, create = syscall.LOCK_SH, false
	}
	operation, err := publisher.openOperationRoot(ctx, false)
	if err != nil {
		logf(stderr, "connectorgen lock-render: open publication root: %v\n", err)
		return 1
	}
	defer operation.close()
	// Admit before creating publication state, then reject a source mutation
	// when the same retained operation rereads it under the publication lock.
	raw, err := publisher.readSourceLock(operation)
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
	if err := publisher.acquireOperation(ctx, operation, mode); err != nil {
		logf(stderr, "connectorgen lock-render: acquire publication operation: %v\n", err)
		return 1
	}
	lockedRaw, err := publisher.readSourceLock(operation)
	if err != nil {
		logf(stderr, "connectorgen lock-render: reread source lock: %v\n", err)
		return 1
	}
	if !bytes.Equal(raw, lockedRaw) {
		logf(stderr, "connectorgen lock-render: source lock changed during admission; retry\n")
		return 1
	}
	if err := operation.openGenerations(ctx, create); err != nil {
		logf(stderr, "connectorgen lock-render: open generation root: %v\n", err)
		return 1
	}
	if check {
		if err := publisher.checkLocked(operation, artifacts.Files, artifacts.Validate); err != nil {
			logf(stderr, "connectorgen lock-render: check published generation: %v\n", err)
			return 1
		}
		logf(stdout, "vNext execution generation is current: %s\n", connector)
		return 0
	}
	pointer, err := publisher.publishLocked(operation, artifacts.Files, artifacts.Validate)
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
