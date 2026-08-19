package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/connectors/boundary"
)

// runOwnership implements `connectorgen ownership [repo-root] [--json] [--base <ref>] [--scope-file <path>]`.
func runOwnership(args []string, stdout, stderr io.Writer) int {
	root := ""
	asJSON := false
	opts := boundary.OwnershipOptions{}
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			asJSON = true
		case a == "-h" || a == "--help" || a == "help":
			logln(stdout, ownershipUsage())
			return 0
		case a == "--base":
			if i+1 >= len(args) {
				logln(stderr, "connectorgen ownership: --base requires a value")
				return 2
			}
			i++
			opts.BaseRef = args[i]
		case strings.HasPrefix(a, "--base="):
			opts.BaseRef = strings.TrimPrefix(a, "--base=")
			if opts.BaseRef == "" {
				logln(stderr, "connectorgen ownership: --base requires a value")
				return 2
			}
		case a == "--scope-file":
			if i+1 >= len(args) {
				logln(stderr, "connectorgen ownership: --scope-file requires a value")
				return 2
			}
			i++
			opts.ScopeFile = args[i]
		case strings.HasPrefix(a, "--scope-file="):
			opts.ScopeFile = strings.TrimPrefix(a, "--scope-file=")
			if opts.ScopeFile == "" {
				logln(stderr, "connectorgen ownership: --scope-file requires a value")
				return 2
			}
		case a == "--changed-path":
			if i+1 >= len(args) {
				logln(stderr, "connectorgen ownership: --changed-path requires a value")
				return 2
			}
			i++
			opts.ChangedPaths = append(opts.ChangedPaths, args[i])
		case strings.HasPrefix(a, "--changed-path="):
			value := strings.TrimPrefix(a, "--changed-path=")
			if value == "" {
				logln(stderr, "connectorgen ownership: --changed-path requires a value")
				return 2
			}
			opts.ChangedPaths = append(opts.ChangedPaths, value)
		case len(a) > 0 && a[0] == '-':
			logf(stderr, "connectorgen ownership: unknown flag %q\n%s\n", a, ownershipUsage())
			return 2
		default:
			if root != "" {
				logf(stderr, "connectorgen ownership: unexpected extra argument %q\n", a)
				return 2
			}
			root = a
		}
	}
	if root == "" {
		var err error
		root, err = repoRoot()
		if err != nil {
			logln(stderr, "connectorgen ownership:", err)
			return 2
		}
	}
	root = filepath.Clean(root)

	report, err := boundary.ValidateOwnership(root, opts)
	if err != nil {
		var cfgErr *boundary.ConfigError
		if errors.As(err, &cfgErr) {
			logln(stderr, "connectorgen ownership:", cfgErr)
			return 2
		}
		logln(stderr, "connectorgen ownership:", err)
		return 1
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			logln(stderr, "connectorgen ownership: encode report:", err)
			return 1
		}
	} else {
		renderOwnershipText(stdout, report)
	}
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func ownershipUsage() string {
	return `usage:
  connectorgen ownership [repo-root] [--json] [--base <ref>] [--scope-file <path>]
  connectorgen ownership [repo-root] [--json] --changed-path <path> [--changed-path <path>...] [--scope-file <path>]

Validates changed paths for one connector implementation target. The optional
scope file is a machine-readable contract:
  {"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["target-slug"]}

Without --scope-file, the command infers the target connector from changed
connector-owned paths so label or tag omission cannot skip the gate.

Exit status:
  0 clean
  1 ownership violations
  2 invalid invocation or validator configuration`
}

func renderOwnershipText(w io.Writer, report boundary.OwnershipReport) {
	for _, f := range report.Findings {
		path := f.Path
		if path == "" {
			path = "<scope>"
		}
		if f.Connector != "" {
			logf(w, "%s: [%s] %s: %s\n", path, f.Rule, f.Connector, f.Message)
		} else {
			logf(w, "%s: [%s] %s\n", path, f.Rule, f.Message)
		}
		if f.Remediation != "" {
			logf(w, "  remediation: %s\n", f.Remediation)
		}
	}
	mode := "worktree"
	if report.BaseRef != "" {
		mode = fmt.Sprintf("base=%s", report.BaseRef)
	}
	target := report.TargetConnector
	if target == "" {
		target = "<none>"
	}
	logf(w, "connectorgen ownership: %s, target=%s, %d changed path(s), %d finding(s)\n", mode, target, len(report.ChangedPaths), len(report.Findings))
}
