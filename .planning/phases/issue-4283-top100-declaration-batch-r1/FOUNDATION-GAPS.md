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
