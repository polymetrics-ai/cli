package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// surfaceSync derives the command-surface metadata that operation-backed
// direct_read commands need in order to actually execute, from the bundle's
// own operations.json.
//
// It exists because the same facts previously had to be hand-copied into
// cli_surface.json, and hand-copying is what let 174 commands claim
// availability "implemented" while the runtime blocked every one of them. A
// derivation cannot drift from its source the way a copy can, and --check
// turns that guarantee into a CI gate.
//
// It fills only what is derivable or governed by a documented default:
//
//   - api_surface  <- the operation's rest.method + rest.path. The endpoint is
//     already tracked in api_surface.json, whose covered_by names this exact
//     command, so this is a join across three consistent files, never an
//     invented endpoint.
//   - flags[].maps_to <- "path.<var>" when the flag's name matches a {var} in
//     the operation's rest.path.
//   - output_policy <- defaultDirectReadOutputPolicy when unset. Operations
//     declare "json", which is not a legal direct-read policy in any layer;
//     the redacting variant is the only supported analogue.
//   - rest.max_bytes <- defaultOperationRESTMaxBytes when unset, matching the
//     engine's own defaultDirectReadMaxBytes.
//
// It never edits a value that is already present, never touches a command that
// is not an implemented operation-backed direct_read, and never invents an
// endpoint for an operation that does not declare one.
const (
	// defaultDirectReadOutputPolicy mirrors engine.directReadPolicyJSONRedacted.
	defaultDirectReadOutputPolicy = "json_redacted"
	// defaultOperationRESTMaxBytes mirrors engine.defaultDirectReadMaxBytes.
	defaultOperationRESTMaxBytes = 1 << 20
)

var surfacePathVarRE = regexp.MustCompile(`\{([^}]+)\}`)

type surfaceSyncStats struct {
	APISurface   int
	OutputPolicy int
	FlagMapsTo   int
	MaxBytes     int
}

func (s surfaceSyncStats) total() int {
	return s.APISurface + s.OutputPolicy + s.FlagMapsTo + s.MaxBytes
}

func runSurfaceSync(args []string, stdout, stderr io.Writer) int {
	dir := filepath.Join("internal", "connectors", "defs")
	check := false
	for _, arg := range args[1:] {
		switch {
		case arg == "--check":
			check = true
		case strings.HasPrefix(arg, "-"):
			logf(stderr, "connectorgen surface-sync: unknown flag %q\n", arg)
			return 2
		default:
			dir = arg
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		logf(stderr, "connectorgen surface-sync: %v\n", err)
		return 1
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	total := surfaceSyncStats{}
	changed := 0
	for _, name := range names {
		stats, err := syncBundle(filepath.Join(dir, name), check)
		if err != nil {
			logf(stderr, "connectorgen surface-sync: %s: %v\n", name, err)
			return 1
		}
		if stats.total() == 0 {
			continue
		}
		changed++
		total.APISurface += stats.APISurface
		total.OutputPolicy += stats.OutputPolicy
		total.FlagMapsTo += stats.FlagMapsTo
		total.MaxBytes += stats.MaxBytes
		verb := "updated"
		if check {
			verb = "would update"
		}
		logf(stdout, "%s: %s api_surface=%d output_policy=%d flag_maps_to=%d rest.max_bytes=%d\n",
			name, verb, stats.APISurface, stats.OutputPolicy, stats.FlagMapsTo, stats.MaxBytes)
	}

	if check && total.total() > 0 {
		logf(stderr, "connectorgen surface-sync: %d connector(s) out of sync, %d field(s) missing; run `connectorgen surface-sync`\n", changed, total.total())
		return 1
	}
	logf(stdout, "connectorgen surface-sync: %d connector(s) scanned, %d field(s) filled across %d connector(s)\n", len(names), total.total(), changed)
	return 0
}

// syncBundle rewrites one bundle's cli_surface.json and operations.json in
// place. In check mode nothing is written and the stats report what would
// change.
func syncBundle(dir string, check bool) (surfaceSyncStats, error) {
	stats := surfaceSyncStats{}

	cliPath := filepath.Join(dir, "cli_surface.json")
	opsPath := filepath.Join(dir, "operations.json")
	cliRaw, err := os.ReadFile(cliPath)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return stats, err
	}
	opsRaw, err := os.ReadFile(opsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}
		return stats, err
	}

	var cli orderedJSON
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		return stats, fmt.Errorf("cli_surface.json: %w", err)
	}
	var ops orderedJSON
	if err := json.Unmarshal(opsRaw, &ops); err != nil {
		return stats, fmt.Errorf("operations.json: %w", err)
	}

	opsByID := map[string]*orderedObject{}
	for _, raw := range arrayField(ops.root, "operations") {
		op, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		if id := stringField(op, "id"); id != "" {
			opsByID[id] = op
		}
	}

	for _, raw := range arrayField(cli.root, "commands") {
		cmd, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		if stringField(cmd, "intent") != "direct_read" ||
			stringField(cmd, "availability") != "implemented" {
			continue
		}
		operation := stringField(cmd, "operation")
		if operation == "" {
			continue
		}
		op := opsByID[operation]
		if op == nil {
			continue
		}
		// Only rest_read operations are executable as direct reads; a
		// binary_download operation gets no invented REST endpoint.
		if stringField(op, "kind") != "rest_read" {
			continue
		}
		restRaw, _ := op.get("rest")
		rest, _ := restRaw.(*orderedObject)
		if rest == nil {
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(stringField(rest, "method")))
		restPath := stringField(rest, "path")
		if method == "" || restPath == "" {
			continue
		}

		if len(arrayField(cmd, "api_surface")) == 0 {
			endpoint := newOrderedObject()
			endpoint.set("method", method)
			endpoint.set("path", restPath)
			cmd.set("api_surface", []any{endpoint})
			stats.APISurface++
		}
		if stringField(cmd, "output_policy") == "" {
			cmd.set("output_policy", defaultDirectReadOutputPolicy)
			stats.OutputPolicy++
		}

		pathVars := map[string]bool{}
		for _, match := range surfacePathVarRE.FindAllStringSubmatch(restPath, -1) {
			pathVars[match[1]] = true
		}
		for _, flagRaw := range arrayField(cmd, "flags") {
			flag, ok := flagRaw.(*orderedObject)
			if !ok || strings.TrimSpace(stringField(flag, "maps_to")) != "" {
				continue
			}
			name := strings.ReplaceAll(stringField(flag, "name"), "-", "_")
			if pathVars[name] {
				flag.set("maps_to", "path."+name)
				stats.FlagMapsTo++
			}
		}

		if maxBytes, _ := rest.get("max_bytes"); !positiveNumber(maxBytes) {
			rest.set("max_bytes", json.Number(fmt.Sprint(defaultOperationRESTMaxBytes)))
			stats.MaxBytes++
		}
	}

	if stats.total() == 0 || check {
		return stats, nil
	}
	if stats.APISurface+stats.OutputPolicy+stats.FlagMapsTo > 0 {
		if err := writeBundleJSON(cliPath, cli, cliRaw); err != nil {
			return stats, err
		}
	}
	if stats.MaxBytes > 0 {
		if err := writeBundleJSON(opsPath, ops, opsRaw); err != nil {
			return stats, err
		}
	}
	return stats, nil
}

// writeBundleJSON re-renders a bundle file with the repository's 2-space
// indentation, preserving whether the original ended with a newline so the
// diff stays limited to the fields that actually changed.
func writeBundleJSON(path string, doc orderedJSON, original []byte) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(doc); err != nil {
		return err
	}
	out := buf.Bytes()
	if !bytes.HasSuffix(original, []byte("\n")) {
		out = bytes.TrimRight(out, "\n")
	}
	return os.WriteFile(path, out, 0o644)
}

func stringField(o *orderedObject, key string) string {
	if o == nil {
		return ""
	}
	v, _ := o.get(key)
	s, _ := v.(string)
	return s
}

func arrayField(o *orderedObject, key string) []any {
	if o == nil {
		return nil
	}
	v, _ := o.get(key)
	a, _ := v.([]any)
	return a
}

// positiveNumber reports whether a decoded JSON value is a number greater than
// zero. Values arrive as json.Number because the decoder preserves exact
// integer formatting.
func positiveNumber(v any) bool {
	number, ok := v.(json.Number)
	if !ok {
		return false
	}
	f, err := number.Float64()
	return err == nil && f > 0
}
