# Google Calendar parity wave 03 plan

## Scope

Branch: `fm/cli-google-calendar-parity-wave03-r1`

Primary issue: #3054. Capability subissues: #3055-#3061.

This is a connector-local parity implementation for `google-calendar` using fixture-only validation. No live Google Calendar API calls, credential requests, provider writes, pushes, PRs, `/no-mistakes`, VPS, Thaalam, or Herdr lifecycle commands are in scope.

## GSD command path and fallback

- Ran `scripts/gsd doctor` successfully.
- Tried the required documented command `scripts/gsd prompt programming-loop init --phase google-calendar-parity-wave03 --dry-run`; the repo-local GSD adapter returned `unknown GSD command: programming-loop`.
- Manual GSD programming-loop fallback is active for this branch, using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and this phase artifact set (`PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`).
- Generated fallback prompt evidence: `GSD-PROMPT.txt` (from `scripts/gsd prompt execute-phase google-calendar-parity-wave03 --dry-run`).

## Skills loaded

- Repository routing: `.agents/agentic-delivery/references/required-skills-routing.md`.
- GSD: `gsd-core`, `.agents/agentic-delivery/references/gsd-pi-adapter.md`, issue-agent/parent workflow references.
- Connector architecture: `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`.
- CLI docs parity: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- Go skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`.

## Official-source evidence

- Read parent/subissue bodies with `gh-axi issue view <issue> --full`; saved exact outputs under `sources/issue-3054.md` through `sources/issue-3061.md`.
- Re-audited public official Google Calendar sources without credentials:
  - Discovery: `https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest`.
  - Reference root and resource pages under `https://developers.google.com/workspace/calendar/api/v3/reference`.
- Parsed the discovery document into `google-calendar-official-operations.json`: 38 operations total (`GET=11`, `POST=15`, `PATCH=4`, `PUT=4`, `DELETE=4`).

## Implementation slices

1. **Ledger and schemas (red static gate):** replace the quarantine hook-only surface with an operation-level `api_surface.json`, operations/direct-read metadata, complete stream list, write actions, and schemas. Run `connectorgen validate` and expect initial failures until fixtures are present.
2. **Fixture-backed conformance:** add sanitized `fixtures/check.json`, stream pages for every executable read stream, and write fixtures for every executable reverse action. Ensure the custom OAuth refresh hook has a fixture/conformance no-op path so replay never calls Google token endpoints.
3. **CLI/config/help metadata:** add `cli_surface.json` with implemented ETL/direct-read/reverse commands and safety notes. Ensure generated manuals/skills/catalog/website data reflect the connector definition.
4. **Runtime registration alignment:** ensure `pm connectors inspect google-calendar --json` reflects bundle capabilities and command surfaces. If the promoted native registration masks engine bundle capabilities, switch only the google-calendar promoted factory to the engine bundle and document why.
5. **Docs and issue addendum:** update connector docs, generated surfaces, and append the captain-policy addendum to #3054-#3061 with actual post-change counts and no certification claim.
6. **Verification and commit:** run required local gates, update `SUMMARY.md`/`RUN-STATE.json`, commit a clean green slice, and stop without push/PR/no-mistakes.

## Safety decisions

- Auth remains OAuth2 refresh-token based. Secrets stay in `spec.json` with `x-secret: true`; fixtures use synthetic placeholder values only.
- Reverse ETL writes are named actions only; there is no generic Google Calendar HTTP method/path/body command.
- Destructive or cancellation actions (`delete_*`, `clear_calendar`, `transfer_calendar_ownership`, `stop_channel`) require typed closed schemas, destructive confirmation, risk text, and reverse ETL plan -> preview -> approval -> execute.
- `cdc` metadata remains false unless a real `CDCReader`/webhook state runtime exists. Google Calendar watch/channel operations can be represented as explicit reverse-ETL management actions; webhook delivery/changefeed consumption remains a documented shared-runtime dependency.
- Certification remains `0`; fixture-only conformance is not live provider certification.

## Orchestration decision

Cycle `plan`: `local_critical_path`. This worker owns one connector and no isolated sub-worker is needed; firstmate owns parent orchestration/integration.
