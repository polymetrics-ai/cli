Refs #4015

## Summary

`surface-sync` now derives `api_surface` from a declared endpoint summary for every implemented intent, before intent-specific operation policy runs. It only binds an exact `METHOD /path` match that is already declared in the bundle's `api_surface.json`, so the generator does not invent endpoints from summary-shaped prose.

Regenerating GitHub recovered the 214 reverse-ETL commands that carried their address in `summary` instead of `api_surface`. The remaining 14 empty records are intentional friendly aliases and are unchanged: `cache delete`, `issue close`, `issue reopen`, `pr close`, `pr comment`, `pr lock`, `pr reopen`, `pr unlock`, `repo archive`, `repo create`, `repo delete`, `repo unarchive`, `secret delete`, and `secret set`.

## TDD evidence

- **Red:** The all-bundle invariant ran against the unmodified broken generated state and failed on 248 endpoint-like summaries: the 214 GitHub defects plus 34 punctuated Workday strings. The focused one-command reverse-ETL fixture independently failed with `api surface fills = 0`. This demonstrated the new guard failing without hand-editing a generated artifact.
- **Green:** The generator joins a summary endpoint to its canonical `api_surface.json` record. The invariant and focused test pass after regeneration; explicit punctuation and friendly-alias cases remain unbound.

The 34 Workday strings do not exactly match their canonical endpoint due to sentence punctuation, so this change intentionally does not guess or modify them. That keeps the new shared behavior source-driven rather than connector-specific.

## Certification accounting

No certification state changed. Before and after the GitHub sweep buckets are:

| Status | Before | After | Delta |
| --- | ---: | ---: | ---: |
| `fixture_required` | 1,466 | 1,466 | 0 |
| `eligible_pending_live` | 25 | 25 | 0 |
| `not_applicable` | 50 | 50 | 0 |
| `schema_conformant` | 29 | 29 | 0 |
| `provider_refused` | 1 | 1 | 0 |
| **Total** | **1,571** | **1,571** | **0** |

## Rules for the rulebook

- Derive common command metadata before intent-specific policy; a routable endpoint belongs to any implemented command that has a declared address, not just the intent that first needed it.
- A summary-derived endpoint must join the connector's canonical `api_surface.json`. A raw-looking string alone is not authority to invent a route.
- Friendly aliases and sentence-punctuated summaries must remain unbound unless a one-to-one canonical endpoint is explicitly declared.
- Prove an invariant red on the real broken input, then green after regeneration; generated artifacts are evidence, never hand-edited fixtures.
- Metadata repair that only restores routing facts must preserve certification status-bucket arithmetic and explain every delta.

## Verification

Passed:

- `go test -timeout 20m ./cmd/connectorgen`
- `go test -timeout 20m ./internal/cli`
- `go vet ./...`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`
- `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`
- `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-sweep`
- `make connector-boundary`, `make connector-runtime-preflight`, `make connector-canon-check`, `make release-workflow-check`
- Website docs, website data, tracked-skills, surface-sync, and certification-sweep generators each ran twice with byte-stable second passes.

`make verify` was decomposed into its individual repository gates because this repository's own instructions prohibit running its long aggregate test suite under the per-command timeout. No gate was skipped; the changed package and `internal/cli` were run separately with `-timeout 20m`.

## Review route

Local code review found no actionable issue. This non-default-base PR records `claude_auto` as the primary external-review route; the automatic trigger is expected on PR open. No manual Claude or Copilot request has been made.

`security/snyk` is known to fail identically on the base branch and is pre-existing.
