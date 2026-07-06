// Package certify: pairing.go implements the create-then-cleanup write
// protocol's declarative WritePairing table, tag generation, and
// record_schema-driven data generation (design
// docs/architecture/connector-certification-design.md §C).
package certify

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// tagPrefix is the fixed lead-in for every certify-created record's tag
// (design §C: "pm-cert-<slug>-<runid8>-<unix-ts>").
const tagPrefix = "pm-cert-"

// NewTag builds a certify write-protocol tag for slug (typically the
// connector name) and runID8 (see NewRunID8), embedding the current UTC
// unix timestamp so the tag is both human-greppable and sortable by
// creation time (design §C tag convention).
func NewTag(slug, runID8 string) string {
	return fmt.Sprintf("%s%s-%s-%d", tagPrefix, slug, runID8, time.Now().UTC().Unix())
}

// NewRunID8 returns a random 8-character lowercase hex string suitable for
// embedding in a certify tag or a batch progress file name (design §C
// "<runid8>").
func NewRunID8() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand.Read failing is effectively unrecoverable on any real
		// platform; fall back to a time-derived value rather than panicking
		// so a certify run degrades rather than crashes.
		return fmt.Sprintf("%08x", uint32(time.Now().UnixNano()))
	}
	return hex.EncodeToString(buf)
}

// IsCertifyTag reports whether s looks like a tag NewTag would produce, for
// the sweeper's live-scan matching (design §C "optional live scan of
// VerifyStreams for aged pm-cert-<slug>-* tags").
func IsCertifyTag(s string) bool {
	return strings.HasPrefix(s, tagPrefix)
}

// WritePairing declares how one write action's create is safely paired with
// a cleanup action, per certification design §C. Zero-value Overrides means
// no per-connector overrides apply.
type WritePairing struct {
	Create       string
	Cleanup      string
	CleanupKind  string // delete | close | archive
	IDField      string // e.g. "number", "name"
	VerifyStream string // e.g. "issues", "labels"
	VerifyField  string // e.g. "title", "name"
	Overrides    map[string]any
}

// cleanupPrefixKinds maps a cleanup action's naming prefix to its
// CleanupKind, per design §C "Default pairing inference: create_X <->
// delete_X | close_X | archive_X". Order matters only in that every prefix
// is tried; ties cannot occur since action names use a single leading verb.
var cleanupPrefixKinds = []struct {
	prefix string
	kind   string
}{
	{"delete_", "delete"},
	{"close_", "close"},
	{"archive_", "archive"},
}

// InferPairing derives a WritePairing for createAction from the set of all
// action names available on the connector (design §C "Default pairing
// inference"), trying delete_X, then close_X, then archive_X where X is
// createAction with its own "create_" prefix stripped. Returns ok=false when
// createAction is not itself a create_-prefixed action, or no matching
// cleanup action exists among available — per design §C, "Unpaired mutating
// actions are never executed live".
func InferPairing(createAction string, available []string) (WritePairing, bool) {
	const createPrefix = "create_"
	if !strings.HasPrefix(createAction, createPrefix) {
		return WritePairing{}, false
	}
	suffix := strings.TrimPrefix(createAction, createPrefix)

	known := make(map[string]bool, len(available))
	for _, a := range available {
		known[a] = true
	}

	for _, ck := range cleanupPrefixKinds {
		candidate := ck.prefix + suffix
		if known[candidate] {
			return WritePairing{Create: createAction, Cleanup: candidate, CleanupKind: ck.kind}, true
		}
	}
	return WritePairing{}, false
}

// builtinPairings holds hand-curated, real (not merely inferred)
// WritePairing tables for connectors whose writes.json this task has
// verified end-to-end (design §C "Per-connector overrides in a declarative
// pairing table"). github's create_label/delete_label pair is keyed by the
// label's own "name" field (labels have no separate numeric ID — the name
// IS the identifier GitHub's API dereferences), giving cleanup_verify a
// clean "entity gone" check via a live read-back of the labels stream.
var builtinPairings = map[string][]WritePairing{
	"github": {
		{
			Create:       "create_label",
			Cleanup:      "delete_label",
			CleanupKind:  "delete",
			IDField:      "name",
			VerifyStream: "labels",
			VerifyField:  "name",
		},
		{
			Create:       "create_issue",
			Cleanup:      "close_issue",
			CleanupKind:  "close",
			IDField:      "number",
			VerifyStream: "issues",
			VerifyField:  "title",
		},
		{
			Create:       "create_milestone",
			Cleanup:      "delete_milestone",
			CleanupKind:  "delete",
			IDField:      "number",
			VerifyStream: "milestones",
			VerifyField:  "title",
		},
	},
}

// PairingsFor returns the known WritePairing entries for connector (empty if
// none are curated). Callers needing a single best pairing for a self-test
// or a --write run should take the first entry with a non-empty
// VerifyStream.
func PairingsFor(connector string) []WritePairing {
	return builtinPairings[connector]
}

// schemaDoc is the minimal subset of a JSON Schema (draft-07, per writes.json
// record_schema) this package needs to drive data generation (design §C
// "required-field heuristics by name").
type schemaDoc struct {
	Required   []string                  `json:"required"`
	Properties map[string]schemaProperty `json:"properties"`
}

type schemaProperty struct {
	Type string `json:"type"`
}

// GenerateRecord builds a minimal record satisfying schema's required
// fields using the certification design §C heuristics: name/title/label
// fields get tag verbatim; email fields get a deterministic
// pm-cert+<runid>@example.com address; url/homepage-shaped string fields get
// a deterministic https://example.com/pm-cert/<runid> value; any other
// required string field also gets tag (a safe default distinguishable in
// review); numeric fields get 1; boolean fields get false. Optional fields
// are left unset entirely.
func GenerateRecord(schemaJSON []byte, tag, runID string) (map[string]any, error) {
	return GenerateRecordWithOverrides(schemaJSON, tag, runID, nil)
}

// GenerateRecordWithOverrides is GenerateRecord plus a final overrides pass:
// any key present in overrides replaces the heuristic-generated value (or is
// added outright), per design §C "Per-connector overrides in a declarative
// pairing table".
func GenerateRecordWithOverrides(schemaJSON []byte, tag, runID string, overrides map[string]any) (map[string]any, error) {
	var doc schemaDoc
	if len(schemaJSON) > 0 {
		if err := json.Unmarshal(schemaJSON, &doc); err != nil {
			return nil, fmt.Errorf("certify: parse record_schema: %w", err)
		}
	}

	rec := make(map[string]any, len(doc.Required))
	for _, field := range doc.Required {
		prop := doc.Properties[field]
		rec[field] = generateFieldValue(field, prop.Type, tag, runID)
	}

	for k, v := range overrides {
		rec[k] = v
	}
	return rec, nil
}

// generateFieldValue applies the design §C required-field heuristics for a
// single field name/type pair.
func generateFieldValue(field, typ, tag, runID string) any {
	lower := strings.ToLower(field)
	switch {
	case typ == "integer" || typ == "number":
		return 1
	case typ == "boolean":
		return false
	case strings.Contains(lower, "email"):
		return "pm-cert+" + runID + "@example.com"
	case strings.Contains(lower, "url") || strings.Contains(lower, "homepage") || strings.Contains(lower, "link"):
		return "https://example.com/pm-cert/" + runID
	case strings.Contains(lower, "name") || strings.Contains(lower, "title") || strings.Contains(lower, "label"):
		return tag
	default:
		// Any other required string-ish field: default to the tag so it is
		// still identifiable and greppable in review, matching design §C's
		// intent that certify-created data must always be recognizable.
		return tag
	}
}
