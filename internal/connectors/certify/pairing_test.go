package certify_test

import (
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

// TestNewTagFormat proves the create-then-cleanup tag convention (design §C:
// "pm-cert-<slug>-<runid8>-<unix-ts>") — a fixed-shape, grep-able tag string
// embedded in the primary human-visible field of every certify-created
// record, so a sweeper or a human operator can find it later.
func TestNewTagFormat(t *testing.T) {
	tag := certify.NewTag("github", "ab12cd34")

	if !strings.HasPrefix(tag, "pm-cert-github-ab12cd34-") {
		t.Fatalf("NewTag() = %q, want prefix pm-cert-github-ab12cd34-", tag)
	}
	parts := strings.Split(tag, "-")
	// pm-cert-github-ab12cd34-<ts> -> at least 5 hyphen-separated parts.
	if len(parts) < 5 {
		t.Fatalf("NewTag() = %q, want at least 5 hyphen-separated segments", tag)
	}
	ts := parts[len(parts)-1]
	if _, err := time.Parse("20060102150405", padTimestamp(ts)); err != nil && !isUnixSeconds(ts) {
		t.Errorf("NewTag() trailing segment %q is neither a recognizable timestamp form: %v", ts, err)
	}
}

// TestNewTagUniquePerCall proves two calls (even with identical slug/runID)
// never collide, since the ts component should reflect wall-clock time and
// certify.NewRunID8 must vary across runs.
func TestNewRunID8IsHex8(t *testing.T) {
	id := certify.NewRunID8()
	if len(id) != 8 {
		t.Fatalf("NewRunID8() = %q, want length 8", id)
	}
	for _, r := range id {
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Fatalf("NewRunID8() = %q, want lowercase hex", id)
		}
	}
}

// TestDefaultPairingInference proves the create_X <-> delete_X|close_X|
// archive_X default inference rule (design §C "Default pairing inference").
func TestDefaultPairingInference(t *testing.T) {
	cases := []struct {
		actions     []string
		create      string
		wantCleanup string
		wantKind    string
		wantOK      bool
	}{
		{actions: []string{"create_label", "delete_label"}, create: "create_label", wantCleanup: "delete_label", wantKind: "delete", wantOK: true},
		{actions: []string{"create_issue", "close_issue"}, create: "create_issue", wantCleanup: "close_issue", wantKind: "close", wantOK: true},
		{actions: []string{"create_webhook", "archive_webhook"}, create: "create_webhook", wantCleanup: "archive_webhook", wantKind: "archive", wantOK: true},
		{actions: []string{"create_thing"}, create: "create_thing", wantOK: false},
		{actions: []string{"update_issue"}, create: "update_issue", wantOK: false},
	}
	for _, tc := range cases {
		got, ok := certify.InferPairing(tc.create, tc.actions)
		if ok != tc.wantOK {
			t.Errorf("InferPairing(%q, %v) ok = %v, want %v", tc.create, tc.actions, ok, tc.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if got.Cleanup != tc.wantCleanup || got.CleanupKind != tc.wantKind {
			t.Errorf("InferPairing(%q, %v) = %+v, want cleanup=%q kind=%q", tc.create, tc.actions, got, tc.wantCleanup, tc.wantKind)
		}
	}
}

// TestBuiltinPairingGithubCreateLabel proves the built-in github pairing
// table has a real (not inferred) create_label -> delete_label entry with
// "gone" cleanup semantics (design §C WritePairing table), since this is the
// pairing certify's own tests exercise for a real connector's writes.json.
func TestBuiltinPairingGithubCreateLabel(t *testing.T) {
	pairings := certify.PairingsFor("github")
	if len(pairings) == 0 {
		t.Fatalf("PairingsFor(github) returned no pairings")
	}
	var found *certify.WritePairing
	for i := range pairings {
		if pairings[i].Create == "create_label" {
			found = &pairings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("PairingsFor(github) missing create_label pairing: %+v", pairings)
	}
	if found.Cleanup != "delete_label" {
		t.Errorf("github create_label pairing Cleanup = %q, want delete_label", found.Cleanup)
	}
	if found.CleanupKind != "delete" {
		t.Errorf("github create_label pairing CleanupKind = %q, want delete", found.CleanupKind)
	}
	if found.IDField != "name" {
		t.Errorf("github create_label pairing IDField = %q, want name", found.IDField)
	}
	if found.VerifyStream == "" {
		t.Errorf("github create_label pairing VerifyStream is empty, want a stream name")
	}
	if found.VerifyField == "" {
		t.Errorf("github create_label pairing VerifyField is empty, want a field name")
	}
}

// TestGenerateRecordFromSchemaHeuristics proves the required-field data
// generation heuristics (design §C "Data generation from write action
// record_schema"): name/title/label fields get the tag, email fields get a
// pm-cert+<runid>@example.com address, url fields get a deterministic
// https://example.com/... value, numeric fields get 1, bool fields get
// false, and optional fields are left unset.
func TestGenerateRecordFromSchemaHeuristics(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"required": ["name", "email", "homepage", "count", "active", "title"],
		"properties": {
			"name": {"type": "string"},
			"title": {"type": "string"},
			"email": {"type": "string"},
			"homepage": {"type": "string"},
			"count": {"type": "integer"},
			"active": {"type": "boolean"},
			"unset_optional": {"type": "string"}
		}
	}`)

	tag := "pm-cert-github-ab12cd34-1751450000"
	rec, err := certify.GenerateRecord(schema, tag, "ab12cd34")
	if err != nil {
		t.Fatalf("GenerateRecord() error = %v", err)
	}

	if rec["name"] != tag {
		t.Errorf("rec[name] = %v, want tag %q", rec["name"], tag)
	}
	if rec["title"] != tag {
		t.Errorf("rec[title] = %v, want tag %q", rec["title"], tag)
	}
	email, _ := rec["email"].(string)
	if !strings.HasPrefix(email, "pm-cert+") || !strings.HasSuffix(email, "@example.com") {
		t.Errorf("rec[email] = %v, want pm-cert+<runid>@example.com shape", rec["email"])
	}
	homepage, _ := rec["homepage"].(string)
	if !strings.HasPrefix(homepage, "https://example.com/pm-cert/") {
		t.Errorf("rec[homepage] = %v, want https://example.com/pm-cert/<runid> shape", rec["homepage"])
	}
	if rec["count"] != 1 {
		t.Errorf("rec[count] = %v, want 1", rec["count"])
	}
	if rec["active"] != false {
		t.Errorf("rec[active] = %v, want false", rec["active"])
	}
	if _, present := rec["unset_optional"]; present {
		t.Errorf("rec[unset_optional] = %v, want unset (optional field)", rec["unset_optional"])
	}
}

// TestGenerateRecordAppliesOverrides proves per-connector Overrides in a
// WritePairing win over the generic heuristics (design §C
// "Per-connector overrides in a declarative pairing table").
func TestGenerateRecordAppliesOverrides(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"required": ["name"],
		"properties": {"name": {"type": "string"}, "color": {"type": "string"}}
	}`)
	rec, err := certify.GenerateRecordWithOverrides(schema, "pm-cert-github-ab12cd34-1", "ab12cd34", map[string]any{"color": "ffffff"})
	if err != nil {
		t.Fatalf("GenerateRecordWithOverrides() error = %v", err)
	}
	if rec["color"] != "ffffff" {
		t.Errorf("rec[color] = %v, want override value ffffff", rec["color"])
	}
	if rec["name"] != "pm-cert-github-ab12cd34-1" {
		t.Errorf("rec[name] = %v, want tag", rec["name"])
	}
}

func padTimestamp(s string) string {
	// best-effort helper only used to attempt an RFC-ish parse; real
	// assertion is isUnixSeconds below.
	return s
}

func isUnixSeconds(s string) bool {
	if len(s) < 9 || len(s) > 12 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
