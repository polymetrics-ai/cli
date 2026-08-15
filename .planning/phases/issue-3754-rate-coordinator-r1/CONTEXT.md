# Context — issue #3754 local and optional shared rate-limit scope registries

## Contract and inline GSD fallback

- Issue #3754 is the Wave-1 coordination foundation under #3855. It consumes #3863's
  secret-free `connectors.CoordinationIdentity` and #3752/#3753's declared rate-limit
  admission seam; it does not alter a credential, provider authentication, connector
  declaration, fencing, parking, resumption, or transport route.
- `scripts/gsd doctor`, all required `scripts/gsd sources` lookups, and
  `go run ./cmd/agentcontractgen check` passed on 2026-08-14. The generated
  `discuss-phase 3754 --auto` prompt cannot resolve a GSD numerical phase:
  `gsd-sdk query init.phase-op 3754` reports `phase_found: false`. This issue directory
  is therefore the required inline/manual GSD fallback. Compatible isolated roles are
  unavailable and the canonical single-worker contract forbids role spawning.
- The approved autonomous decisions below are the `--auto` discussion record. They are
  bounded by the issue and launch brief, which explicitly prohibit connector-specific
  logic, credential use, generic execution paths, and #3867 parking/resumption.

## Locked decisions

1. `RateLimitPolicy` declares a coordination policy. Its zero/absent value is
   `process_local`; only an explicit `require_shared` value may select external
   coordination. Runtime configuration can provide a coordinator endpoint, but it can
   never upgrade a local policy or inherit shared enforcement into another policy.
2. The always-present registry remains process-local and is labelled exactly as such:
   it coordinates one `pm` process only and makes no account-wide or cross-process
   claim. The user-visible inspect/status projection is limited to this provenance;
   it never renders an opaque key, a binding, a subject, config values, or credentials.
3. The optional coordinator is DragonflyDB/Redis-compatible, ephemeral, atomic, and
   uses server time plus expiry/reset TTLs. Its only key material is the existing
   `(connector, policy ID, connectors.RateLimitScopeKey)` tuple. A raw subject is
   transient only inside `CoordinationIdentity.RateScopeKey`; a credential is not an
   input anywhere in this slice.
4. A `require_shared` policy fails closed when no shared registry is configured or the
   coordinator is unreachable. The result is an exported typed error with a stable
   reason/component field naming the missing shared coordinator, while omitting
   protected identity material. It must never fall back to the local limiter.
5. Shared admission supports the declared budget models atomically and reports an
   explicit unavailable/unsupported typed reason rather than approximating a policy.
   Provider reset observations may tighten shared state but this issue does not park,
   persist a run, schedule a retry, or resume anything (#3867).
6. Production connector bundles remain untouched. Test bundles exercise both modes;
   GitHub REST/GraphQL budget authoring remains #3990. The scope is connector-neutral.

## Scope and evidence boundary

- Owned production paths: `internal/coordination/**`, `internal/connectors/connsdk/**`,
  `internal/connectors/engine/**`, invocation-scoped configuration/wiring only where
  needed to make explicit shared policy reachable, plus targeted CLI inspect/help/docs
  parity if that wiring adds a public projection.
- Required tests include atomic grant/block, cancellation, process-local provenance,
  `require_shared` no-fallback refusal, opaque-key absence, and a real two-process
  shared-budget integration test gated on an explicit local Dragonfly endpoint.
- The six protected surfaces are argv, environment, files, logs, receipts, and delivery
  evidence. Tests use only non-secret synthetic public scope subjects and assert that
  neither raw subject nor opaque identity is returned by status/error/output. No test
  reads or prints process environment values, credentials, vault records, or headers.

## Canonical references

- `AGENTS.md` — lifecycle, safety, verification, and direct-read contracts.
- `docs/architecture/runtime-dependencies.md` — Dragonfly is ephemeral coordination only.
- `docs/runtime/SETUP.md` — optional local runtime endpoints and lifecycle.
- `.planning/phases/issue-3753-rate-limit-enforcement-r1/{CONTEXT.md,PLAN.md,TDD-LEDGER.md}` —
  admission/resolver seam consumed here.
- `.planning/phases/issue-3863-secret-free-coordination-identity-r1/{CONTEXT.md,PLAN.md,TDD-LEDGER.md}` —
  mandatory opaque identity contract.
- `internal/connectors/coordination_identity.go` and
  `internal/coordination/rate_limits.go` — current key derivation and local registry.

## Code context

- `internal/coordination/rate_limits.go` owns the process-local key and deterministic
  clock seam. It already has context-aware local admission and must remain the default.
- `internal/coordination/dragonfly.go` owns the narrow Redis-compatible client. It is
  the appropriate optional transport boundary; durable state must not be added.
- `internal/connectors/engine/rate_limit_runtime.go` resolves a declared policy to the
  local registry today. It is the sole shared-policy selection point.
- `internal/connectors/connsdk/rate_limits.go` and
  `internal/connectors/engine/schema/rate_limits.schema.json` own declaration shape
  and validation; no connector-specific branch is permitted.
