package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

// surface-reconcile makes api_surface operation rows accountable to the same
// runtime preflight that guards `availability: implemented`. It deliberately
// does not infer coverage from a command declaration: only a command that
// matches the endpoint and passes commandrunner.Preflight can replace an
// operation row with covered_by.direct_read.
//
// A row without such a command is still a known blocked fact, so the tool
// derives its reason from the current command state. Unknown operation models,
// malformed source shapes, and failed bundle loading are refusals or errors;
// they are never rewritten into plausible prose.

type surfaceReconcileStats struct {
	Scanned   int `json:"scanned"`
	Covered   int `json:"covered"`
	Blocked   int `json:"blocked"`
	Unchanged int `json:"unchanged"`
	Refused   int `json:"refused"`
}

func (s surfaceReconcileStats) changed() int {
	return s.Covered + s.Blocked
}

func (s *surfaceReconcileStats) add(other surfaceReconcileStats) {
	s.Scanned += other.Scanned
	s.Covered += other.Covered
	s.Blocked += other.Blocked
	s.Unchanged += other.Unchanged
	s.Refused += other.Refused
}

type surfaceReconcileReport struct {
	Connector string                `json:"connector"`
	Stats     surfaceReconcileStats `json:"stats"`
}

func runSurfaceReconcile(args []string, stdout, stderr io.Writer) int {
	dir := filepath.Join("internal", "connectors", "defs")
	check := false
	asJSON := false
	reasonContains := ""
	dirSet := false
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			logln(stdout, surfaceReconcileUsage())
			return 0
		case "--check":
			check = true
		case "--json":
			asJSON = true
		case "--reason-contains":
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
				logln(stderr, "connectorgen surface-reconcile: --reason-contains requires text")
				return 2
			}
			i++
			reasonContains = args[i]
		default:
			if strings.HasPrefix(arg, "-") {
				logf(stderr, "connectorgen surface-reconcile: unknown flag %q\n", arg)
				return 2
			}
			if dirSet {
				logf(stderr, "connectorgen surface-reconcile: unexpected extra argument %q\n", arg)
				return 2
			}
			dir = arg
			dirSet = true
		}
	}

	bundleDirs, err := surfaceReconcileBundleDirs(dir)
	if err != nil {
		logf(stderr, "connectorgen surface-reconcile: %v\n", err)
		return 1
	}

	reports := make([]surfaceReconcileReport, 0, len(bundleDirs))
	total := surfaceReconcileStats{}
	for _, bundleDir := range bundleDirs {
		stats, err := reconcileBundle(bundleDir, check, reasonContains)
		if err != nil {
			logf(stderr, "connectorgen surface-reconcile: %s: %v\n", filepath.Base(bundleDir), err)
			return 1
		}
		if stats.Scanned > 0 {
			reports = append(reports, surfaceReconcileReport{Connector: filepath.Base(bundleDir), Stats: stats})
		}
		total.add(stats)
	}

	if asJSON {
		payload := struct {
			ConnectorsScanned int                      `json:"connectors_scanned"`
			Bundles           []surfaceReconcileReport `json:"bundles"`
			Total             surfaceReconcileStats    `json:"total"`
		}{ConnectorsScanned: len(bundleDirs), Bundles: reports, Total: total}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(payload); err != nil {
			logf(stderr, "connectorgen surface-reconcile: encode report: %v\n", err)
			return 1
		}
	} else {
		for _, report := range reports {
			verb := "reconciled"
			if check {
				verb = "would reconcile"
			}
			logf(stdout, "%s: %s covered=%d blocked=%d unchanged=%d refused=%d\n",
				report.Connector, verb, report.Stats.Covered, report.Stats.Blocked,
				report.Stats.Unchanged, report.Stats.Refused)
		}
		logf(stdout, "connectorgen surface-reconcile: %d connector(s) scanned; covered=%d blocked=%d unchanged=%d refused=%d\n",
			len(bundleDirs), total.Covered, total.Blocked, total.Unchanged, total.Refused)
	}

	if check && total.changed() > 0 {
		logf(stderr, "connectorgen surface-reconcile: %d row(s) need deterministic reclassification; rerun without --check to apply\n", total.changed())
		return 1
	}
	return 0
}

func surfaceReconcileUsage() string {
	return `usage: connectorgen surface-reconcile [dir] [--check] [--json] [--reason-contains text]

Derive direct-read api_surface coverage and blocked reasons from real command
runtime preflight. In --check mode, report pending reclassifications without
writing files and exit 1 when a row would change.`
}

func surfaceReconcileBundleDirs(dir string) ([]string, error) {
	clean := filepath.Clean(dir)
	if isBundleDir(clean) {
		return []string{clean}, nil
	}
	entries, err := os.ReadDir(clean)
	if err != nil {
		return nil, err
	}
	bundles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(clean, entry.Name())
		if isBundleDir(candidate) {
			bundles = append(bundles, candidate)
		}
	}
	sort.Strings(bundles)
	return bundles, nil
}

// reconcileBundle reclassifies only operation-model direct_read ledger rows.
// It loads the disk bundle through the engine and calls commandrunner.Preflight
// against that engine connector, so a generated covered_by value has the same
// admission proof as the CLI path. In check mode it mutates only the decoded
// in-memory document and reports the prospective changes.
func reconcileBundle(dir string, check bool, reasonContains string) (surfaceReconcileStats, error) {
	stats := surfaceReconcileStats{}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return stats, err
	}
	bundle, err := engine.Load(os.DirFS(filepath.Dir(absDir)), filepath.Base(absDir))
	if err != nil {
		return stats, fmt.Errorf("load runtime bundle: %w", err)
	}
	if bundle.Surface == nil || bundle.CLISurface == nil {
		return stats, nil
	}

	apiPath := filepath.Join(absDir, "api_surface.json")
	raw, err := os.ReadFile(apiPath)
	if err != nil {
		return stats, err
	}
	var surface orderedJSON
	if err := json.Unmarshal(raw, &surface); err != nil {
		return stats, fmt.Errorf("api_surface.json: %w", err)
	}
	endpoints := arrayField(surface.root, "endpoints")
	if len(endpoints) != len(bundle.Surface.Endpoints) {
		return stats, fmt.Errorf("api_surface.json endpoint count %d does not match loaded surface count %d", len(endpoints), len(bundle.Surface.Endpoints))
	}
	connector := engine.New(bundle, nil)
	changed := false
	for i, rawEndpoint := range endpoints {
		endpoint, ok := rawEndpoint.(*orderedObject)
		if !ok {
			return stats, fmt.Errorf("api_surface endpoint %d is not an object", i+1)
		}
		operationRaw, hasOperation := endpoint.get("operation")
		if !hasOperation {
			continue
		}
		operation, ok := operationRaw.(*orderedObject)
		if !ok {
			stats.Refused++
			continue
		}
		if reasonContains != "" && !strings.Contains(stringField(operation, "reason"), reasonContains) {
			continue
		}
		stats.Scanned++
		if stringField(operation, "model") != "direct_read" {
			stats.Refused++
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(stringField(endpoint, "method")))
		path := stringField(endpoint, "path")
		if method == "" || path == "" {
			stats.Refused++
			continue
		}

		passing, reasons := directReadCandidates(connector, bundle.CLISurface.Commands, method, path)
		if len(passing) > 0 {
			endpoint.remove("operation")
			endpoint.set("covered_by", directReadCoverage(passing))
			stats.Covered++
			changed = true
			continue
		}

		reason := blockedDirectReadReason(method, path, reasons)
		rowChanged := false
		if stringField(operation, "status") != "blocked" {
			operation.set("status", "blocked")
			rowChanged = true
		}
		if blockedByDefault, _ := operation.get("blocked_by_default"); blockedByDefault != true {
			operation.set("blocked_by_default", true)
			rowChanged = true
		}
		if stringField(operation, "reason") != reason {
			operation.set("reason", reason)
			rowChanged = true
		}
		if rowChanged {
			changed = true
			stats.Blocked++
		} else {
			stats.Unchanged++
		}
	}
	if !changed || check {
		return stats, nil
	}
	if err := writeBundleJSON(apiPath, surface, raw); err != nil {
		return stats, err
	}
	return stats, nil
}

// directReadCandidates returns the commands that really reach one endpoint and
// a deterministic set of current reasons when none can. Preflight is run only
// for implemented candidates; a planned command is a useful blocked reason but
// cannot be treated as reachable.
func directReadCandidates(connector connectors.Connector, commands []engine.CLICommand, method, path string) ([]string, []string) {
	type candidate struct {
		path         string
		availability string
		reason       string
		passes       bool
	}
	candidates := make([]candidate, 0)
	for _, command := range commands {
		if command.Intent != "direct_read" || !commandMatchesEndpoint(command, method, path) {
			continue
		}
		entry := candidate{path: command.Path, availability: command.Availability}
		if command.Availability == "implemented" {
			if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err == nil {
				entry.passes = true
			} else {
				entry.reason = runtimePreflightReason(err)
			}
		}
		candidates = append(candidates, entry)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].path < candidates[j].path })
	passing := make([]string, 0, len(candidates))
	reasons := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.passes {
			passing = append(passing, candidate.path)
			continue
		}
		if candidate.availability != "implemented" {
			reasons = append(reasons, fmt.Sprintf("%q is declared %s, not implemented", candidate.path, candidate.availability))
			continue
		}
		reasons = append(reasons, fmt.Sprintf("%q fails runtime preflight: %s", candidate.path, candidate.reason))
	}
	return passing, reasons
}

func commandMatchesEndpoint(command engine.CLICommand, method, path string) bool {
	return len(command.APISurface) == 1 &&
		strings.EqualFold(strings.TrimSpace(command.APISurface[0].Method), method) &&
		command.APISurface[0].Path == path
}

func runtimePreflightReason(err error) string {
	var blocked *commandrunner.BlockedCommandError
	if errors.As(err, &blocked) && strings.TrimSpace(blocked.Reason) != "" {
		return blocked.Reason
	}
	return err.Error()
}

func directReadCoverage(commands []string) *orderedObject {
	coverage := newOrderedObject()
	if len(commands) == 1 {
		coverage.set("direct_read", commands[0])
		return coverage
	}
	values := make([]any, len(commands))
	for i, command := range commands {
		values[i] = command
	}
	coverage.set("direct_reads", values)
	return coverage
}

func blockedDirectReadReason(method, path string, reasons []string) string {
	if len(reasons) == 0 {
		return fmt.Sprintf("No reachable direct-read command declares %s %s.", method, path)
	}
	return "No reachable direct-read command: " + strings.Join(reasons, "; ") + "."
}
