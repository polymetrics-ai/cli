# GSD lifecycle trace — cli-top50-sweep-resume3-r1 (workday-rest)

Branch `fm/cli-top50-sweep-resume2-r1`. Program `cli-top50-fixed-schema-sweep-r1`, landing order #2
under the captain's largest-first reversal, behind github.

Continues `gsd-top50-sweep-continue-r1.md`. The delivery contract in `AGENTS.md` outranks the
dispatch brief where they differ (captain, 2026-08-07); the brief said "re-derive nothing", and the
standing rule "trust the live artifact, never the ledger" plus "verify your own counts" applies once
a derivation defect is suspected. That conflict is what produced finding 34.

## Command provenance

Resolved with `scripts/gsd sources <command>` (a **Node** script — `bash scripts/gsd` dies with a
syntax error):

```
/Users/karthiksivadas/.treehouse/cli-83d592/10/cli/.gsd/commands.json
/Users/karthiksivadas/.treehouse/cli-83d592/10/cli/.gsd/upstream.lock.json
/Users/karthiksivadas/.treehouse/cli-83d592/10/cli/.gsd/official-docs/COMMANDS.md
```

Gate: `GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow` → **exit 0**.

## Lifecycle

| Phase | Where it landed |
| --- | --- |
| `discuss-phase` | inherited: `PLAN.md` hazards 1–7 from slice 1, re-opened by hazards 8–9 this slice |
| `plan-phase --tdd` | `PLAN.md` (slices, hazards 8–9), `RUN-STATE.json`, `DERIVED-OPERATIONS.json` regenerated at 907 |
| `execute-phase` | `TDD-LEDGER.md` cycle 2 (red at 907) → reads → mutations |
| `verify-work` | `VERIFICATION.md` — every gate run, real output quoted |
| `code-review` | pending; runs under `/no-mistakes` per the firstmate contract |

## Inline execution, recorded as required

Executed **inline** rather than by spawning isolated agents. The canonical parent-worker contract
forbids spawning an orchestrator, shepherd, planner, reviewer, verifier, GSD role, or extra worker
for this job; the single canonical worker processes ready waves inline. Recorded here per
`AGENTS.md`.

## Required skills

`golang-how-to` routing → `golang-testing` (the red-first surface test and its tightening),
`golang-cli` (command surface, group help, flag shapes), `golang-security` (no credential became a
flag; `tenant` and `access_token` stay connection-spec fields).

## What this slice changed about the sweep's shared knowledge

- **Finding 34** — deduping on the resolved path is not enough; dedup on the **templated** path too.
  workday-rest is 907, not the recorded 920 or slice 1's 916.
- **Finding 35** — read binary from `produces`, never from a `?type=` name. I recorded six binary
  variants on the name and corrected to two on the evidence **before authoring**.
- **Finding 36** — `go test ./internal/cli/` inherits Go's 10m default; the project gate is
  `-timeout 20m`. The bare form dies mid-run and looks like a failure you caused.
- **Finding 37** — a collapsed query-string behaviour needs `omit_when_absent`; a fixed `query`
  value validates clean, passes every gate, and silently changes the endpoint's default behaviour.
- **Finding 38** — verify a "0 partial" against the specs; it is indistinguishable from a broken
  required-field recursion.
- **Known-red correction** — `TestGoldenTranscripts` fails **11** subtests, not the one the handoff
  recorded. Proven pre-existing by stashing this slice and comparing the failing sets.

## Repo safety overlay — confirmed

- No secret requested, printed, stored, or summarized. `tenant` and `access_token` remain connection
  spec fields; neither became a command flag.
- No dependency added.
- No credentialed connector check run. Every one of the 911 binary invocations was `--help`, which
  resolves no credentials and calls no provider. The only network access was re-fetching Workday's
  **public** service directory and its 52 public specs.
- Reverse ETL remains plan → preview → approval → execute; all 252 write actions carry the approval
  string and `risk`.
- No generic shell, generic HTTP write, or generic SQL write tool exposed.

## CLI help/docs/website parity overlay — partially closed, remainder stated

- **Runtime help: done.** All 911 commands render and were invoked to prove it (`NAME` line asserted,
  not exit status).
- **`pm workday-rest` bare namespace**: renders the group summary and exits 0, per the documented
  namespace behaviour.
- **`docs/connectors/workday-rest/`: regenerated** (MANUAL.md, SKILL.md), and the 1,029 other doc
  files a bare `pm docs generate` rewrote were reverted — pre-existing `main` drift, finding F6.
  `pm docs validate` passes.
- **Website catalogs and golden transcripts: NOT regenerated**, by design. Shared generated artifacts
  regenerate **once at the end of the sweep**; per connector they would churn ~1,031 files of
  pre-existing drift on every commit. Both are carried as known-unmet items in `VERIFICATION.md` and
  must be discharged before the PR merges.
