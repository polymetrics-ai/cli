# Foundation gaps — increment 001

The machine-readable source is `FOUNDATION-GAP-REASONS.json`. This document
states the current, non-live disposition after PR #4286; source composition is
declared for all ten bundles and is not an open foundation gap.

## Generic typed reverse-ETL eligibility destination

- ID: `generic-typed-destination-executor`
- Affected connectors: Docker Hub, Notion, Stripe, Bitbucket, GitLab,
  CircleCI, Sentry, Vercel, Asana, and Jira.
- Evidence: `internal/app/issue_label_warehouse_transport.go:85-95` supplies
  only the closed `issue_label_destination`; its builder invokes
  `issueLabelTransportConnectorContract`.
- Recovery: register a connector-neutral typed destination
  `DefinitionFactory` selected by the definition, with per-connector evidence,
  explicit source bindings, acknowledgement and per-mode apply strategies.

This is an eligibility gap for every direct-write endpoint, not a second
endpoint class. A typed write action remains a direct write until the full
destination contract exists.

## Docker Hub response-less HEAD operations

- ID: `head-response-less-operation-executor`
- Affected operations: `HEAD /v2/namespaces/{namespace}/repositories/{repository}`;
  `HEAD /v2/namespaces/{namespace}/repositories/{repository}/tags`; and
  `HEAD /v2/namespaces/{namespace}/repositories/{repository}/tags/{tag}`.
- Evidence: `internal/connectors/engine/bundle.go:2676-2681` admits only GET
  and POST as `rest_read`.
- Recovery: add a typed bounded HEAD/status executor and distinct terminal
  command contract.

## Docker Hub operation-scoped pagination

- ID: `operation-scoped-rest-pagination`
- Affected operations: `GET /v2/auditlogs/{account}` and `GET
  /v2/scim/2.0/Users`.
- Evidence: `internal/connectors/engine/direct_read_paginate.go:126-130`
  consumes only `streams.json` base pagination.
- Recovery: add source-derived per-operation `PaginationSpec` support to the
  direct-read paginator.

The Docker Hub SCIM media-type and secret-response records remain
source-operation declarations, not additional transport claims; their exact
dispositions and recoverability remain in `REJECTION-LIST.json`.

## Docker Hub source-import YAML key normalization

- ID: `source-import-yaml-scalar-key-normalization`
- Affected operation inventories: all 54 operations in Docker Hub's locked
  `latest.yaml` and all 249 operations in Asana's locked OpenAPI document;
  descriptor generation stops before it can preserve any one operation.
- Evidence: `cmd/connectorgen/sourceimport.go:1053-1065` parses the exact
  locked artifact and `cmd/connectorgen/sourceimport.go:1305-1318` rejects a
  numeric YAML scalar mapping key. The non-credentialed reproduction is
  `connectorgen source-import dockerhub`, which refuses the numeric response
  key at the immutable-tags PATCH response map.
- Recovery: make the shared importer canonicalize scalar mapping keys to their
  exact string spelling where the OpenAPI grammar names JSON object members,
  while retaining duplicate and non-scalar key refusals. A connector-local
  source rewrite or manually invented descriptor is not valid provenance.

## Main-base source-lock refresh refusals

- The source importer correctly refuses a changed public artifact rather than
  silently replacing a pin. Its exact `source-lock refresh required: fetched
  artifact does not match locked bytes and SHA-256` result now affects Notion,
  Bitbucket, GitLab, CircleCI, Vercel, and Jira.
- This is distinct from an importer/parser failure. Updating any URL, byte
  count, SHA-256, or documented-operation denominator is a source-lock
  refresh decision and is not performed in this reconciliation lane without
  captain direction and a complete re-derivation.

## Stripe recursive source-schema refusal

- Stripe reaches its locked source, but importer preflight rejects the public
  OpenAPI reference cycle `#/components/schemas/file` while examining
  `GET /v1/account` response `200`. No descriptor is emitted.
- This is distinct from source-lock drift and the YAML scalar-key limitation;
  the closed source importer needs a cycle-safe representation that preserves
  the source contract rather than truncating or inventing that schema.

## Docker Hub SCIM body-schema dialect validation (separate)

- This is not the YAML mapping-key importer failure. `connectorgen validate`
  reaches the retained Docker Hub declarations and independently refuses the
  OpenAPI-only `example` keyword inside the SCIM request body schema.
- Exact non-credentialed evidence: command 43, `scim user create`, operation
  `dockerhub.post__v2_scim_2.0_users`, fails at
  `properties.name.properties.familyName`; command 44, `scim user update`,
  operation `dockerhub.put__v2_scim_2.0_users__id_`, fails at
  `properties.name.properties.familyName`, `properties.name.properties.givenName`,
  and `properties.schemas.items`, each with `unknown keyword "example"`.
- Status: genuine body-schema dialect/projection incompatibility, to be
  reconciled from the canonical descriptor after Docker Hub source import
  succeeds. Do not hand-edit derived body contracts or treat this as a source
  rewrite workaround.

## e338cd301 source-lock refresh results

- **Notion and Bitbucket:** their refreshed public artifacts pass hash/byte
  qualification, but response-schema expansion exceeds the shared 32-reference
  depth bound (`cmd/connectorgen/sourceimport.go:65,4269-4272`). The concrete
  failures are Notion `GET /v1/blocks/meeting_notes` response `200` and
  Bitbucket pull-request comments `GET` response `200`.
- **GitLab:** the provider document declares
  `POST /api/v4/groups/{id}/(-/)epics/{epic_iid}/issues/{epic_issue_id}`
  without a required `epic_issue_id` path parameter. The importer refuses that
  unrepresentable contract at `cmd/connectorgen/sourceimport.go:6036-6048`.
- **Vercel:** its public `POST /api-keys` response carries
  `patternProperties`, which the OpenAPI 3.0 source-schema grammar refuses at
  `cmd/connectorgen/sourceimport.go:4311-4315`. This occurs before the
  separately known read-only source-coverage work.
- **CircleCI and Jira:** both imports emit canonical descriptors, then stop at
  `cmd/connectorgen/sourceprojection.go:137-143,210-211` because 27 and 16
  source mutations respectively have no existing complete executable action.
  CircleCI includes its secret-bearing context environment-variable PUT; PR
  #4334 is verified open, unmerged, and behind main, but the missing-action
  count is broader than that one env-only requirement.
- No checked-in source lock is refreshed from these measurements: a valid
  refresh must retain a complete verified inventory and derived declaration,
  not only a new byte count and SHA-256.
