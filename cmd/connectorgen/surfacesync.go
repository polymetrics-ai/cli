package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// surfaceSync derives the command-surface metadata that executable commands
// need in order to actually execute, from the bundle's own declarations.
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
//   - api_surface  <- any command's declared endpoint summary, independent of
//     its intent. A raw `METHOD /path` summary is an imported address, not a
//     human description; an optional write-action annotation is ignored. The
//     parsed endpoint must already appear in api_surface.json, so this is a
//     join across consistent files, never an invented endpoint. An
//     operation-backed endpoint remains an authoritative, more specific source
//     for direct_read, direct_write, binary_download, text_export, and
//     status_check.
//   - flags[].maps_to <- "path.<var>" when the flag's name matches a {var} in
//     that endpoint's path.
//   - flags[].required <- true when the mapped REST path parameter is declared
//     required. A caller cannot supply a required path input by another route.
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
//   - binary_download and text_export carry no output policy at all: the
//     response becomes a file, not a JSON body.
//   - rest.max_bytes <- defaultOperationRESTMaxBytes when unset or
//     non-positive, matching the direct operation executors' default. A
//     positive value is the operation's own declaration and is left alone.
//
// It never touches a command that is not implemented, and never invents an
// endpoint: a command must declare a parseable endpoint summary or an
// operation endpoint before api_surface can be synchronized.
const (
	// defaultDirectReadOutputPolicy mirrors engine.directReadPolicyJSONRedacted.
	defaultDirectReadOutputPolicy = "json_redacted"
	// defaultOperationRESTMaxBytes mirrors engine.defaultDirectReadMaxBytes.
	defaultOperationRESTMaxBytes = 1 << 20
)

var surfacePathVarRE = regexp.MustCompile(`\{([^}]+)\}`)

// surfaceSummaryEndpointRE identifies the source-material form used when a
// command has an endpoint but no operation-backed transport. The optional
// annotation names the write action that produced the row; it is not part of
// the provider address.
var surfaceSummaryEndpointRE = regexp.MustCompile(`^([A-Z]+) (\/[^[:space:]]+)(?: \([^)]*\))?$`)

// surfaceSyncFields counts one outcome across the synced fields.
type surfaceSyncFields struct {
	APISurface   int
	OutputPolicy int
	FlagMapsTo   int
	FlagRequired int
	FlagDerived  int
	MaxBytes     int
}

func (f surfaceSyncFields) total() int {
	return f.APISurface + f.OutputPolicy + f.FlagMapsTo + f.FlagRequired + f.FlagDerived + f.MaxBytes
}

func (f surfaceSyncFields) String() string {
	return fmt.Sprintf("api_surface=%d output_policy=%d flag_maps_to=%d flag_required=%d flag_derived=%d rest.max_bytes=%d",
		f.APISurface, f.OutputPolicy, f.FlagMapsTo, f.FlagRequired, f.FlagDerived, f.MaxBytes)
}

func (f *surfaceSyncFields) add(other surfaceSyncFields) {
	f.APISurface += other.APISurface
	f.OutputPolicy += other.OutputPolicy
	f.FlagMapsTo += other.FlagMapsTo
	f.FlagRequired += other.FlagRequired
	f.FlagDerived += other.FlagDerived
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
	sourceProjectionDrift := 0
	for _, name := range names {
		projection, err := syncCheckedInSourceProjection(filepath.Join(dir, name), name, check)
		if err != nil {
			logf(stderr, "connectorgen surface-sync: %s source projection: %v\n", name, err)
			return 1
		}
		if projection.Changed() {
			sourceProjectionDrift++
			verb := "updated"
			if check {
				verb = "would update"
			}
			logf(stdout, "%s: source projection %s writes=%d cli=%d\n", name, verb, projection.Writes, projection.CLI)
		}
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
	ledgerStats, err := syncRuntimeOperationEndpointLedger(dir, check)
	if err != nil {
		logf(stderr, "connectorgen surface-sync: runtime operation endpoint ledger: %v\n", err)
		return 1
	}
	if ledgerStats.Changed {
		verb := "updated"
		if check {
			verb = "would update"
		}
		logf(stdout, "runtime operation endpoint ledger: %s %d endpoint(s)\n", verb, ledgerStats.Entries)
	}

	if check && (total.total() > 0 || ledgerStats.Changed || sourceProjectionDrift > 0) {
		logf(stderr, "connectorgen surface-sync: %d connector(s) out of sync, %d field(s) missing, %d field(s) divergent, %d source projection(s) drifted, runtime endpoint ledger drift=%t; run `connectorgen surface-sync`\n",
			changed, total.Filled.total(), total.Corrected.total(), sourceProjectionDrift, ledgerStats.Changed)
		return 1
	}
	logf(stdout, "connectorgen surface-sync: %d connector(s) scanned, %d field(s) filled and %d field(s) corrected across %d connector(s)\n",
		len(names), total.Filled.total(), total.Corrected.total(), changed)
	return 0
}

func syncCheckedInSourceProjection(bundleDir, connector string, check bool) (sourceProjectionStats, error) {
	lockPath := filepath.Join(bundleDir, "sources", connector+"-operation-source-lock.json")
	lockRaw, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sourceProjectionStats{}, nil
		}
		return sourceProjectionStats{}, err
	}
	descriptorPath := filepath.Join(bundleDir, "sources", connector+"-operation-descriptor.json")
	raw, err := os.ReadFile(descriptorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sourceProjectionStats{}, fmt.Errorf("canonical source descriptor is missing")
		}
		return sourceProjectionStats{}, err
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(raw, &descriptor); err != nil {
		return sourceProjectionStats{}, fmt.Errorf("parse canonical source descriptor: %w", err)
	}
	if descriptor.SchemaVersion != 2 && descriptor.SchemaVersion != 3 {
		return sourceProjectionStats{}, fmt.Errorf("source descriptor schema_version = %d, want 2 or 3", descriptor.SchemaVersion)
	}
	// A source reference is evidence only. Validate its descriptor against the
	// exact lock-owned contract before passing it to a materializer, so a
	// removed gap or invented execution field cannot turn citation data into a
	// surface-sync input.
	reference, err := sourceImportLockDeclaresSourceReference(lockRaw)
	if err != nil {
		return sourceProjectionStats{}, fmt.Errorf("parse source lock reference header: %w", err)
	}
	if reference {
		lock, err := parseSourceImportLock(lockRaw, connector)
		if err != nil {
			return sourceProjectionStats{}, fmt.Errorf("parse source lock: %w", err)
		}
		if findings := validateSourceDescriptorAgainstLock(connector, filepath.ToSlash(filepath.Join("sources", connector+"-operation-descriptor.json")), lock, descriptor); len(findings) != 0 {
			return sourceProjectionStats{}, fmt.Errorf("validate canonical source descriptor: %s", findings[0].Message)
		}
	}
	return projectSourceDescriptorToBundle(bundleDir, sourceImportResult{
		Operations:     descriptor.Operations,
		InboundEvents:  descriptor.InboundEvents,
		Extensions:     descriptor.Extensions,
		GraphQLSchemas: descriptor.GraphQLSchemas,
	}, check)
}

type runtimeOperationEndpointLedgerStats struct {
	Entries int
	Changed bool
}

func syncRuntimeOperationEndpointLedger(dir string, check bool) (runtimeOperationEndpointLedgerStats, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return runtimeOperationEndpointLedgerStats{}, err
	}
	ledger := make(map[string][]engine.OperationEndpointLedgerEntry)
	sourceFS := withoutRuntimeOperationEndpointLedgerFS{FS: os.DirFS(dir)}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		bundle, err := engine.Load(sourceFS, name)
		if err != nil {
			return runtimeOperationEndpointLedgerStats{}, fmt.Errorf("load %s: %w", name, err)
		}
		if bundle.Surface == nil {
			return runtimeOperationEndpointLedgerStats{}, fmt.Errorf("bundle %s has no api_surface.json", name)
		}
		ledger[name] = engine.OperationDirectReadEndpointLedgerEntries(bundle)
	}
	stats := runtimeOperationEndpointLedgerStats{}
	for _, entries := range ledger {
		stats.Entries += len(entries)
	}
	raw, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return stats, err
	}
	raw = append(raw, '\n')
	path := filepath.Join(dir, engine.RuntimeOperationEndpointLedgerFile)
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, raw) {
		return stats, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return stats, err
	}
	stats.Changed = true
	if check {
		return stats, nil
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return stats, err
	}
	return stats, nil
}

type withoutRuntimeOperationEndpointLedgerFS struct {
	fs.FS
}

func (f withoutRuntimeOperationEndpointLedgerFS) Open(name string) (fs.File, error) {
	if name == engine.RuntimeOperationEndpointLedgerFile {
		return nil, fs.ErrNotExist
	}
	return f.FS.Open(name)
}

// syncBundle rewrites one bundle's cli_surface.json and operations.json in
// place. In check mode nothing is written and the stats report what would
// change.
func syncBundle(dir string, check bool) (surfaceSyncStats, error) {
	stats := surfaceSyncStats{}

	cliPath := filepath.Join(dir, "cli_surface.json")
	opsPath := filepath.Join(dir, "operations.json")
	apiSurfacePath := filepath.Join(dir, "api_surface.json")
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
	declaredEndpoints, err := declaredAPISurfaceEndpoints(apiSurfacePath)
	if err != nil {
		return stats, err
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
		if stringField(cmd, "availability") != "implemented" {
			continue
		}
		if method, endpointPath, ok := surfaceEndpointFromSummary(stringField(cmd, "summary")); ok &&
			declaredEndpoints[surfaceEndpointKey(method, endpointPath)] {
			synchronizeCommandAPISurface(cmd, method, endpointPath, &stats)
		}

		intent := stringField(cmd, "intent")
		if intent != "direct_read" && intent != "direct_write" && intent != "binary_download" && intent != "text_export" && intent != "status_check" {
			continue
		}
		if intent == "direct_read" {
			// Legacy surfaces can predate parameter import and still expose a
			// provider cursor directly. A direct read has one opaque cursor
			// channel: --page-cursor. Retaining a raw cursor lets callers bypass
			// the page-context contract, so it is derived drift rather than an
			// author choice. This applies even to legacy direct reads without an
			// operation-backed REST declaration. A size override is deliberately
			// not removed: it changes the page window rather than naming the next
			// page, and the engine measures completeness against the size actually
			// sent.
			stats.Corrected.FlagDerived += removeLegacyDirectReadCursorFlags(cmd)
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
		// a file download or export never borrows a REST one, so a mismatched
		// pair is left untouched for the validator to report.
		kind := stringField(op, "kind")
		blockName := "rest"
		if intent == "binary_download" || intent == "text_export" {
			blockName = "binary"
		}
		if (intent == "direct_read" && kind != "rest_read") ||
			(intent == "direct_write" && kind != "rest_write") ||
			(intent == "binary_download" && kind != "binary_download") ||
			(intent == "text_export" && kind != "text_export") ||
			(intent == "status_check" && kind != "rest_status") {
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

		// DERIVED: the operation endpoint is authoritative for operation-backed
		// commands, so it corrects an imported summary endpoint if they diverge.
		synchronizeCommandAPISurface(cmd, method, endpointPath, &stats)

		// A direct write's response policy is part of the operation's own
		// preview-bound contract, so it must exactly match. Direct reads use a
		// supported default and file downloads or exports carry no body policy.
		switch policy := strings.TrimSpace(stringField(cmd, "output_policy")); {
		case intent == "binary_download" || intent == "text_export":
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
		case intent == "status_check":
			if policy != "status" {
				cmd.set("output_policy", "status")
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

		pathVars := map[string]bool{}
		for _, match := range surfacePathVarRE.FindAllStringSubmatch(endpointPath, -1) {
			pathVars[match[1]] = true
		}
		for _, flagRaw := range arrayField(cmd, "flags") {
			flag, ok := flagRaw.(*orderedObject)
			if !ok {
				continue
			}
			name := strings.ReplaceAll(stringField(flag, "name"), "-", "_")
			if !pathVars[name] {
				// Flags that name no path variable map to query or body
				// targets the operation does not determine; leave them alone.
				continue
			}
			// DERIVED: a flag named after a path variable resolves that
			// variable. Any other target leaves the path unresolvable.
			want := "path." + name
			switch got := strings.TrimSpace(stringField(flag, "maps_to")); {
			case got == "":
				flag.set("maps_to", want)
				stats.Filled.FlagMapsTo++
			case got != want:
				flag.set("maps_to", want)
				stats.Corrected.FlagMapsTo++
			}
		}

		// DERIVED: fixed-operation command flags come from the operation's
		// imported parameter set, never from hand-authoring. This includes the
		// closed header.<declared-name> channel for REST, status, binary, and
		// text-export operations; a connector with no imported parameters is
		// unaffected.
		stats.Filled.FlagDerived += deriveCommandParameterFlags(cmd, block)

		// DERIVED: a REST path parameter marked required is an executable
		// command input, not a display preference. Its mapped CLI flag must
		// therefore be required before commandrunner attempts path
		// interpolation or creates a provider request. This intentionally
		// applies to every REST operation kind and every connector, while
		// leaving query/body requiredness to their own declared contracts.
		filled, corrected := deriveRequiredPathFlagRequiredness(cmd, block)
		stats.Filled.FlagRequired += filled
		stats.Corrected.FlagRequired += corrected
		filled, corrected = deriveRequiredHeaderFlagRequiredness(cmd, block)
		stats.Filled.FlagRequired += filled
		stats.Corrected.FlagRequired += corrected
		filled, corrected = deriveHeaderFlagRepeatability(cmd, block)
		stats.Filled.FlagDerived += filled
		stats.Corrected.FlagDerived += corrected

		// DEFAULTED for REST direct operations: a binary_download or text_export
		// operation must declare its own positive max_bytes at bundle load, and
		// a positive rest.max_bytes is the operation's own declaration rather
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

// surfaceEndpointFromSummary extracts the declared provider endpoint from the
// source-material summary form. Friendly human summaries intentionally do not
// match, so an ergonomic alias with no one-to-one endpoint remains unbound.
func surfaceEndpointFromSummary(summary string) (method, path string, ok bool) {
	parts := surfaceSummaryEndpointRE.FindStringSubmatch(strings.TrimSpace(summary))
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// declaredAPISurfaceEndpoints records the canonical endpoint set already
// declared by the bundle. A raw summary with sentence punctuation is not an
// endpoint declaration, so it cannot silently invent a binding merely because
// it resembles one.
func declaredAPISurfaceEndpoints(path string) (map[string]bool, error) {
	endpoints := map[string]bool{}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return endpoints, nil
	}
	if err != nil {
		return nil, fmt.Errorf("api_surface.json: %w", err)
	}
	var surface orderedJSON
	if err := json.Unmarshal(raw, &surface); err != nil {
		return nil, fmt.Errorf("api_surface.json: %w", err)
	}
	for _, rawEndpoint := range arrayField(surface.root, "endpoints") {
		endpoint, ok := rawEndpoint.(*orderedObject)
		if !ok {
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(stringField(endpoint, "method")))
		endpointPath := stringField(endpoint, "path")
		if method != "" && endpointPath != "" {
			endpoints[surfaceEndpointKey(method, endpointPath)] = true
		}
	}
	return endpoints, nil
}

// synchronizeCommandAPISurface applies one declaration-owned endpoint to a
// command. The schema allows exactly method and path on an entry, so replacing
// a divergent array loses no author-owned field.
func synchronizeCommandAPISurface(cmd *orderedObject, method, path string, stats *surfaceSyncStats) {
	existing := arrayField(cmd, "api_surface")
	switch {
	case len(existing) == 0:
		cmd.set("api_surface", derivedAPISurface(method, path))
		stats.Filled.APISurface++
	case len(existing) != 1 || !endpointMatches(existing[0], method, path):
		cmd.set("api_surface", derivedAPISurface(method, path))
		stats.Corrected.APISurface++
	}
}

// removeLegacyDirectReadCursorFlags removes old raw provider-cursor flags from
// one direct-read command. The parameter importer already omits these for
// newly imported operations; this keeps hand-authored pre-import surfaces
// honest as well.
//
// Only opaque cursor selectors are removed. Parameters such as page_size,
// limit, and offset remain because a caller may legitimately select a window
// size or explicit offset, and direct-read pagination reports that caller-owned
// position without inventing a --page number for it.
func removeLegacyDirectReadCursorFlags(cmd *orderedObject) int {
	flags := arrayField(cmd, "flags")
	if len(flags) == 0 {
		return 0
	}
	kept := make([]any, 0, len(flags))
	removed := 0
	for _, raw := range flags {
		flag, ok := raw.(*orderedObject)
		if ok && isLegacyDirectReadCursorFlag(flag) {
			removed++
			continue
		}
		kept = append(kept, raw)
	}
	if removed > 0 {
		cmd.set("flags", kept)
	}
	return removed
}

// isLegacyDirectReadCursorFlag identifies an opaque provider cursor by the
// mapped request field, with summary wording for ambiguous `after`/`before`
// names. The latter is intentionally semantic: GitHub notifications' `before`
// is a timestamp filter and must remain available, whereas an `after`
// described as a cursor must not become a second paging API.
func isLegacyDirectReadCursorFlag(flag *orderedObject) bool {
	target := strings.TrimSpace(stringField(flag, "maps_to"))
	var parameter string
	switch {
	case strings.HasPrefix(target, "query."):
		parameter = strings.TrimPrefix(target, "query.")
	case strings.HasPrefix(target, "body."):
		parameter = strings.TrimPrefix(target, "body.")
	default:
		return false
	}
	normalized := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(parameter)))
	switch normalized {
	case "cursor", "startcursor", "endcursor", "pagecursor", "nextcursor", "nextpagecursor",
		"pagetoken", "nextpagetoken", "continuationtoken", "nexttoken", "scrollid":
		return true
	case "after", "before":
		summary := strings.ToLower(stringField(flag, "summary"))
		return strings.Contains(summary, "cursor") || strings.Contains(summary, "page token") || strings.Contains(summary, "pagination token")
	default:
		return false
	}
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

// deriveCommandParameterFlags synchronizes provider-owned flag semantics from
// the operation declaration. Human-authored prose and flags without a
// path/query/header mapping remain authored; source-owned type, bounds,
// requiredness, repeatability, values, location, and deletion do not drift.
//
// Paging parameters never arrive here: params-import excludes them, because
// paging is answered by the connector's declared pagination spec through
// --page/--page-cursor.
func deriveCommandParameterFlags(cmd *orderedObject, rest *orderedObject) int {
	// An absent block means this operation has not been through the immutable
	// provider-parameter importer, so this synchronizer has no source-owned
	// authority over its existing command flags. An explicitly present empty
	// block is different: params-import verified the provider exposes no caller
	// parameters and stale path/query/header flags must be removed.
	if _, declared := rest.get("parameters"); !declared {
		return 0
	}
	params := arrayField(rest, "parameters")
	wanted := map[string]*orderedObject{}
	order := make([]string, 0, len(params))
	for _, raw := range params {
		param, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		name := stringField(param, "name")
		location := stringField(param, "in")
		if location != "query" && location != "path" && location != "header" {
			continue
		}
		flagName := strings.ReplaceAll(name, "_", "-")
		if location == "path" && stringField(param, "cli_name") != "" {
			// A declaration can name a safe runtime path placeholder differently
			// from the provider's resource identity. The alias remains closed:
			// maps_to still names the exact provider path parameter below.
			flagName = stringField(param, "cli_name")
		}
		if location == "header" {
			// Header names are case-insensitive and may overlap a path/query
			// parameter. The generated CLI flag carries its location explicitly,
			// while maps_to remains the exact provider header name.
			flagName = "header-" + strings.ToLower(strings.ReplaceAll(name, "_", "-"))
		}
		if name == "" {
			continue
		}
		flag := newOrderedObject()
		flag.set("name", flagName)
		flag.set("type", derivedFlagType(param))
		if summary := stringField(param, "summary"); summary != "" {
			flag.set("summary", summary)
		}
		if values := arrayField(param, "values"); len(values) > 0 {
			flag.set("values", append([]any(nil), values...))
		}
		// Path requiredness is deliberately synchronized by
		// deriveRequiredPathFlagRequiredness below. Keeping that one owner lets
		// surface-sync distinguish an absent required marker from a corrected
		// false one, and prevents parameter-semantic reconciliation from
		// silently swallowing that accounting boundary.
		if required, _ := param.get("required"); required == true && location != "path" {
			flag.set("required", true)
		}
		if repeatable, _ := param.get("repeatable"); location == "header" && repeatable == true {
			flag.set("repeatable", true)
		}
		if maxBytes, present := param.get("max_bytes"); present {
			flag.set("max_bytes", maxBytes)
		}
		if minimum, present := param.get("minimum"); present {
			flag.set("minimum", minimum)
		}
		if maximum, present := param.get("maximum"); present {
			flag.set("maximum", maximum)
		}
		flag.set("maps_to", location+"."+name)
		key := normalizedDerivedFlagName(flagName)
		wanted[key] = flag
		order = append(order, key)
	}

	changes := 0
	flags := make([]any, 0, len(arrayField(cmd, "flags"))+len(wanted))
	seen := map[string]bool{}
	for _, raw := range arrayField(cmd, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			flags = append(flags, raw)
			continue
		}
		key := normalizedDerivedFlagName(stringField(flag, "name"))
		derived, exists := wanted[key]
		mapping := stringField(flag, "maps_to")
		providerOwned := strings.HasPrefix(mapping, "query.") || strings.HasPrefix(mapping, "path.") || strings.HasPrefix(mapping, "header.")
		if !exists {
			if providerOwned {
				changes++
				continue
			}
			flags = append(flags, flag)
			continue
		}
		seen[key] = true
		if summary := stringField(flag, "summary"); summary != "" {
			derived.set("summary", summary)
		}
		// Existing requiredness is an accounting boundary: a new source-owned
		// flag receives the provider's required marker, while an existing flag
		// leaves an absent/false marker for the location-specific synchronizer
		// to classify as a fill or correction. This preserves the established
		// direct-read query contract and gives path/header repair one owner.
		if sourceRequired, _ := derived.get("required"); sourceRequired == true {
			if required, present := flag.get("required"); !present {
				derived.remove("required")
			} else if required == false {
				derived.set("required", false)
			}
		}
		// Preserve path requiredness exactly until the dedicated path
		// synchronizer accounts for it. That includes an already-correct true:
		// dropping it here and immediately adding it back below leaves files
		// unchanged but reports phantom generator drift on every --check.
		// False and absent markers still flow to the same synchronizer, which
		// records their correction/fill once and makes the next pass a no-op.
		if strings.HasPrefix(mapping, "path.") {
			if required, present := flag.get("required"); present && required == false {
				derived.set("required", false)
			} else if present && required == true {
				derived.set("required", true)
			}
		}
		if !orderedSemanticEqual(flag, derived) {
			changes++
		}
		flags = append(flags, derived)
	}
	for _, key := range order {
		if seen[key] {
			continue
		}
		flags = append(flags, wanted[key])
		changes++
	}
	if changes > 0 {
		cmd.set("flags", flags)
	}
	return changes
}

func normalizedDerivedFlagName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "-", "_"))
}

// deriveRequiredHeaderFlagRequiredness mirrors the path rule for the other
// caller-owned, fixed-operation input location. The engine independently
// rejects an absent required header before I/O; synchronizing it here makes
// generated help and the command parser explain the requirement earlier.
func deriveRequiredHeaderFlagRequiredness(cmd, block *orderedObject) (filled, corrected int) {
	required := map[string]bool{}
	for _, raw := range arrayField(block, "parameters") {
		parameter, ok := raw.(*orderedObject)
		if !ok || stringField(parameter, "in") != "header" {
			continue
		}
		if isRequired, _ := parameter.get("required"); isRequired == true {
			required[strings.ToLower(stringField(parameter, "name"))] = true
		}
	}
	if len(required) == 0 {
		return 0, 0
	}
	for _, raw := range arrayField(cmd, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		target, mapped := strings.CutPrefix(strings.TrimSpace(stringField(flag, "maps_to")), "header.")
		if !mapped || !required[strings.ToLower(target)] {
			continue
		}
		isRequired, present := flag.get("required")
		if isRequired == true {
			continue
		}
		flag.set("required", true)
		if present {
			corrected++
		} else {
			filled++
		}
	}
	return filled, corrected
}

func deriveHeaderFlagRepeatability(cmd, block *orderedObject) (filled, corrected int) {
	repeatable := map[string]bool{}
	for _, raw := range arrayField(block, "parameters") {
		parameter, ok := raw.(*orderedObject)
		if !ok || stringField(parameter, "in") != "header" {
			continue
		}
		if declared, _ := parameter.get("repeatable"); declared == true {
			repeatable[strings.ToLower(strings.ReplaceAll(stringField(parameter, "name"), "_", "-"))] = true
		}
	}
	for _, raw := range arrayField(cmd, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		target, mapped := strings.CutPrefix(strings.TrimSpace(stringField(flag, "maps_to")), "header.")
		if !mapped {
			continue
		}
		want := repeatable[strings.ToLower(strings.ReplaceAll(target, "_", "-"))]
		got, present := flag.get("repeatable")
		if want {
			if got == true {
				continue
			}
			flag.set("repeatable", true)
			if present {
				corrected++
			} else {
				filled++
			}
			continue
		}
		if present {
			flag.remove("repeatable")
			corrected++
		}
	}
	return filled, corrected
}

// deriveRequiredPathFlagRequiredness synchronizes requiredness for flags that
// resolve REST path parameters. A required path cannot be supplied by any
// other command input, so leaving its mapped flag optional defers a caller
// mistake until interpolation or provider I/O. Query and body parameters have
// distinct contracts and are deliberately not changed here.
func deriveRequiredPathFlagRequiredness(cmd, rest *orderedObject) (filled, corrected int) {
	requiredPathParameters := map[string]bool{}
	for _, raw := range arrayField(rest, "parameters") {
		parameter, ok := raw.(*orderedObject)
		if !ok || stringField(parameter, "in") != "path" {
			continue
		}
		required, _ := parameter.get("required")
		if required == true {
			requiredPathParameters[stringField(parameter, "name")] = true
		}
	}
	if len(requiredPathParameters) == 0 {
		return 0, 0
	}

	for _, raw := range arrayField(cmd, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		parameter, mapped := strings.CutPrefix(strings.TrimSpace(stringField(flag, "maps_to")), "path.")
		if !mapped || !requiredPathParameters[parameter] {
			continue
		}
		required, present := flag.get("required")
		if required == true {
			continue
		}
		flag.set("required", true)
		if present {
			corrected++
		} else {
			filled++
		}
	}
	return filled, corrected
}

// derivedFlagType maps a provider parameter type onto the command surface's own
// flag vocabulary. An enum is expressed as type "enum" plus its values, which
// is the form the runtime already validates before any network call.
func derivedFlagType(param *orderedObject) string {
	if len(arrayField(param, "values")) > 0 {
		return "enum"
	}
	switch stringField(param, "type") {
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "string_array"
	default:
		return "string"
	}
}
