// Command connectorgen is the wave0 migration-tooling CLI for the
// declarative connector engine (design §C.3):
//
//	validate [dir] [--json]   loads and validates every bundle under dir
//	                           (default internal/connectors/defs), exit 1 on
//	                           any finding
//	gen                        regenerates hooks/hookset/hookset_gen.go and
//	                           native/nativeset/nativeset_gen.go
//	new <name>                 scaffolds internal/connectors/defs/<name>/
//
// cmd/connectorgen does not replace cmd/registrygen in wave0 (that happens in
// wave6); it is purely additive tooling for the new defs/ bundle format.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the full CLI entry point (argv without the program name); it is
// exercised directly by tests rather than shelling out to a built binary.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		logln(stderr, usage())
		return 2
	}

	switch args[0] {
	case "validate":
		return runValidate(args, stdout, stderr)
	case "gen":
		return runGen(args, stdout, stderr)
	case "new":
		return runNew(args, stdout, stderr)
	case "-h", "--help", "help":
		logln(stdout, usage())
		return 0
	default:
		logf(stderr, "connectorgen: unknown subcommand %q\n%s\n", args[0], usage())
		return 2
	}
}

// logln and logf write a diagnostic/status line to w (stdout or stderr) and
// deliberately discard the write error: these are terminal streams on a
// short-lived CLI process, and there is no recovery action available if a
// write to them fails (the process is about to exit with a non-zero code
// regardless). Named helpers make that discard an explicit, reviewed
// decision rather than a silent one.
func logln(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, a...)
}

func logf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func usage() string {
	return `usage:
  connectorgen validate [dir] [--json]   (default dir: internal/connectors/defs)
  connectorgen gen
  connectorgen new <name>`
}

// runValidate implements `connectorgen validate [dir] [--json]`.
func runValidate(args []string, stdout, stderr io.Writer) int {
	dir := ""
	asJSON := false
	for _, a := range args[1:] {
		switch a {
		case "--json":
			asJSON = true
		default:
			if dir != "" {
				logf(stderr, "connectorgen validate: unexpected extra argument %q\n", a)
				return 2
			}
			dir = a
		}
	}

	if dir == "" {
		root, err := repoRoot()
		if err != nil {
			logln(stderr, "connectorgen validate:", err)
			return 1
		}
		dir = filepath.Join(root, "internal/connectors/defs")
	}

	fsys := os.DirFS(dir)
	report, err := validateDir(fsys)
	if err != nil {
		logln(stderr, "connectorgen validate:", err)
		return 1
	}

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			logln(stderr, "connectorgen validate: encode report:", err)
			return 1
		}
	} else {
		renderText(stdout, report)
	}

	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

// renderText renders a Report as human-readable lines: one finding per line
// naming connector/file/rule, followed by a summary count. The summary
// line's wording ("N findings") is a stable self-verify contract (PLAN.md/
// SPEC.md grep for "0 findings") and is deliberately unaffected by
// Warnings, which render separately (N2, wave0 REVIEW.md carried flag: a
// warning never blocks the gate or changes the finding count).
func renderText(w io.Writer, report Report) {
	for _, f := range report.Findings {
		logf(w, "%s: %s: [%s] %s\n", f.Connector, f.File, f.Rule, f.Message)
	}
	if len(report.Findings) == 0 {
		logf(w, "connectorgen validate: %d connector(s) checked, 0 findings\n", report.ConnectorsChecked)
	} else {
		logf(w, "connectorgen validate: %d connector(s) checked, %d finding(s)\n", report.ConnectorsChecked, len(report.Findings))
	}
	for _, wr := range report.Warnings {
		logf(w, "%s: %s: [warning:%s] %s\n", wr.Connector, wr.File, wr.Rule, wr.Message)
	}
	if len(report.Warnings) > 0 {
		logf(w, "connectorgen validate: %d warning(s)\n", len(report.Warnings))
	}
}

// repoRoot finds the module root by walking up to the directory containing
// go.mod (mirrors cmd/registrygen's repoRoot).
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from working directory")
		}
		dir = parent
	}
}
