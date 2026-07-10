# Plan: issue-212 Help Scout all-operations implementation

Date: 2026-07-10
Branch: `feat/212-helpscout-all-ops`
Parent issue: #212
Depends on: #213 metadata refresh branch/PR #236

## GSD path

- Adapter check: `scripts/gsd doctor` passed.
- Planning prompt: `scripts/gsd prompt plan-phase issue-212-helpscout-all-ops --skip-research --tdd`.
- Required programming loop: attempted `scripts/gsd prompt programming-loop init --phase issue-212-helpscout-all-ops --dry-run`; adapter reports `unknown GSD command: programming-loop`. Manual GSD/TDD fallback is active and recorded here.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- CLI parity reference: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- Connector architecture references: migration handoff, conventions, and v2 design.

## Objective

Implement Help Scout operation parity using the GitHub connector pattern: every official Help Scout Inbox API endpoint is represented as one of:

- existing ETL stream;
- implemented bounded JSON direct read;
- typed reverse-ETL write action that can be planned, previewed, approved, and executed through existing reverse gates;
- bounded binary/file operation metadata; or
- explicit duplicate/deprecated/disallowed/product-scope block where applicable.

Do not add raw generic HTTP writes, generic shell, generic SQL, unbounded binary downloads, new dependencies, credentialed checks, or secret values.

## Slices

1. **Red tests / validation probes**
   - Add failing tests for generic JSON direct-read output policy support.
   - Add Help Scout operation coverage tests: no endpoint remains with a blocked operation except binary/duplicate/disallowed categories, and every non-binary mutation has a write action.

2. **Direct-read engine policy**
   - Add an explicit `json` direct-read output policy to `cli_surface.schema.json`, commandrunner validation, and engine direct-read policy handling.
   - Preserve GitHub contents policies and sensitive repository path checks.

3. **Help Scout operation ledger and writes**
   - Generate `operations.json` for all 145 Help Scout endpoints.
   - Generate typed `writes.json` for non-binary mutating endpoints using explicit path fields and request-body fields derived from official docs request examples where available.
   - Mark destructive/admin/sensitive writes with approval/risk metadata and destructive confirmation where appropriate.
   - Keep binary/file endpoints as bounded metadata operations, not raw downloads.

4. **Help Scout direct-read CLI surface**
   - Convert JSON GET endpoints from planned blocked rows to implemented direct-read commands with explicit path/query flags and `output_policy: json`.
   - Keep existing stream commands intact.
   - Add operation-backed binary/file metadata commands for bounded binary endpoints; these remain feature-gated by the operation executor until a dedicated binary executor exists.

5. **Docs/help parity**
   - Update `metadata.json`, `docs.md`, generated connector manual/skill/catalog, and website connector data.
   - Verify runtime help/inspection surfaces without credentials.

## Human gates / exclusions

- No credentialed Help Scout calls.
- No dependency additions.
- No secret values in files, prompts, logs, or PR body.
- No raw generic write operation or arbitrary method/path command.
- Reverse ETL execution remains gated by plan → preview → approval token → typed confirmation.
- Binary downloads are bounded metadata operations only unless the existing operation runner supports safe file materialization.

## Checkpoints

1. Plan artifact commit.
2. Red test commit if useful.
3. Engine policy and Help Scout operation metadata commit after targeted validation.
4. Docs/generated parity commit.
5. Final verification commit and PR update.
