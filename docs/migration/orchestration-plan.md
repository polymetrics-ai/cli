# Migration Orchestration Plan — connector-architecture-v2

Status: approved (2026-07-02). Execution vehicle: GSD Universal Programming Loop
(phases = waves; see `.planning/ROADMAP.md`). Models: coordinator/planner/reviewer = `fable`;
backend/tester/security/reliability executors = `sonnet` (Sonnet 5).

## Calibration (measured)

- **558 connector packages** under `internal/connectors/` (excluding connsdk, registryset,
  _quarantine), ~309k lines of connector Go. Directory names already unified (no
  source-/destination- prefixes) — the clean break is a catalog/slug/CLI problem.
- **Size buckets**: S <300 loc: 90 · M 300–699: 294 · L 700–899: 155 · XL ≥900: 20 (max github
  3,865). Mid-band uniformity (identical `<name>.go` + `streams.go` + `_test.go` shape on connsdk)
  makes migration highly templatable.
- **catalog_data.json**: 646 entries (590 source, 56 destination), 100% have documentation_url.
  Runtime kinds: declarative_http_go 502, native_go 52, database_go 25, file_go 11,
  destination_go 56.
- **Write coverage today near zero**: 4 packages have write.go (stripe = reference pattern).
- **Parallel authoring designed in**: connectors self-register via init(); `cmd/registrygen`
  regenerates the registry by scanning dirs ("run once at convergence"). Shared files (agents
  NEVER touch; orchestrator-only): `registryset/registry_gen.go`, `catalog_data.json`,
  `icon_data.json`, top-level `internal/connectors/*.go`, `go.mod`.

## Budget truth

**~110 Sonnet executors covers Pass A** (migration at capability parity): bundle sizes S=12/agent,
M=7, L=4, XL=1 → ~105 executor runs + reviewers/repairs. **Pass B** (full capability expansion —
research each API's documented surface, add all streams + write actions) needs **~350 additional
agent runs** (~525 total, ~160–200M tokens, ~8–11 working days).
**DECISION (user): finalize after the pilot** — wave1 produces real per-connector cost data
(`docs/migration/pilot-costs.json`), then choose full Pass B vs capped tiered expansion
(top-100 full surface, tail = parity + mechanical CRUD writes).

## Waves (= GSD phases)

| Wave | Purpose | Executors | Gate |
|---|---|---|---|
| 0 | Engine, connectorgen, conformance v2, certify core, conventions, goldens (stripe, searxng, postgres) | 10–12 | engine tests green; goldens pass conformance + parity; connectorgen validate rejects seeded bad input |
| 1 | Pilot: xkcd, vitally, bitly, calendly, sentry, chargebee, zendesk-support, monday, github, gmail (1 agent each) | 10 + 3 reviewers | all 4 verification layers + Fable line-by-line review; conventions patched |
| 2 | Fan-out declarative-HTTP S+M | ~49 | wave gate |
| 3 | Fan-out declarative-HTTP L+XL | ~56 | wave gate; 100% review of XL |
| 4 | Non-HTTP kinds (database/file/destination/native) | ~15 | wave gate |
| 5 | Pass B capability expansion (roster per pilot decision) | ~300 + 40 critics | completeness ≥95% or documented skips |
| 6 | Convergence: catalog from manifests, slug clean break, docs, full certification | 6–8 | full build+test+lint+100% conformance (minus quarantine); HUMAN GATE before deletion |

## Per-agent task spec (Pass A)

Self-contained prompt (template in `docs/prompts/universal-programming-loop-prompts.md`): assigned
connector dirs (exclusive), runtime_kind, loc, doc URLs, pre-extracted manifest dump; references
`docs/migration/conventions.md` + goldens; FORBIDDEN: shared/generated files, unassigned dirs,
go.mod — new dep or engine feature = typed BLOCKER, never a workaround. Self-verify: connectorgen
validate + package build/vet/test + TestConformance/<name>.

Structured result (docs/migration/result.schema.json): per connector `{name, status:
migrated|partial|blocked, files_changed, streams_before, streams_after, write_actions_added,
escape_hatches[{file,reason}], fixtures_added, conformance{passed,failing_tests},
blockers[{type: AUTH_COMPLEX|NON_REST|DOCS_UNREACHABLE|SCHEMA_AMBIGUOUS|NEEDS_NEW_DEP|ENGINE_GAP,
reason, evidence}], notes}`. Pass B adds `api_surface_endpoints_total`, `endpoints_implemented`,
`endpoints_skipped[{path, reason}]`.

## Parallel-mutation safety

All agents in a wave share one worktree — safe because assigned dirs are disjoint and no shared
file is touched. Enforced, not trusted: after each wave a `git status --porcelain` path-guard
asserts every changed path is under assigned dirs (+ docs/migration/); violations reverted, bundle
failed (one retry). Wave-close (single-writer, orchestrator only):

```
go run ./cmd/connectorgen gen     # (wave0+; registrygen until the flip)
go build ./... && go test ./internal/connectors/...
golangci-lint run ./internal/connectors/...
git add -A && git commit -m "wave N: <k> connectors migrated"
```

## Verification pyramid

1. **Agent self-check**: connectorgen validate, package build/vet/test, per-connector conformance.
2. **Wave gate** (deterministic): path guard, regen, full build, full connector tests, full
   conformance, lint. Failures map to owning bundle → repair agent with error log (1 retry, then
   quarantine).
3. **Adversarial review**: 20% sample + 100% of XL; checklist: schema fidelity vs live API docs
   (fetch documentation_url; spot-check 3 streams + every write action's method/path/required
   fields), fixture realism, escape-hatch justification, secret redaction, conventions adherence.
   >30% sample failure → halt wave; Fable inspects for systemic template defect.
4. **Completeness critic** (Pass B): api_surface.json vs implemented → coverage % →
   `docs/migration/coverage-report.json`.

## Failure handling

Blocked connectors keep their current working implementation; recorded in
`docs/migration/quarantine.json` `{name, blocker_type, reason, evidence, attempts}` and flagged in
the catalog. ≥3 connectors hitting the same ENGINE_GAP → extend engine (mini wave-0 increment) →
targeted un-quarantine wave. Expected quarantine 3–5% (SOAP/GraphQL/AWS-signed APIs). Resume:
`docs/migration/status.json` (orchestrator-written after each bundle) — planner recomputes
`remaining = inventory − completed − quarantined`; each bundle completion = one git commit.

## Artifacts

- `docs/migration/inventory.json` — generated: name, path, loc, runtime_kind, bucket, catalog
  slugs, doc URLs, current stream count (wave 0 deliverable).
- `docs/migration/conventions.md` — the single migration recipe (wave 0, patched after pilot).
- `docs/migration/result.schema.json`, `docs/migration/review.schema.json` — agent I/O contracts.
- `docs/migration/status.json` — completed/quarantined/needs_repair per wave.
- `docs/migration/pilot-costs.json` — wave 1 deliverable feeding the Pass B decision.
- `docs/migration/quarantine.json`, `docs/migration/coverage-report.json`.

## Budget & timeline estimates

Migration bundle agent 150–300k tokens (avg ~220k); Pass B research/expansion 250–450k (avg
~350k); reviewer ~120k; repair ~150k. Pass A ≈ 30M tokens; Pass B ≈ 125M; waves 0/6 + overhead ≈
10M. Wall-clock at concurrency 12–16: wave 0: 2–3d · pilot: 0.5–1d · waves 2–4: 1.5–2d · Pass B:
3–4d · convergence: 1d → **~8–11 working days**.

## Pass B decision (USER, 2026-07-02, post-pilot)

**Measure first.** Pass A fan-out proceeds once the S3 engine mini-wave lands. At wave5 start:
expand 5 representative connectors (spread across S/M/L, read-only vs write-capable APIs) to get
real Pass B per-connector costs, then auto-scale the roster to the data — full ~350-agent
expansion if affordable, tiered top-100 otherwise. Matches docs/migration/pilot-costs.json's
recommendation.
