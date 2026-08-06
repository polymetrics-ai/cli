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
// direct_read, direct_write, and binary_download commands need in order to
// actually execute, from the bundle's own operations.json.
//
// It exists because the same facts previously had to be hand-copied into
// cli_surface.json, and hand-copying is what let 174 commands claim
// availability "implemented" while the runtime blocked every one of them. A
// derivation cannot drift from its source the way a copy can, and --check
// turns that guarantee into a CI gate.
//
// Two field classes, with different rules, because "derived" and "defaulted"
// are not the same guarantee:
//
// DERIVED — the operation is the only source of truth, so a present value that
// disagrees with it is drift, not a choice. These are compared, not merely
// filled, or a hand-edited api_surface would pass --check clean while help,
// docs and the website advertised an endpoint the executor never calls:
//
//   - api_surface  <- the operation's endpoint method + path, taken from the
//     block its kind declares (rest for direct_read and direct_write, binary
//     for binary_download). The endpoint is already tracked in
//     api_surface.json, so this is a join across consistent files, never an
//     invented endpoint.
//   - flags[].maps_to <- "path.<var>" when the flag's name matches a {var} in
//     that endpoint's path, or "identifier_set.<name>" when it matches an
//     operation-declared caller-supplied identifier set. The latter takes
//     precedence for path_segment sets: its single item is still an explicit
//     set rather than an ordinary path parameter.
//
// DEFAULTED — the bundle author's value wins; only an absent or unusable one is
// replaced:
//
//   - direct_read output_policy <- defaultDirectReadOutputPolicy when unset or
//     unsupported. Operations declare "json", which is not a legal
//     direct-read policy in any layer; the redacting variant is the only
//     supported analogue.
//   - direct_write output_policy <- operations[].output_policy exactly. The
//     response contract is bound into its preview digest, so this is derived,
//     not a display preference.
//   - a binary_download carries no output policy at all: the response becomes
//     a file, not a JSON body.
//   - rest.max_bytes <- defaultOperationRESTMaxBytes when unset or
//     non-positive, matching the direct operation executors' default. A
//     positive value is the operation's own declaration and is left alone.
//
// It never touches a command that is not an implemented operation-backed
// direct_read, direct_write, or binary_download, and never invents an endpoint
// for an operation that does not declare one.
const (
	// defaultDirectReadOutputPolicy mirrors engine.directReadPolicyJSONRedacted.
	defaultDirectReadOutputPolicy = "json_redacted"
	// defaultOperationRESTMaxBytes mirrors engine.defaultDirectReadMaxBytes.
	defaultOperationRESTMaxBytes = 1 << 20
)

var surfacePathVarRE = regexp.MustCompile(`\{([^}]+)\}`)

// surfaceSyncFields counts one outcome across the four synced fields.
type surfaceSyncFields struct {
	APISurface   int
	OutputPolicy int
	FlagMapsTo   int
	MaxBytes     int
}

func (f surfaceSyncFields) total() int {
	return f.APISurface + f.OutputPolicy + f.FlagMapsTo + f.MaxBytes
}

func (f surfaceSyncFields) String() string {
	return fmt.Sprintf("api_surface=%d output_policy=%d flag_maps_to=%d rest.max_bytes=%d",
		f.APISurface, f.OutputPolicy, f.FlagMapsTo, f.MaxBytes)
}

func (f *surfaceSyncFields) add(other surfaceSyncFields) {
	f.APISurface += other.APISurface
	f.OutputPolicy += other.OutputPolicy
	f.FlagMapsTo += other.FlagMapsTo
	f.MaxBytes += other.MaxBytes
}

// surfaceSyncStats separates a field that was ABSENT and got filled from one
// that was PRESENT and disagreed with its source. Collapsing the two would let
// the tool report a clean fill while it silently rewrote a hand-edited value.
type surfaceSyncStats struct {
	Filled    surfaceSyncFields
	Corrected surfaceSyncFields
}

func (s surfaceSyncStats) total() int {
	return s.Filled.total() + s.Corrected.total()
}

// cliFieldTotal counts the fields that live in cli_surface.json, so the writer
// only rewrites the file it actually changed.
func (s surfaceSyncStats) cliFieldTotal() int {
	return s.total() - s.Filled.MaxBytes - s.Corrected.MaxBytes
}

func (s surfaceSyncStats) opsFieldTotal() int {
	return s.Filled.MaxBytes + s.Corrected.MaxBytes
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
		total.Filled.add(stats.Filled)
		total.Corrected.add(stats.Corrected)
		verb := "updated"
		if check {
			verb = "would update"
		}
		logf(stdout, "%s: %s filled %s; corrected %s\n", name, verb, stats.Filled, stats.Corrected)
	}

	if check && total.total() > 0 {
		logf(stderr, "connectorgen surface-sync: %d connector(s) out of sync, %d field(s) missing and %d field(s) divergent; run `connectorgen surface-sync`\n",
			changed, total.Filled.total(), total.Corrected.total())
		return 1
	}
	logf(stdout, "connectorgen surface-sync: %d connector(s) scanned, %d field(s) filled and %d field(s) corrected across %d connector(s)\n",
		len(names), total.Filled.total(), total.Corrected.total(), changed)
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
		intent := stringField(cmd, "intent")
		if intent != "direct_read" && intent != "direct_write" && intent != "binary_download" {
			continue
		}
		if stringField(cmd, "availability") != "implemented" {
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
		// The endpoint block is whichever one the operation's kind declares.
		// A direct operation never borrows a binary operation's endpoint and
		// a binary download never borrows a REST one, so a mismatched pair is
		// left untouched for the validator to report.
		kind := stringField(op, "kind")
		blockName := "rest"
		if intent == "binary_download" {
			blockName = "binary"
		}
		if (intent == "direct_read" && kind != "rest_read") ||
			(intent == "direct_write" && kind != "rest_write") ||
			(intent == "binary_download" && kind != "binary_download") {
			continue
		}
		blockRaw, _ := op.get(blockName)
		block, _ := blockRaw.(*orderedObject)
		if block == nil {
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(stringField(block, "method")))
		endpointPath := stringField(block, "path")
		if method == "" || endpointPath == "" {
			continue
		}

		// DERIVED: the operation's endpoint is the only source, so a present
		// api_surface that disagrees with it is drift and gets replaced. The
		// schema allows exactly method and path on an entry, so replacing the
		// whole array loses nothing an author could have written.
		existing := arrayField(cmd, "api_surface")
		switch {
		case len(existing) == 0:
			cmd.set("api_surface", derivedAPISurface(method, endpointPath))
			stats.Filled.APISurface++
		case len(existing) != 1 || !endpointMatches(existing[0], method, endpointPath):
			cmd.set("api_surface", derivedAPISurface(method, endpointPath))
			stats.Corrected.APISurface++
		}

		// A direct write's response policy is part of the operation's own
		// preview-bound contract, so it must exactly match. Direct reads use a
		// supported default and binary downloads carry no body policy.
		switch policy := strings.TrimSpace(stringField(cmd, "output_policy")); {
		case intent == "binary_download":
			if cmd.remove("output_policy") {
				stats.Corrected.OutputPolicy++
			}
		case intent == "direct_write":
			want := stringField(op, "output_policy")
			switch {
			case policy == "":
				cmd.set("output_policy", want)
				stats.Filled.OutputPolicy++
			case policy != want:
				cmd.set("output_policy", want)
				stats.Corrected.OutputPolicy++
			}
		case policy == "":
			cmd.set("output_policy", defaultDirectReadOutputPolicy)
			stats.Filled.OutputPolicy++
		case !directReadOutputPolicies[policy]:
			// A policy no layer supports is not a deliberate choice; the
			// runtime refuses it. A supported one is left exactly as authored.
			cmd.set("output_policy", defaultDirectReadOutputPolicy)
			stats.Corrected.OutputPolicy++
		}

		identifierSets := map[string]bool{}
		if intent == "direct_read" {
			for _, setRaw := range arrayField(block, "caller_supplied_identifier_sets") {
				set, ok := setRaw.(*orderedObject)
				if !ok {
					continue
				}
				if name := stringField(set, "name"); name != "" {
					identifierSets[name] = true
				}
			}
		}
		pathVars := map[string]bool{}
		for _, match := range surfacePathVarRE.FindAllStringSubmatch(endpointPath, -1) {
			pathVars[match[1]] = true
		}
		for _, flagRaw := range arrayField(cmd, "flags") {
			flag, ok := flagRaw.(*orderedObject)
			if !ok {
				continue
			}
			flagName := stringField(flag, "name")
			want := ""
			if identifierSets[flagName] {
				want = "identifier_set." + flagName
			} else if name := strings.ReplaceAll(flagName, "-", "_"); pathVars[name] {
				want = "path." + name
			}
			if want == "" {
				// Flags that name no path variable map to query or body
				// targets the operation does not determine; leave them alone.
				continue
			}
			// DERIVED: an operation-declared identifier set or a path variable
			// is owned by the operation contract. Any other target either drops
			// the caller input or leaves the endpoint unresolvable.
			rawMapsTo := stringField(flag, "maps_to")
			switch got := strings.TrimSpace(rawMapsTo); {
			case got == "":
				flag.set("maps_to", want)
				stats.Filled.FlagMapsTo++
			case got != want || rawMapsTo != got:
				flag.set("maps_to", want)
				stats.Corrected.FlagMapsTo++
			}
		}

		// DEFAULTED for REST direct operations: a binary_download operation
		// must declare its own positive max_bytes at bundle load, and a
		// positive rest.max_bytes is the operation's own declaration rather
		// than anything this tool derives.
		if intent == "direct_read" || intent == "direct_write" {
			maxBytes, present := block.get("max_bytes")
			if !positiveNumber(maxBytes) {
				block.set("max_bytes", json.Number(fmt.Sprint(defaultOperationRESTMaxBytes)))
				if present {
					stats.Corrected.MaxBytes++
				} else {
					stats.Filled.MaxBytes++
				}
			}
		}
	}

	if stats.total() == 0 || check {
		return stats, nil
	}
	if stats.cliFieldTotal() > 0 {
		if err := writeBundleJSON(cliPath, cli, cliRaw); err != nil {
			return stats, err
		}
	}
	if stats.opsFieldTotal() > 0 {
		if err := writeBundleJSON(opsPath, ops, opsRaw); err != nil {
			return stats, err
		}
	}
	return stats, nil
}

// derivedAPISurface builds the single-endpoint api_surface an operation-backed
// command implies.
func derivedAPISurface(method, path string) []any {
	endpoint := newOrderedObject()
	endpoint.set("method", method)
	endpoint.set("path", path)
	return []any{endpoint}
}

// endpointMatches reports whether an existing api_surface entry already states
// exactly the operation's endpoint. Method comparison is case-insensitive
// because that is how every consumer reads it; path comparison is exact,
// because a path is a template the executor substitutes into.
func endpointMatches(raw any, method, path string) bool {
	endpoint, ok := raw.(*orderedObject)
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(stringField(endpoint, "method")), method) &&
		stringField(endpoint, "path") == path
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
