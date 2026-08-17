# Context — reverse-ETL API-surface derivation r1

## Task Delivery Header

- Issue: Refs #4015 — production MVP certification surface accountability.
- Base branch: `integration/4015-mvp-flat-r1` at `db494bc8fba63023b9ca79b022c2b9dd638aaf76`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Direct PR from `fm/cli-surface-reverse-etl-api-surface-r1`, with API-confirmed PR base after opening.
- Task: Derive command `api_surface` from every declared endpoint summary independent of intent; regenerate affected artifacts; prevent an implemented command with a parseable endpoint summary but no API surface.
- Template fallback: `.agents/agentic-delivery/contracts/task-delivery-header-template.md` is absent from this integration-base snapshot. This header follows the current phase artifacts and records the required fields.

## Observed baseline

- `internal/connectors/defs/github/certification-sweep.json` has 228 implemented commands without `api_surface`.
- 214 summaries are parseable endpoint addresses: `METHOD /path`, optionally followed by one parenthesized write-action identifier. The suffix is metadata, not part of the endpoint.
- The other 14 are intentional friendly aliases and retain no API surface: `issue close`, `issue reopen`, `pr close`, `pr comment`, `pr lock`, `pr reopen`, `pr unlock`, `repo create`, `repo delete`, `repo archive`, `repo unarchive`, `secret set`, `secret delete`, and `cache delete`.
- Current generated sweep buckets are `fixture_required=1466`, `eligible_pending_live=25`, `not_applicable=50`, `schema_conformant=29`, and `provider_refused=1`: `1466 + 25 + 50 + 29 + 1 = 1571`.

## Causal path and decision

`cmd/connectorgen/surfacesync.go` currently filters the loop to `direct_read`, `direct_write`, and `binary_download` before deriving `api_surface`. The 228 GitHub reverse-ETL commands bind `writes.json` actions rather than operations, so they never reach that derivation even when their generated summary already records the endpoint.

The fix is a generic endpoint-summary derivation before intent-specific behavior. It recognizes a connector-relative `METHOD /path` prefix (with an optional terminal action annotation), synchronizes the single `api_surface` reference, and leaves ordinary human summaries untouched. Existing operation-derived endpoint synchronization remains authoritative for the operation-backed intents. This uses no connector-specific identifier or allowlist and does not assign an endpoint to the 14 aliases.

`validate.go` already consumes endpoint coverage when an API-surface reference exists; it does not require one for `reverse_etl`, which explains why the field-placement defect passed validation. The new repository-wide test closes that gap without changing reverse-ETL execution behavior.

## GSD execution note

`scripts/gsd doctor`, all required `scripts/gsd sources` calls, `go run ./cmd/agentcontractgen check`, and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved. The canonical contract prohibits GSD role spawning, so this direct-PR worker executes the prompts inline and records the manual fallback here and in the PR body.

## Discussion resolution

The brief resolves the only material design choice: derive an address once from the generic command summary predicate, never through a GitHub-specific exception and never by assigning aliases a guessed endpoint. No runtime write or credentialed provider operation is in scope.
