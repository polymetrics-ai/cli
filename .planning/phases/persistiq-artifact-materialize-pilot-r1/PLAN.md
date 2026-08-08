# PersistIQ artifact materialization pilot - plan

> **Scope locked by captain:** one connector only. This phase is a documented
> inline/manual GSD fallback because the pilot is not in `ROADMAP.md` and the
> available Codex runtime cannot provide the interactive Pi phase workers.
> The fallback does not waive red/green evidence or any static gate.

## Goal

Fetch the ledger-linked PersistIQ OpenAPI artifact, reconcile all 21 provider
operations to the repository model, stage a materialized bundle with existing
`connectorgen batch` tooling, run static/runtime-preflight gates, invoke every
generated command through the real no-credential `pm` binary, and report exact
wall-clock timing and counts. Stop after PersistIQ.

## Timed pilot slices

### Slice 1 — identify the link (completed before this plan)

- Read the external ledger record for `persistiq` and emit its
  `artifact_url`.
- Evidence: `https://persistiq.com/api-docs/v1/swagger.json`; shell `real`
  time `0.02s`.

### Slice 2 — map the operation model (completed before this plan)

- Map the ledger's 21 operations by method/path using the existing PersistIQ
  surface inventory, without fetching a new artifact.
- Evidence: 11 `etl`, 1 `direct_read`, 7 `reverse_etl`, 2 `direct_write`, 0
  `binary_download`, 0 unclassified; shell `real` time `0.04s`.
- Reconcile every row after fetch; a count or method/path mismatch is a
  reportable failure, not a reason to invent coverage.

### Slice 3 — fetch and parse the artifact

- Fetch exactly the ledger URL once into the phase-local artifact cache with a
  bounded timeout and redirect handling.
- Record exact byte count and SHA-256 in the phase report/summary.
- Parse as YAML/JSON and require OpenAPI 3.x or Swagger 2.x. The existing
  batch parser remains authoritative; a failed parse is a failed pilot stage.
- No credentials or provider requests beyond the public artifact fetch.

### Slice 4 — materialize and gate

- Run `connectorgen batch plan` with `--connector persistiq --size 1` against
  the ledger.
- Copy the current PersistIQ source bundle to a phase-local source root and
  materialize into a separate phase-local destination, using the existing
  `batch materialize --artifact-dir` path.
- Run `connectorgen validate` (0 findings), `surface-sync --check` (no drift),
  `batch gate`/runtime preflight, the repository's implemented-command test,
  and all applicable package/build gates.
- If materialization fails, preserve the exact report and do not claim later
  gates passed. Do not hand-author a replacement generator or weaken a gate.

### Slice 5 — real binary reachability and report

- Build the real `pm` binary from the staged/installed PersistIQ bundle.
- Run `pm persistiq` bare namespace and every generated command with safe help
  arguments; count reachable versus failed commands and record command names.
- State explicitly: implemented only if static gates pass; not certified and
  never provider-exercised in either case.
- Report wall-clock times for all five slices and the total.

## TDD evidence plan

1. **RED:** Before production bundle edits, run a real PersistIQ command against
   the baseline bundle and capture the observed unknown-command/missing-surface
   failure. The red assertion is that the pilot's generated command set is not
   fully reachable before materialization.
2. **GREEN:** After the existing batch materializer produces the staged bundle,
   repeat the same command-reachability sweep and static gates. Record the
   actual output; no expected pass is declared in advance.
3. **REFACTOR:** Keep changes generated/staged to PersistIQ plus this phase's
   evidence. Run formatting/validation checks and inspect the diff for no
   unrelated connector paths.

## Required GSD command evidence

Resolved successfully with `scripts/gsd sources`:

```text
discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review
```

`go run ./cmd/agentcontractgen check` passed before planning. The actual phase
was executed inline because the phase is unscheduled and the runtime worker
contract is unavailable; this manual fallback is recorded here, in RUN-STATE,
and in the final report.

## Required skills loaded

- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
  `gsd-verify-work`, `gsd-code-review`

## CLI parity checklist

- [ ] `pm persistiq` bare namespace: contextual help and exit 0.
- [ ] Every generated `pm persistiq <command> --help` reaches the command
  parser without credentials/network.
- [ ] `pm help <topic>` and docs/website generation are not applicable unless
  the materializer adds a new command surface; record the result explicitly.
- [ ] `surface-sync --check` is clean and generated flag metadata is derived,
  not hand-authored.
