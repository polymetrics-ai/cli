// Package wasabistatsapi implements the wasabi-stats-api bundle's AuthHook
// and RecordHook (conventions.md Tier-2 table), resolving the recorded
// AUTH_COMPLEX quarantine blocker (docs/migration/quarantine.json): legacy's
// requester() (internal/connectors/wasabi-stats-api/wasabi_stats_api.go:
// 129-143) branches auth mode on the CONTENTS of the api_key secret at
// runtime -- Bearer connsdk.APIKeyHeader is used unless
// strings.SplitN(key, ":", 2) yields exactly two parts, in which case it
// switches to connsdk.Basic(parts[0], parts[1]). The engine's `when` grammar
// (equality/membership/truthiness over a whole resolved reference,
// conventions.md Sec3) has no string-split/substring primitive to branch on
// a SUBSTRING SHAPE of a secret's own value, so this cannot be expressed as
// a declarative when-gated dual-auth candidate list.
//
// The RecordHook ports a second, unrelated legacy behavior that also has no
// single-template computed_fields equivalent: a per-record `id`-fallback
// derivation (first non-empty of bucket, then date, then the literal stream
// name) applied only when the raw record's own `id` is absent
// (wasabi_stats_api.go:115-117). Two hook interfaces total, at the Tier-2
// cap (conventions.md Sec1).
package wasabistatsapi

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("wasabi-stats-api", func() engine.Hooks { return Hooks{} })
}

// Hooks is the wasabi-stats-api hook set. It implements engine.AuthHook and
// engine.RecordHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "wasabi-stats-api" }

// Authenticator resolves the content-based Bearer-vs-Basic connsdk.
// Authenticator for spec (mode "custom", hook "wasabi-stats-api").
// spec.Value carries the templated api_key secret (API-CONTRACT.md's
// documented field-mapping convention -- wasabi's AuthSpec has no dedicated
// field for "a secret whose contents determine the auth mode", and Value is
// otherwise unused by the custom mode, mirroring gmail's reuse of Token for
// its refresh token).
func (Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	key, err := engine.Interpolate(spec.Value, authVars(cfg))
	if err != nil {
		return nil, fmt.Errorf("wasabi-stats-api custom auth: resolve api_key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("wasabi-stats-api custom auth: api_key is required")
	}

	// Exactly legacy's branch: strings.SplitN(key, ":", 2) yielding 2 parts
	// switches to Basic auth; anything else (zero ':' at all) stays Bearer.
	if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
		return connsdk.Basic(parts[0], parts[1]), nil
	}
	return connsdk.APIKeyHeader("Authorization", key, "Bearer "), nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// MapRecord implements engine.RecordHook, porting legacy's per-record `id`
// fallback (wasabi_stats_api.go:115-117) verbatim: when the RAW record has
// no `id` at all, derive one as the first non-empty value of raw["bucket"],
// then raw["date"], then the literal stream name -- applied only when
// projected["id"] is not already set (a record whose raw payload DID carry
// an `id` already survived ordinary schema projection untouched, matching
// legacy's `if item["id"] == nil` guard exactly). Every record is kept
// (keep=true always); this hook never drops records.
func (Hooks) MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
	if projected == nil {
		projected = connsdk.Record{}
	}
	if projected["id"] == nil {
		projected["id"] = firstNonEmpty(fmt.Sprint(raw["bucket"]), fmt.Sprint(raw["date"]), stream)
	}
	return projected, true, nil
}

// firstNonEmpty ports legacy's helper of the same name
// (wasabi_stats_api.go:200-207) verbatim: returns the first value that,
// once trimmed, is non-empty and not the literal string "<nil>" (fmt.Sprint's
// rendering of a Go nil interface value, produced here when raw["bucket"]/
// raw["date"] is itself absent).
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}
