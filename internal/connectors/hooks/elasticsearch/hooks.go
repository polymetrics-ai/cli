// Package elasticsearch implements the elasticsearch bundle's AuthHook and
// RecordHook (conventions.md Tier-2 table), the Tier-2 escape hatch for two
// unrelated gaps the declarative dialect cannot express on its own:
//
//  1. Auth: legacy's requester()/auth() (internal/connectors/elasticsearch/
//     elasticsearch.go:178-198) sends "Authorization: ApiKey
//     base64(apiKeyId:apiKeySecret)" when both an API key id and secret are
//     configured. The engine's api_key_header AuthSpec mode sets a single
//     already-resolved Value verbatim (no encoding step at all), and its
//     base64 filter (interpolate.go) only ever base64-encodes the resolved
//     text of ONE {{ }} reference in isolation — resolveExpr processes each
//     {{ }} occurrence independently, so there is no declarative way to
//     concatenate two separate config/secret references with a literal ":"
//     and base64-encode the JOINED result in a single template. This is the
//     exact mechanics connsdk.Basic already performs (base64(user:pass)) but
//     with a "Basic " prefix hardcoded; elasticsearch needs the identical
//     encoding under an "ApiKey " prefix instead, which no AuthSpec field
//     exposes. ENGINE_GAP: composite-secret base64 header encoding.
//  2. Records: legacy's mapHit (elasticsearch.go:142-153) flattens every key
//     of the raw hit's _source object up to the TOP LEVEL of the emitted
//     record, then stamps `id` from the hit's _id — dropping _index/_score/
//     _id themselves. Schema-mode projection only copies exact top-level key
//     matches (no nested-object flatten primitive exists in the dialect at
//     all — RecordsSpec.KeyedObject explodes a keyed OBJECT of records, a
//     different shape), and passthrough mode would keep _source nested
//     rather than spread. RecordHook.MapRecord ports the flatten+id-stamp
//     verbatim.
//
// Two hook interfaces total, at the Tier-2 cap (conventions.md §1). Auth is
// resolved via a "custom" AuthSpec candidate ordered FIRST in streams.json's
// base.auth list (api_key_id present) with a "when"-gated declarative basic
// candidate second (username present) and an unconditional "none" candidate
// last — reproducing legacy's exact api-key-then-basic-then-none precedence
// (auth() checks apiKeyId+secret first, then username, else nil).
package elasticsearch

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("elasticsearch", func() engine.Hooks { return Hooks{} })
}

// Hooks is the elasticsearch hook set. It implements engine.AuthHook and
// engine.RecordHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "elasticsearch" }

// Authenticator resolves the composite API-key connsdk.Authenticator for
// spec (mode "custom", hook "elasticsearch"). spec.Value carries the
// templated config.api_key_id (streams.json's base.auth entry), reused here
// purely as the "this candidate's when already proved api_key_id is set"
// signal — the actual id/secret values are read directly from cfg so this
// hook, not AuthSpec, owns the two-value base64(id:secret) composition
// (mirrors wasabi-stats-api's reuse of the Value field for a
// custom-auth-only purpose per API-CONTRACT.md's documented convention).
func (Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	id := firstNonEmpty(cfg.Config["api_key_id"], cfg.Config["apiKeyId"])
	secret := firstNonEmpty(cfg.Secrets["api_key_secret"], cfg.Secrets["apiKeySecret"])
	if id != "" && secret != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(id + ":" + secret))
		return connsdk.APIKeyHeader("Authorization", encoded, "ApiKey "), nil
	}
	if username := strings.TrimSpace(cfg.Config["username"]); username != "" {
		return connsdk.Basic(username, cfg.Secrets["password"]), nil
	}
	return connsdk.AuthFunc(func(context.Context, *http.Request) error { return nil }), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// MapRecord implements engine.RecordHook for the "documents" stream only
// (the "indices" stream's _cat/indices rows need no flattening and pass
// this hook through untouched via keep=true, projected unchanged). Ports
// elasticsearch.go's mapHit verbatim: every key of the raw hit's _source
// object is spread onto the projected record's top level, then `id` is
// stamped from the raw hit's `_id` field — matching legacy's exact
// field set (no _index/_score/_id survive).
func (Hooks) MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
	if stream != "documents" {
		return projected, true, nil
	}
	if projected == nil {
		projected = connsdk.Record{}
	}
	if src, ok := raw["_source"].(map[string]any); ok {
		for k, v := range src {
			projected[k] = v
		}
	} else if src, ok := raw["_source"].(connsdk.Record); ok {
		for k, v := range src {
			projected[k] = v
		}
	}
	if id, ok := raw["_id"].(string); ok {
		projected["id"] = id
	}
	return projected, true, nil
}
