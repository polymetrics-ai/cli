# Zoom five-class parity lane R1

## Task Delivery Header

- Issue: Refs #4265 — feat(zoom): deliver five-class parity foundation cohort
- Base branch: `main`
- Merges into: `fm/cli-zoom-parity-lane-r1` → `main`
- Delivery: Draft parent PR open against `main`; each wave lands through a reviewed sub-PR into the parent branch. The parent remains draft until all five waves have passed their gates. Only the captain may merge it to `main`.
- Working branch: `fm/cli-zoom-parity-lane-r1`
- Task: Deliver the bounded Zoom five-class foundation cohort: retain existing ETL; salvage 70 direct reads, compatible REST writes including canonical deletes, two reverse-ETL actions, and one bounded Clip download. Keep every definition under `internal/connectors/defs/zoom/`; record non-certifiable work explicitly rather than claiming live certification.
- Verification: Per wave: focused definition/conformance/commandrunner tests, `connectorgen validate`, `surface-sync --check`, connector certification sweep/gate, generated docs/catalog checks, `make connector-runtime-preflight`, `make connector-boundary`, and `make verify` before every push. Parent: full `make verify`, integrated review coverage, and certification-gate `PROCEED` at each enforced transition.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Existing ETL remains executable | fake | Replay fixtures exercise the real commandrunner without credentials; live proof is blocked until approved server-to-server OAuth consumption exists. |
| Direct-read cohort is executable | fake | Deterministic provider replay asserts returned record counts, parameter mapping, generated mappings, and runtime preflight; no approved live credential path exists. |
| Direct-write/delete cohort is safe | fake | Fixture tests prove typed plan/preview/approval and destructive output contracts; live mutation requires captain-approved disposable scope and the create-then-cleanup ledger. |
| Reverse-ETL actions remain gated | fake | Fixture tests prove typed action and reverse safety lifecycle; no run-owned cleanup/readback pairing or captain execution approval exists. |
| Clip download is bounded binary candidate | fake | Fixture test proves the bounded download contract; permitted live Clip resource and supported auth are unverified. |
| Certification status is honest | live | Certification gate and generated matrix reject absent/invalid accepted live-evidence records, so this lane must remain explicitly uncertified until proof exists. |

## Source and ownership policy

- Authoritative scope: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-zoom-foundation-delta-r1/report.md`.
- Exactly one connector target: `zoom`.
- Connector-owned production paths: `internal/connectors/defs/zoom/**` plus target generated docs/catalog artifacts only when repository generators require them.
- Foundation changes are forbidden in this lane. The operation-backed delete `write_action_kind` fix and `file_upload` executor (G12) are external prerequisites, not Zoom workarounds.
- Preserve the existing Zoom provider ledger and three stream-backed ETL commands exactly.

## Issue map and delivery state

| Wave | Issue | GSD phase | Branch | Base | Initial dependency |
| --- | --- | --- | --- | --- | --- |
| ETL | #4266 | `zoom-etl-certification-parity-r1` | `test/4266-zoom-etl-certification-parity` | parent | certification foundation on `main` |
| Direct read | #4267 | `zoom-direct-read-parity-r1` | `feat/4267-zoom-direct-read-parity` | parent | ETL wave integrated |
| Direct write/delete | #4268 | `zoom-direct-write-parity-r1` | `feat/4268-zoom-direct-write-parity` | parent | direct read + delete action-kind foundation fix |
| Reverse ETL | #4269 | `zoom-reverse-etl-parity-r1` | `feat/4269-zoom-reverse-etl-parity` | parent | direct write wave |
| Binary download | #4270 | `zoom-binary-download-parity-r1` | `feat/4270-zoom-binary-download-parity` | parent | direct read wave |

Parent canonical state: `parent_draft_pr` pending seed verification and push.

## Setup evidence

- Worktree isolation: `pwd -P` and `git rev-parse --show-toplevel` both returned `/Users/karthiksivadas/.treehouse/cli-83d592/1/cli` before the branch was created.
- Parent issue: [#4265](https://github.com/polymetrics-ai/cli/issues/4265); child issues: #4266, #4267, #4268, #4269, #4270. The GitHub sub-issue relation is established.
- GSD adapter: `scripts/gsd doctor` passed; `scripts/gsd list` included all five required lifecycle commands.
- GSD provenance: `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` all resolved to the locked local adapter sources.
- Contract: `go run ./cmd/agentcontractgen check` passed (`canonical contract and registered projections are current`).

## Required skills

- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`

The manual GSD execution fallback is intentional: the canonical contract forbids spawning compatible isolated GSD roles in this lane. Generated prompts will be executed inline and recorded below.
