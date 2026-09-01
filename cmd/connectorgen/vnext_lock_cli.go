package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runLockRender(args []string, stdout, stderr io.Writer) int {
	connector, defsRoot, check, code := parseVNextLockArgs("lock-render", args, stderr)
	if code != 0 {
		return code
	}
	lockPath := filepath.Join(defsRoot, connector, "source.lock.json")
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		logf(stderr, "connectorgen lock-render: read source lock: %v\n", err)
		return 1
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var lock vNextSourceLock
	if err := decoder.Decode(&lock); err != nil {
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
	outputs, err := renderVNextExecutionBundle(canonical)
	if err != nil {
		logf(stderr, "connectorgen lock-render: render execution bundle: %v\n", err)
		return 1
	}
	names := sortedOutputNames(outputs)
	for _, name := range names {
		outputPath := filepath.Join(defsRoot, connector, filepath.FromSlash(name))
		if check {
			current, readErr := os.ReadFile(outputPath)
			if readErr != nil || !bytes.Equal(current, outputs[name]) {
				logf(stderr, "connectorgen lock-render: %s differs from source.lock.json\n", filepath.ToSlash(outputPath))
				return 1
			}
			continue
		}
		if err := writeGeneratedArtifact(outputPath, outputs[name]); err != nil {
			logf(stderr, "connectorgen lock-render: write %s: %v\n", filepath.ToSlash(outputPath), err)
			return 1
		}
	}
	if check {
		logf(stdout, "vNext execution bundle is current: %s\n", connector)
	} else {
		logf(stdout, "rendered vNext execution bundle: %s (%d files)\n", connector, len(names))
	}
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

func writeGeneratedArtifact(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".connectorgen-vnext-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
