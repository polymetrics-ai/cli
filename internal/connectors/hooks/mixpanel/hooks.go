// Package mixpanel implements the Tier-2 hook set for the mixpanel migration
// (docs/migration/conventions.md §1's Tier-2 hooks table), porting
// internal/connectors/mixpanel (read-only reference, unchanged) exactly.
//
// Two hook interfaces are implemented, at the Tier-2 cap:
//
//   - AuthHook resolves HTTP Basic auth with two fully INDEPENDENT
//     credential fallback chains (username: config value, else a secret,
//     else empty; password: one secret, else another, else empty) that the
//     declarative when-gated auth-candidate dialect cannot express — when
//     supports only one truthiness/equality/membership condition per
//     candidate, with no AND/OR combinator, so a candidate list cannot gate
//     on "this specific username source AND this specific password source"
//     together without risking the wrong candidate firing first for some
//     legacy-valid combination (see defs/mixpanel/docs.md's Auth setup).
//   - RecordHook resolves the engage stream's multi-source field fallbacks
//     (distinct_id/email/created), the same "no coalesce filter in this
//     dialect" shape, ported from legacy's mixpanelProfileRecord.
package mixpanel

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("mixpanel", func() engine.Hooks { return New() })
}

// Hooks is the mixpanel hook set. It implements engine.AuthHook and
// engine.RecordHook only.
type Hooks struct{}

// New returns a fresh mixpanel Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "mixpanel" }

// Authenticator resolves HTTP Basic auth for spec (mode "custom", hook
// "mixpanel"), reading straight from cfg's raw config/secrets maps (NOT
// engine.Interpolate, which hard-errors on an absent key) to reproduce
// legacy's mixpanelCredentials fallback exactly: username is cfg.Config
// ["username"], else cfg.Secrets["username_secret"], else "" (basic auth
// permits an empty username); password is cfg.Secrets["password"], else
// cfg.Secrets["api_secret"], else "".
func (h *Hooks) Authenticator(_ context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	username, password := credentials(cfg)
	if strings.TrimSpace(username) == "" && strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("mixpanel: requires config username or secret username_secret, and secret password or api_secret")
	}
	return connsdk.Basic(username, password), nil
}

// credentials mirrors legacy's mixpanelCredentials (mixpanel/mixpanel.go)
// field-for-field.
func credentials(cfg connectors.RuntimeConfig) (string, string) {
	username := cfg.Config["username"]
	if username == "" {
		username = cfg.Secrets["username_secret"]
	}
	password := cfg.Secrets["password"]
	if password == "" {
		password = cfg.Secrets["api_secret"]
	}
	return username, password
}

// MapRecord post-processes the engage stream's multi-source field
// fallbacks; every other stream (cohorts/annotations) is a same-named,
// direct schema projection needing no hook involvement, so it passes
// projected through unchanged.
func (h *Hooks) MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
	if stream != "engage" {
		return projected, true, nil
	}
	props, _ := raw["$properties"].(map[string]any)
	out := connsdk.Record{
		"distinct_id": first(raw["$distinct_id"], raw["distinct_id"]),
		"email":       first(mapGet(props, "$email"), raw["$email"], raw["email"]),
		"created":     first(mapGet(props, "$created"), raw["created"]),
	}
	return out, true, nil
}

// mapGet safely reads key from m, tolerating a nil map (props may be absent
// or not itself an object on a malformed/partial record).
func mapGet(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}

// first mirrors legacy's first(...) helper (mixpanel/mixpanel.go) exactly:
// returns the first value that is neither nil nor an empty/whitespace-only
// string.
func first(values ...any) any {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		if value != nil {
			return value
		}
	}
	return nil
}
