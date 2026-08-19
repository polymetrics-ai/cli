# Sync transport disposition — increment 001

## Current state after PR #4286

PR #4286 landed on `main` at `acb85dc03`. It closes the prior estate-wide
source-registration concern: the production composition discovers every
definition-owned source declaration using
`declarative_api/declarative_stream_source` and collects that declaration's
own conformance evidence. Batch 1 therefore adds a source-only
`sync_transport.json` to Docker Hub, Notion, Stripe, Bitbucket, GitLab,
CircleCI, Sentry, Vercel, Asana, and Jira.

Each source declaration has a concrete allowlist equal to its loaded
`streams.json` names, the closed modes supported by the bounded declarative
collection adapter, and conservative `at_least_once`, `source_ordered`,
`not_available` delivery facts. `go test -timeout 20m ./internal/app -run
'^TestOpenRegistersDefinitionOwnedProductionTransports$'` proves production
composition opens with all declarations; `connectorgen validate` proves their
structural contract. This is non-live proof only.

## Remaining reverse-ETL eligibility gap

- ID: `generic-typed-destination-executor`
- Affected connectors: Docker Hub, Notion, Stripe, Bitbucket, GitLab,
  CircleCI, Sentry, Vercel, Asana, and Jira.
- State: `foundation-gap`, recoverable; zero direct-write endpoints are
  reverse-ETL-eligible.
- Evidence: `internal/app/issue_label_warehouse_transport.go:85-95` registers
  the only declarative destination, `issue_label_destination`, and its
  `BuildDestination` calls `issueLabelTransportConnectorContract`. That closed
  contract needs an issues stream plus apply, replace, and cleanup label
  actions. The ten `writes.json` files have no `transport_binding` action.
- Minimal change: register a connector-neutral typed destination
  `DefinitionFactory` selected by the definition, with per-connector evidence,
  explicit source bindings, acknowledgement and per-mode apply strategies.

Every write endpoint remains classified as `direct_write`, including an enabled
typed create, update, upsert, or delete action. Reverse ETL is a separate
eligibility attribute, not a second class for that endpoint. No destination
descriptor, typed action binding, acknowledgement, or strategy is fabricated
from REST documentation. Live credentialed certification remains pending for
every connector.
