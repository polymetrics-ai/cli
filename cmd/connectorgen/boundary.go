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

// runBoundary implements `connectorgen boundary [repo-root] [--json] [--base <ref>]`.
func runBoundary(args []string, stdout, stderr io.Writer) int {
	root := ""
	asJSON := false
	opts := boundary.Options{}
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			asJSON = true
		case a == "-h" || a == "--help" || a == "help":
			logln(stdout, boundaryUsage())
			return 0
		case a == "--base":
			if i+1 >= len(args) {
				logln(stderr, "connectorgen boundary: --base requires a value")
				return 2
			}
			i++
			opts.BaseRef = args[i]
		case strings.HasPrefix(a, "--base="):
			opts.BaseRef = strings.TrimPrefix(a, "--base=")
			if opts.BaseRef == "" {
				logln(stderr, "connectorgen boundary: --base requires a value")
				return 2
			}
		case len(a) > 0 && a[0] == '-':
			logf(stderr, "connectorgen boundary: unknown flag %q\n%s\n", a, boundaryUsage())
			return 2
		default:
			if root != "" {
				logf(stderr, "connectorgen boundary: unexpected extra argument %q\n", a)
				return 2
			}
			root = a
		}
	}
	if root == "" {
		var err error
		root, err = repoRoot()
		if err != nil {
			logln(stderr, "connectorgen boundary:", err)
			return 2
		}
	}
	root = filepath.Clean(root)

	report, err := boundary.Scan(root, opts)
	if err != nil {
		var cfgErr *boundary.ConfigError
		if errors.As(err, &cfgErr) {
			logln(stderr, "connectorgen boundary:", cfgErr)
			return 2
		}
		logln(stderr, "connectorgen boundary:", err)
		return 1
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			logln(stderr, "connectorgen boundary: encode report:", err)
			return 1
		}
	} else {
		renderBoundaryText(stdout, report)
	}
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func boundaryUsage() string {
	return `usage:
  connectorgen boundary [repo-root] [--json] [--base <ref>]

Scans shared production Go for connector-specific policy outside defs, hooks,
natives, generated outputs, tests, fixtures, and documentation paths.

Exit status:
  0 clean
  1 policy violations
  2 invalid invocation or scanner configuration`
}

func renderBoundaryText(w io.Writer, report boundary.Report) {
	for _, f := range report.Findings {
		logf(w, "%s:%d: [%s] %s %q: %s\n", f.Path, f.Line, f.Rule, f.Connector, f.Match, f.Message)
		if f.Remediation != "" {
			logf(w, "  remediation: %s\n", f.Remediation)
		}
	}
	for _, wr := range report.Warnings {
		logf(w, "%s:%d: [warning:%s] %s %q: %s\n", wr.Path, wr.Line, wr.Rule, wr.Connector, wr.Match, wr.Message)
	}
	mode := report.Mode
	if report.BaseRef != "" {
		mode = fmt.Sprintf("%s(base=%s)", report.Mode, report.BaseRef)
	}
	logf(w, "connectorgen boundary: %s, %d file(s) checked, %d finding(s), %d warning(s), %d exception(s) applied\n", mode, report.CheckedFiles, len(report.Findings), len(report.Warnings), len(report.Exceptions))
}
