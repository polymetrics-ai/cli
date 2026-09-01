# TDD Ledger: Batch R1 vNext source-lock cutover

## Planned evidence

| Slice | Red characterization | Green contract | Refactor/verification |
| --- | --- | --- | --- |
| Runtime dependency | Reference connectors require embedded authoring/certification/admission material. | GitHub, GitLab, and Asana load and reach credential/approval preflight from execution JSON alone. | Audit embedded/runtime reads and run focused plus fleet tests. |
| Connector-local invalidity | Global ledgers or one bundle error can suppress the fleet. | Malformed required execution JSON rejects that connector without hiding healthy connectors. | Assert stable typed diagnostics and deterministic discovery. |
| Canonical rendering | Existing source-lock paths do not own a single canonical all-lane projection. | One vNext model renders byte-stable existing execution JSON through shared schema refs. | Check every rendered file and reject stale output. |
| Lane semantics | Retention/certification state can hide documented commands and source operations cannot express all lanes canonically. | Direct, binary, ETL, reverse ETL, sync, and explicit-empty lanes are surfaced without provider switches. | Run the same all-lane contract for every Batch R1 connector. |

## Actual evidence

### 2026-09-01 — inherited ledger reconciliation

- The inherited `Red: pending` / `Green: pending` record contradicts the branch handoff's claimed reference-cohort green checkpoint. It is retained as history, not accepted as evidence for this continuation.
- Baseline: clean isolated continuation branch `fm/cli-batch1-vnext-cutover-r2` at `0b214b79eeb871238ce8454cd7b896e71e2746a7`, with that SHA proven reachable from `origin/fm/cli-top100-declaration-batch-r1`.
- Manual GSD fallback: the adapter resolves every required lifecycle command and the canonical contract check passes, but its generated commands cannot execute this pre-existing named phase because `.planning/ROADMAP.md` is absent. The inline artifacts in this directory carry the required lifecycle evidence.

### 2026-09-01 — cleanup slice A: native fixture bypass

- Red target: a native connector supplied only `config.mode=fixture` must not report a successful check or emit canned records; it must continue through normal credential/config validation before provider I/O.
- Red command: `go test -timeout 20m ./internal/connectors/native/alpha-vantage -run '^TestFixtureModeNoLongerBypassesCredentialBoundary$'`.
- Green contract: the same invocation returns the connector's missing-credential error with no provider request, and the production implementation contains no fixture-mode branch.
- Follow-on residual proof: scan production native/hook/engine/connectorgen code, definitions, generated docs/skills, and website sources for a fixture, importer, certification, retention, compatibility, feature-flag, or second-executor execution/admission path. Any retained mention must be connector-local provider provenance only and is recorded by path and reason in `VERIFICATION.md`.

### 2026-09-01 — cohort migration template

- Red: before each named connector migration, `lock-render <connector> --check` fails against the newly written lock or the connector lacks a usable declaration/credential-boundary witness.
- Green: `lock-render <connector>` produces byte-identical execution JSON; all seven lanes are explicit; malformed execution JSON rejects locally; and an isolated, credential-free command reaches the ordinary missing-credential or approval boundary without provider I/O.
- Connector sequence: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel. Each receives its own red/green entry, review record, commit SHA, and normal push evidence.
