// Command connectorgen is the declarative connector engine tooling CLI
// (design §C.3):
//
//	validate [dir] [--json]   loads and validates every bundle under dir
//	                           (default internal/connectors/defs), exit 1 on
//	                           any finding
//	validate [dir] --connector <name> --require-operational-contract <profile>
//	                           additionally proves one declared capability
//	                           profile has a closed executable declaration
//	boundary [repo] [--json] [--base <ref>]
//	                           scans shared Go for connector-specific policy
//	                           outside definition-owned locations
//	ownership [repo] [--json] [--base <ref>] [--scope-file <path>]
//	                           validates changed paths for exactly one target connector
//	gen                        regenerates hooks/hookset/hookset_gen.go
//	lock-render <connector>    projects immutable schema-v4 source.lock.json
//	                           through the canonical operation descriptor into
//	                           a staged, closed authoring generation; --check
//	                           verifies its selected generation without writing
//	batch gate --manifest <path> --report <path>
//	                           records independent candidate validation and
//	                           runtime-preflight results
//
// It owns bundle validation plus generated hook imports for the
// connector-architecture-v2 runtime.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

func runMain(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "lock-render" {
		return run(args, stdout, stderr)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runContext(ctx, args, stdout, stderr)
}

// run is the full CLI entry point (argv without the program name); it is
// exercised directly by tests rather than shelling out to a built binary.
func run(args []string, stdout, stderr io.Writer) int {
	return runContext(context.Background(), args, stdout, stderr)
}

func runContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		logln(stderr, usage())
		return 2
	}

	switch args[0] {
	case "validate":
		return runValidate(args, stdout, stderr)
	case "boundary":
		return runBoundary(args, stdout, stderr)
	case "ownership":
		return runOwnership(args, stdout, stderr)
	case "gen":
		return runGen(args, stdout, stderr)
	case "lock-render":
		return runLockRenderContext(ctx, args, stdout, stderr)
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
	connectorgen validate [dir] [--json] [--connector <name>]   (default dir: internal/connectors/defs)
  connectorgen boundary [repo-root] [--json] [--base <ref>]
  connectorgen ownership [repo-root] [--json] [--base <ref>] [--scope-file <path>]
	connectorgen gen
	connectorgen lock-render <connector> [--defs <dir>] [--check]`
}

// runValidate implements `connectorgen validate [dir] [--json]`.
func runValidate(args []string, stdout, stderr io.Writer) int {
	dir := ""
	asJSON := false
	connector := ""
	connectorSet := false
	for index := 1; index < len(args); index++ {
		a := args[index]
		switch a {
		case "--json":
			asJSON = true
		case "--connector":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logf(stderr, "connectorgen validate: %s requires a value\n", a)
				return 2
			}
			index++
			value := args[index]
			if strings.TrimSpace(value) == "" {
				logf(stderr, "connectorgen validate: %s requires a non-empty value\n", a)
				return 2
			}
			if connectorSet {
				logln(stderr, "connectorgen validate: --connector may be specified only once")
				return 2
			}
			connector = value
			connectorSet = true
		default:
			if strings.HasPrefix(a, "-") {
				logf(stderr, "connectorgen validate: unknown flag %q\n", a)
				return 2
			}
			if dir != "" {
				logf(stderr, "connectorgen validate: unexpected extra argument %q\n", a)
				return 2
			}
			dir = a
		}
	}
	if connectorSet && !namePattern.MatchString(connector) {
		logf(stderr, "connectorgen validate: invalid connector name %q\n", connector)
		return 2
	}

	if dir == "" {
		root, err := repoRoot()
		if err != nil {
			logln(stderr, "connectorgen validate:", err)
			return 1
		}
		dir = filepath.Join(root, "internal/connectors/defs")
	}

	if connectorSet {
		if isBundleDir(dir) {
			if filepath.Base(filepath.Clean(dir)) != connector {
				logf(stderr, "connectorgen validate: connector %q does not match bundle directory %q\n", connector, filepath.Base(filepath.Clean(dir)))
				return 2
			}
		} else {
			dir = filepath.Join(dir, connector)
		}
	}
	report, err := validatePath(dir)
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

// validatePath validates either a defs root containing connector bundle
// directories or a single connector bundle directory. The latter keeps
// connector-local gates easy to run without accidentally treating schemas/ and
// fixtures/ as sibling connector bundles.
func validatePath(dir string) (Report, error) {
	cleanDir := filepath.Clean(dir)
	if isBundleDir(cleanDir) {
		absDir, err := filepath.Abs(cleanDir)
		if err != nil {
			return Report{}, fmt.Errorf("validate: resolve bundle dir: %w", err)
		}
		findings, warnings := validateBundleDir(os.DirFS(filepath.Dir(absDir)), filepath.Base(absDir))
		return Report{Findings: findings, Warnings: warnings, ConnectorsChecked: 1}, nil
	}
	return validateDir(os.DirFS(cleanDir))
}

func isBundleDir(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "metadata.json"))
	return err == nil && !info.IsDir()
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
// go.mod.
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
