# Plan: Front Operation Ledger (#192)

Issue: #192  
Parent: #188  
Parent branch: `feat/188-front-cli-parity`  
Branch: `feat/192-front-operation-ledger`  
Primary write scope: `internal/connectors/defs/front/api_surface.json`, `.planning/phases/issue-192-front-operation-ledger/`

## GSD command path

- Planning prompt used: `scripts/gsd prompt plan-phase 192 --skip-research --tdd`.
- Programming-loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-192-front-operation-ledger --dry-run`.
- Result: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual GSD fallback: follow the repo's universal loop inline: plan, red validation, production edit, green validation, commit/push, automated review routing.

## Required skills loaded

Recorded from the parent/#189 session and still applicable to this Go/CLI connector metadata task:

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-spf13-cobra`
- `golang-spf13-viper`
- `golang-lint`

## Scope

Replace Front's 10-row legacy `api_surface.json` with an explicit operation-ledger-mode inventory for the official Front ReadMe OpenAPI registry:

- `core-api.json` registry UUID `103mteemr3o8hk5`.
- `channel-api.json` registry UUID `48cgkyj15mnqc4ttp`.
- 255 REST operations captured from the public ReadMe API registry without credentials.
- Preserve existing executable stream coverage for the 6 implemented streams.
- Mark every other REST operation as a blocked operation row with a closed-vocabulary `operation.model`, risk, reason, and source URL/notes.
- Do not add executable streams, direct reads, binary downloads, writes, dependencies, credentials, or runtime behavior in this slice.

## Official-source reconciliation

The parent issue records a 342-operation baseline. The current public ReadMe OpenAPI registries expose 255 REST operations:

- Core REST registry: 244 operations (`GET=123`, `POST=70`, `PATCH=23`, `PUT=1`, `DELETE=27`).
- Channel registry: 11 operations (`POST=6`, `PATCH=3`, `PUT=2`).
- Combined registry: 255 operations (`GET=123`, `POST=76`, `PATCH=26`, `PUT=3`, `DELETE=27`).

`https://dev.frontapp.com/llms.txt` exposes 346 API Reference links. The 91 links not represented in the OpenAPI operation set are category/guide pages, plugin SDK/browser-context methods, Channel API overview pages, and data-model pages. This phase will:

1. make `api_surface.json` complete for the official REST OpenAPI registries;
2. record the 91 non-REST/API-reference links in source notes as not part of the REST connector execution surface; and
3. avoid inventing missing 342-row method/path data.

## Classification policy

- Existing implemented streams keep `covered_by.stream`: `/contacts`, `/conversations`, `/inboxes`, `/tags`, `/teammates`, `/channels`.
- GET endpoints returning structured records but not yet executable use `operation.model: direct_read`, `status: blocked`, `risk: low` or `medium`, and a reason naming the follow-up lane (#191 stream runner or #193 direct read).
- Attachment/download endpoints use `operation.model: binary_read`, `risk: medium`, and a reason naming #194 bounded binary output policy.
- POST/PATCH/PUT mutations use `operation.model: sensitive_reverse_etl` unless clearly admin/configuration-impacting.
- DELETE and irreversible/removal endpoints use `operation.model: destructive_action`.
- Admin/configuration surfaces (rules, inbox/channel/team/teammate group/admin-style configuration, knowledge base mutation, templates, signatures, shifts, views, custom app/channel sync/calls where applicable) use `admin_reverse_etl` or `destructive_action` depending on method/risk.
- `/me` stays blocked as `disallowed`/low-risk token identity introspection, not a syncable stream.
- Duplicate method/path pairs are deduplicated with one canonical row; if a duplicate appears in source capture notes, record it there rather than inventing an executable surface.

## TDD loop

1. Plan checkpoint commit before production edits.
2. Red validation: assert `api_surface.json` is not in operation-ledger mode and does not contain the 255 captured official REST operations.
3. Generate/rewrite `internal/connectors/defs/front/api_surface.json` from the captured public registries and deterministic classifier.
4. Green validation:
   - `jq empty internal/connectors/defs/front/api_surface.json`
   - count/method split check against the captured 255-operation registry
   - `go run ./cmd/connectorgen validate internal/connectors/defs`
   - focused engine/connectorgen tests if affected
5. Commit/push implementation checkpoint.

## Safety gates

- No secrets or credentialed Front checks.
- No reverse ETL execution.
- No new dependencies.
- No generic shell, generic HTTP write, generic SQL write, raw API, or raw mutation escape hatch.
- This slice is ledger-only; blocked operation rows must not become executable CLI actions.
- Destructive/admin operations are blocked by default and routed to #195 for typed policy work.

## Planned verification

See `VERIFICATION.md` for exact commands and results ledger.
