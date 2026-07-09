# Plan: Issue #134 HubSpot CLI Surface Metadata

Date: 2026-07-10
Parent issue: #132
Sub-issue: https://github.com/polymetrics-ai/cli/issues/134
Parent branch: `feat/132-hubspot-cli-parity`
Suggested sub-branch: `feat/134-hubspot-cli-surface-metadata`

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed.
- `scripts/gsd prompt programming-loop init --phase issue-134-hubspot-cli-surface-metadata --dry-run` is expected to be unavailable for the same pinned-registry reason recorded in the parent plan (`programming-loop` is not a registered command).
- Active fallback: manual universal programming loop with TDD evidence, using parent `plan-phase` and `execute-phase` prompts generated through `scripts/gsd`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-graphql`
- `golang-lint`

## Objective

Create validated HubSpot command-surface metadata from the official HubSpot OpenAPI inventory. The first slice should be metadata and validation only; execution remains blocked until the runner/direct-read/stream/write lanes implement safe dispatch.

## Scope for this slice

In scope:

- Add/adjust CLI surface schema validation so HubSpot can classify binary/file commands explicitly as `binary` instead of smuggling them through `direct_read`, `local_workflow`, or `direct_write`.
- Add a HubSpot bundle scaffold under `internal/connectors/defs/hubspot/` with `metadata.json`, `spec.json`, `streams.json`, `docs.md`, `cli_surface.json`, and any typed `operations.json` needed for command metadata references.
- Populate representative HubSpot provider-like command families from official OpenAPI areas:
  - CRM objects: contacts, companies, deals, tickets, line items, products, quotes.
  - CRM metadata: properties, pipelines, associations, owners.
  - Marketing and automation: forms, marketing emails, lists, workflows/events where present in specs.
  - Commerce: payment links, invoices, discounts, fees, tax rates.
  - Files and binary/file-like operations.
  - Settings/admin/sensitive areas as planned reverse ETL or issue-linked blockers; never as raw writes.
- Add tests that fail before implementation and pass after implementation.

Out of scope for this slice:

- Live HubSpot credentials.
- Executing HubSpot reads/writes.
- Full stream schemas/fixtures (issue #136).
- Bounded direct-read executor policies beyond metadata (issue #138).
- POST body/query execution engine work (issue #139).
- Sensitive/admin write execution policy beyond metadata classification (issue #140).
- Help renderer/docs/website parity beyond connector docs notes (issue #135), except not-applicable evidence.

## Full-surface safety interpretation for #134

- Sensitive/admin/destructive operations must be classified as typed reverse ETL or typed operation blockers with issue-linked evidence, not permanently excluded for risk alone.
- Binary operations must be classified as `binary` metadata with bounded policy follow-up, not as local workflows or generic downloads.
- Commands must not use `direct_write`, `raw_api`, generic HTTP, arbitrary GraphQL mutation body, generic shell write, or generic SQL write.
- Planned commands may be present for discoverability, but must not be executable until their safe lane lands.

## TDD plan

1. Red: add a connectorgen validation test proving `intent: "binary"` with a typed `binary_download` operation passes.
2. Red: add a validation test proving implemented `binary` commands without a typed operation are rejected.
3. Red: add a HubSpot CLI-surface metadata test proving the bundle exists, validates, has expected command groups/intents, and contains no `raw_api` or `direct_write` commands.
4. Green: update schema/validator and add HubSpot bundle metadata.
5. Refactor: keep generated/large metadata deterministic and source-commented through docs/plan evidence.

## Targeted verification

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'CLISurface|HubSpot'
go test ./internal/connectors/engine -run 'CLISurface|HubSpot'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Broader verification before sub-issue handoff

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## CLI help/docs/website parity

Applies partially. `cli_surface.json` is connector metadata and may render in existing connector guide/manual surfaces, but this slice does not add an executable `pm hubspot ...` dispatcher. Record as follows:

- Runtime help (`pm help hubspot`, `pm hubspot --help`): not applicable until #135/#136 introduce a user-visible dispatcher.
- Bare namespace behavior: not applicable until dispatcher exists.
- `docs/connectors/hubspot/**`: update if generated manual/skill output changes in this slice.
- `docs/cli/**` and `website/**`: not applicable unless #135 help renderer is pulled into scope.

## Human gates

- New dependencies.
- Live credentials or live HubSpot API checks.
- Auth scope changes.
- Destructive/admin external actions.
- Reverse ETL execution.
- Binary transfer runtime enablement without explicit destination policy.
- Generic raw write tools.
