# Foundation gaps — increment 001

This declaration-only increment does not claim a generic API sync transport. The complete, operation-level foundation reasons are machine-readable in `FOUNDATION-GAP-REASONS.json`; the entries retain their exact source operation and recovery condition. `TRANSPORT-GAP.md` records the explicit, recoverable transport decision for all ten connectors.

## Shared generic declarative API transport registration

- State: open, tracked by [#4093](https://github.com/polymetrics-ai/cli/issues/4093).
- Affected connectors: Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.
- Effect: no `sync_transport.json` is emitted for this cohort. Source ETL and reverse-ETL destination transport are marked `foundation-gap`, `recoverable: true` rather than claiming an executor or a conformance run which does not apply to these bundles.
- Evidence: `internal/synctransport/definition_composition.go:145-168` requires an exact registered factory and accepted evidence. `internal/app/issue_label_warehouse_transport.go:54-103` admits the source only under the GitHub-specific evidence constant, while lines 322-368 restrict the destination to the issue-label action contract. `internal/connectors/certify/stages_transport_internal_test.go:89` proves failure before any provider call if registration is missing.
- Recovery: deliver #4093 with a connector-neutral source factory and per-bundle evidence, plus a closed typed destination adapter with explicit bindings, acknowledgement, and per-mode strategies. Then derive each connector-local descriptor from its pinned source and rerun the non-live sweep.

## Schema-shaped operations that remain disabled

The rejection list classifies 1,602 operations as `schema-incompatible`. This increment intentionally did not create request, response, pagination, or body schemas where the pinned provider document lacked a bounded usable contract. Those entries become enabled only after the provider publishes a usable schema or the engine implements the documented shape.

## Docker Hub response-less HEAD operations

- State: open connector/runtime foundation gap, recorded in
  `FOUNDATION-GAP-REASONS.json` and Docker Hub's declaration-disposition
  ledger.
- Affected operations: `HEAD /v2/namespaces/{namespace}/repositories/{repository}`;
  `HEAD /v2/namespaces/{namespace}/repositories/{repository}/tags`; and
  `HEAD /v2/namespaces/{namespace}/repositories/{repository}/tags/{tag}`.
- Evidence: the pinned Docker Hub OpenAPI records these as HEAD existence checks
  with no response-body content contract. `internal/connectors/engine/bundle.go:2676-2681`
  admits only GET and POST for `rest_read`, refusing HEAD before a network call.
- Recovery: add a bounded typed HEAD/status executor and a separate terminal
  command contract that cannot be presented as a JSON direct read; then add
  connector-local source, fixture, and non-live evidence. Until then all three
  remain `foundation-gap`, disabled, and recoverable.

## Docker Hub operation-scoped pagination

- Affected operations: `GET /v2/auditlogs/{account}` and
  `GET /v2/scim/2.0/Users`.
- Evidence: the pinned document declares `page`/`page_size` for the audit-log
  collection and `startIndex`/`count` for the SCIM collection. The runtime
  reads only `streams.json` base pagination at
  `internal/connectors/engine/direct_read_paginate.go:126-130` and refuses a
  conflicting raw paging parameter at lines 181-183.
- Recovery: add source-derived per-operation `PaginationSpec` support to the
  direct-read paginator. The two disabled operations can then gain bounded
  direct-read commands without exposing raw cursors.

## Docker Hub SCIM write media type

- Affected operations: `POST /v2/scim/2.0/Users` and
  `PUT /v2/scim/2.0/Users/{id}`.
- Evidence: the pinned requests require `application/scim+json`, while
  `internal/connectors/engine/direct_write.go:639-672` admits only
  `application/json` and `application/x-www-form-urlencoded`.
- Recovery: add a closed `application/scim+json` branch to the typed executor,
  retain the pinned schemas, and add connector-local non-live proof. No generic
  media-type or raw-body escape hatch is acceptable.

## Docker Hub credential minting and secret responses

- Affected operations: `POST /v2/access-tokens`; `POST
  /v2/orgs/{name}/access-tokens`; `POST /v2/users/login`; `POST
  /v2/users/2fa-login`; and `POST /v2/auth/token`.
- Evidence: the pinned response schemas expose `token` or `access_token` for
  each create/exchange response. The login and exchange inputs are now exact
  `x-secret` fields in `spec.json`; the operation contracts declare
  `secret_sensitive` and `sensitive_policy`. Yet
  `internal/connectors/engine/bundle.go:2772-2776` says that this is schema and
  validator support only and that live secret writes remain blocked.
- Recovery: add a typed live secret-write execution path and a secret-response
  contract that redacts stdout, logs, and warehouse records and routes a
  returned credential to secure storage. Then enable the five already-declared
  contracts and add connector-local non-live proof.
- Disposition: these are `foundation-gap`, recoverable. The eight personal and
  organization access-token list/detail/update/delete routes are enabled now:
  their pinned responses return metadata rather than a secret. Docker Hub has
  zero `unsafe-to-exercise` rows.

## Docker Hub OpenAPI path-item parameter inheritance

- Evidence: `cmd/connectorgen/paramsimport.go:161-203` reads only
  `operation.Parameters` and lines 211-214 import that slice. Docker Hub
  carries required path parameters on path items, so this bundle derives the
  exact inherited parameter contracts directly from its pinned document.
- Recovery: merge path-item and operation parameters with operation-level
  overrides in `params-import`; then regenerate and check the command flags.
  This limitation does not disable an operation because the connector-local
  parameters remain exact source-derived declarations.
