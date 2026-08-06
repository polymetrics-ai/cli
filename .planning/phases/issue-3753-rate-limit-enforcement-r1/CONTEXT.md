# Context — issue-3753-rate-limit-enforcement-r1

## Contract and inline GSD fallback

- Parent foundation: #3750. Delivery slices: #3753 (resolver attachment) and #3754 (scoped registries).
- The repository currently has no production `rate_limits.json`; test-only bundles are required and `defs/defs.go` remains unchanged.
- `scripts/gsd doctor`, all five `scripts/gsd sources` lookups, and `go run ./cmd/agentcontractgen check` passed on 2026-08-06.
- `scripts/gsd prompt discuss-phase 3753` and `scripts/gsd prompt plan-phase 3753 --tdd` were generated and followed inline. Pi interactive execution is unavailable and the canonical single-worker contract prohibits role spawning.

## Fixed implementation decisions

1. A declared policy is selected only when all selector dimensions match: `all` or the declared method/path pair, plus optional non-secret `config.tier` and `config.auth_type`. Unknown, not-applicable, absent, and non-matching declarations do not attach a new limiter.
2. A policy scope must name its non-secret config key through `scope.subject_config`. The raw config value is used only as the transient input to `connectors.CoordinationIdentity.RateScopeKey`; only its opaque projection reaches the registry. `subject_kind` maps to the established `connectors.RateScopeKind` vocabulary and an unsupported kind is refused.
3. Matching policies compose conservatively: each matching budget is admitted before one logical requester send. The registry key contains connector name, policy ID, and opaque scope key; credential revision, secret maps, raw bindings, and raw subjects are never read or stored.
4. The always-available registry is process-local. This slice provides its registry/limiter seam only; it does not add #3755 operator output or claim cross-process protection while a coordinator is unavailable.
5. `streams.json.base.rate_limit` remains the legacy page-loop limiter. A matching `rate_limits.json` policy is an additional requester-level admission; it never creates or replaces a legacy limiter for an undeclared/non-matching request.

## Changed-path scope

- `internal/coordination/`: opaque local registry and deterministic clock seam.
- `internal/connectors/connsdk/`: typed actual-cost observation, without raw headers/bodies.
- `internal/connectors/engine/`: resolver, runtime wiring for check/read/writes/direct/binary requesters, and legacy precedence.
- `docs/migration/conventions.md`: enforced declaration semantics and precedence.
- `internal/connectors/engine/testdata/`: test-only declared bundles. No production connector bundle or embed change.

## Exclusions

- No database limits, shared daemon lifecycle, generic HTTP/SQL tools, production declarations, credentials, runtime subjects in output, or #3755 output/help/UI work.
