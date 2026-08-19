# Plan — issue 4276 Zoom certification scope

## Task Delivery Header

- Issue: Refs #4276 — chore(certification): add Zoom to certification scope
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` with the generator-owned Zoom shard
  and status projection committed, verification recorded, and no scope expansion
  beyond Zoom.
- Working branch: fm/cli-zoom-certification-scope-r1
- Task: Add only Zoom to the central certification allowlist, generate only the
  resulting certification artifacts, and prove the certification gate no longer
  halts because Zoom is absent from the capability matrix. Zoom must remain
  honestly uncertified until its parity evidence is proven.
- Verification: Reproduce the pre-change gate; run a red scoped matrix check;
  regenerate with repository generators; inspect the changed-path set; run the
  post-change gate, scoped/global matrix checks, required repository gates, and
  `make verify` before pushing.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Zoom enters only the certification scope | live | The generated `status.json` scope and its connector status contain `zoom`; no other new connector is present. |
| Zoom receives a generated proof shard | live | `certification-matrix --connector zoom --check` validates the committed Zoom shard. |
| The transition gate no longer rejects Zoom as absent | live | The post-change `certification-gate` verdict for `zoom`/`integrate_sub_pr` has no `capability/zoom/missing` failure. |
| Scope addition does not counterfeit certification | live | Generated Zoom status remains `COMMUNITY BUILD, UNCERTIFIED` unless accepted evidence proves otherwise. |

## Locked discussion decisions

- The one connector added to this final-reducer allowlist is `zoom`; no other
  connector is in scope.
- `cmd/connectorgen/certificationallowlist.go` is the only production source
  edit. The two certification-status tests must assert exact generated
  allowlist consistency rather than the prior fixed count.
- Generator output is limited to `internal/connectors/defs/zoom/certification-matrix.json`
  and `internal/connectors/certifications/status.json`. If the generator would
  alter Zoom connector-definition sources owned by the parity lane, stop.
- The certification design requires proof-bearing evidence to make a connector
  certified, not to add it to scheduler scope. Incomplete Zoom cells must yield
  a retry or other non-pass gate verdict, never a fabricated pass.
- The initial checkpoint was rebased onto `origin/main` at `31bfe62eb` before
  final verification because its delete action-kind generator fix changes
  certification classification output. The Zoom shard and full status
  projection are always regenerated from the rebased source, never retained as
  pre-rebase bytes.
- This is an internal generator/status change, not a `pm` CLI contract change;
  runtime help, manual, `docs/cli/**`, website documentation, and completions
  are not applicable.

## TDD execution slices

1. **Red:** run the protected gate and scoped matrix check for Zoom on the
   unmodified scope. Record the absent-capability and non-allowlisted failures.
2. **Green:** add the one allowlist entry, generate the Zoom shard and full
   status projection through `connectorgen`, then prove both matrix checks and
   the post-change gate behavior.
3. **Review/proof:** inspect the narrow diff, confirm Zoom stays uncertified,
   run repository verification including `make verify`, and review the changed
   source and generated JSON for scope drift.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, and `golang-lint`.

Resolved command path: `scripts/gsd doctor`; `scripts/gsd sources` and
`scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. The task supplies the locked scope and an
autonomous direct-PR delivery condition; Pi-compatible isolated workflow agents
are unavailable, so the lifecycle is executed inline and recorded here.
