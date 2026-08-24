// Command connectorgen is the declarative connector engine tooling CLI
// (design §C.3):
//
//	validate [dir] [--json]   loads and validates every bundle under dir
//	                           (default internal/connectors/defs), exit 1 on
//	                           any finding
//	boundary [repo] [--json] [--base <ref>]
//	                           scans shared Go for connector-specific policy
//	                           outside definition-owned locations
//	ownership [repo] [--json] [--base <ref>] [--scope-file <path>]
//	                           validates changed paths for exactly one target connector
//	gen                        regenerates hooks/hookset/hookset_gen.go and
//	                           native/nativeset/nativeset_gen.go
//	surface-sync [dir] [--check]
//	                           derives operation-backed command metadata
//	                           (api_surface, output_policy, flag maps_to,
//	                           rest.max_bytes) and the compact runtime
//	                           direct-read endpoint ledger from bundle sources
//	source-import <connector> --out <path> [--defs <dir>] [--check]
//	                           verifies a connector-owned source lock and
//	                           emits canonical provider contracts for later
//	                           declaration materializers
//	source-retain <connector> --retrieved-at <RFC3339> --license <text> --terms <text>
//	                           explicitly obtains, verifies, and stores source
//	                           bytes using the lock-selected identity
//	surface-reconcile [dir] [--check] [--json] [--reason-contains text]
//	                           derives direct-read api_surface coverage and
//	                           blocked reasons from runtime preflight
//	batch plan --ledger <path> --out <path>
//	                           turns provider-artifact ledger evidence into a
//	                           deterministic, reviewable connector batch
//	batch materialize --manifest <path> --source-defs-root <path> ...
//	                           copies a reviewed source bundle and derives its
//	                           cited provider-artifact inventory and CLI surface
//	batch gate --manifest <path> --report <path>
//	                           records independent candidate validation and
//	                           runtime-preflight results
//	new <name>                 scaffolds internal/connectors/defs/<name>/
//
// It owns bundle validation plus generated hook/native import sets for the
// connector-architecture-v2 runtime.
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
	case "boundary":
		return runBoundary(args, stdout, stderr)
	case "ownership":
		return runOwnership(args, stdout, stderr)
	case "gen":
		return runGen(args, stdout, stderr)
	case "surface-sync":
		return runSurfaceSync(args, stdout, stderr)
	case "params-import":
		return runParamsImport(args, stdout, stderr)
	case "source-import":
		return runSourceImport(args, stdout, stderr)
	case "source-retain":
		return runSourceRetain(args, stdout, stderr)
	case "evidence-gate":
		return runEvidenceGate(args, stdout, stderr)
	case "surface-reconcile":
		return runSurfaceReconcile(args, stdout, stderr)
	case "certification-matrix":
		return runCertificationMatrix(args, stdout, stderr)
	case "certification-sweep":
		return runCertificationSweep(args, stdout, stderr)
	case "certification-candidates":
		return runCertificationCandidates(args, stdout, stderr)
	case "certification-subject":
		return runCertificationSubject(args, stdout, stderr)
	case "certification-evidence":
		return runCertificationEvidence(args, stdout, stderr)
	case "operation-evidence":
		return runOperationEvidence(args, stdout, stderr)
	case "batch":
		return runBatch(args, stdout, stderr)
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
  connectorgen boundary [repo-root] [--json] [--base <ref>]
  connectorgen ownership [repo-root] [--json] [--base <ref>] [--scope-file <path>]
	connectorgen gen
	connectorgen surface-sync [dir] [--check]  (default dir: internal/connectors/defs)
	connectorgen source-import <connector> --out <path> [--defs <dir>] [--check]
	connectorgen source-retain <connector> [--defs <dir>] --retrieved-at <RFC3339> --license <text> --terms <text>
	connectorgen evidence-gate <evidence-manifest.json> <TDD-LEDGER.md> <REVIEW.md>
	connectorgen surface-reconcile [dir] [--check] [--json] [--reason-contains text]  (default dir: internal/connectors/defs)
	connectorgen certification-matrix [repo-root] (--connector <name> [--check] | --all | --check)
	connectorgen certification-sweep [repo-root] --connector <name> [--check]
	connectorgen certification-candidates [repo-root] --connector <name> [--check]
	connectorgen certification-subject [repo-root] [--pm <built-pm>] [--check]
	connectorgen certification-evidence (transport|change-capture) --connector <name> --report <path> --binary-sha <sha256> --from-env password=<ENV> --run-id <id> --record-prefix <id> [--repo-root <path>]
	connectorgen certification-evidence report --connector <name> --report <path> --external-proof <path> --record-prefix <id> [--repo-root <path>]
	connectorgen certification-evidence draft --draft <.tmp/live-certification/drafts/record.json> [--repo-root <path>]
	connectorgen operation-evidence [repo-root] [--output <path>] [--fixed-100 <path>] [--check] [--write-fixed-100]
	connectorgen batch plan --ledger <path> --out <path> [--size <1-40>] [--connector <name>] [--min-operations <n>] [--max-operations <n>]
  connectorgen batch materialize --manifest <path> --source-defs-root <path> --retrieved-at <YYYY-MM-DD> --report <path> [--defs-root <path>] [--artifact-dir <path>] [--connector <name>]
  connectorgen batch gate --manifest <path> --report <path> [--defs-root <path>] [--connector <name>]
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
