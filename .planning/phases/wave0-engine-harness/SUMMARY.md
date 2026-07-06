# Phase Summary — wave0-engine-harness

Status: **completed** (2026-07-02). Final HEAD: 7fb4eb6 (+ phase-close commit). Reviewer verdict:
**GO for wave1-pilot** (re-review after gap-loop cycle 1; REVIEW.md).

## What was built

| Deliverable | Location | Proof |
|---|---|---|
| Declarative engine (loader, draft-07 subset validator, interpolator w/ chained filters + join + literals + CRLF/`..` guards, auth selection w/ ctx-threaded hooks, 6 paginators w/ SSRF guards, read path w/ MaxPages + typed header semantics, write path, connector assembly, derived sync modes) | internal/connectors/engine/ | 200+ tests, 85.7% coverage, race-clean |
| Definition/DefinitionProvider (verbatim RawSpec) | internal/connectors/definition.go | added alongside legacy, interface untouched |
| defs bundles + go:embed | internal/connectors/defs/ | 3 goldens validate clean |
| Golden: stripe (declarative HTTP + form writes) | defs/stripe/ + parity_stripe_test.go | engine-vs-legacy parity incl. app-persisted cursor round-trip; real wire-shape fixtures |
| Golden: searxng (read-only, optional bearer-proxy auth) | defs/searxng/ + parity_searxng_test.go | RAW record equality (no normalizations) |
| Golden: postgres (Tier-3 component split + bundle) | native/postgres/ + defs/postgres/ | fixture-mode parity + error-classification table; no registration (grep-guard) |
| Hook seam (5 interfaces) + hookset/nativeset gen | engine/hooks.go, hooks/, native/nativeset/ | dispatch tests |
| cmd/connectorgen (validate\|gen\|new) | cmd/connectorgen/ | 16 seeded-invalid bundles, 14+ rule classes; byte-stable gen |
| Conformance v2 (10 static + 8 dynamic real-engine fixture-replay checks, numeric+string cursors) | internal/connectors/conformance/ | TestConformance/{stripe,searxng,postgres} green |
| Certify core (report/history, in-process cli.Run harness, source stages 0–11, all-output secret scanning incl. JSON-escaped) | internal/connectors/certify/ | proven vs sample; sabotage self-test |
| cmd/inventorygen + inventory.json | cmd/inventorygen/, docs/migration/inventory.json | 557 connectors: S137/M388/L31/XL1 |
| Lint gates | .golangci.yml, Makefile (lint, connectorgen-validate in verify) | make verify green end-to-end |
| Migration recipe + agent I/O contracts | docs/migration/{conventions.md,result.schema.json,review.schema.json} | citations verified; deviation ledger w/ meta-rule |

## Process record

- 18 TDD pairs + 1 docs task + repairs, all red-first with ledger evidence (TDD-LEDGER.md → traces/).
- Coexistence held: legacy tree zero-diff vs main except the sanctioned registrygen skip-map (+15
  lines); registryset regen byte-identical (557 imports).
- Verification: make verify (incl. smoke) green ×2; gsd-verifier 6/6 criteria + 7/7 EVAL metrics;
  security pass-with-findings (0 block; majors M1/M2 fixed in gap loop); Fable review NO-GO →
  gap-loop cycle 1 (3 blocks, 10 flags → all blocks + 9 flags fixed) → re-review **GO**.
- ENGINE_GAPs found by goldens and fixed same-phase: MaxPages unwired; when-absent-key semantics;
  app-layer cursor round-trip (B1 — the parity test initially masked it; now honest).

## Carried into wave1-pilot (documented follow-ups, REVIEW.md re-review section)

N1 formatCursorForAssertion github_date_range alignment (fix before github bundle lands) ·
N2 all-digits non-unix start-value misread (consider validate-time guard) · N3 relative next-URLs
fail closed (pilot note) · N4 stale doc comment incrementalLowerBoundValue · N5 `..%5C`/`x%2F..%2Fy`
residuals (decode-before-route servers) · searxng subreddit narrowing (wave6 human-gate item) ·
gmail 3-legged OAuth needs an AuthHook pattern decision in the pilot · conformance pkg coverage
83.1% (no gate) · 11MB blob remains in git history (optional rewrite before any push).

## Numbers for the pilot cost model

Wave0 actual: 14 executor runs + 1 repair + 2 gap-loop runs + 3 verification agents + planner +
plan-checker (Sonnet executors ~100k–460k tokens each; Fable planner/reviewer ~130k–200k).
Real bucket histogram (inventory.json): S137/M388/L31/XL1 → projected Pass A fan-out ~77 bundle
agents (down from ~105 estimate).
