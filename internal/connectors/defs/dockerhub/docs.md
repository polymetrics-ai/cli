# Overview

Docker Hub exposes four public ETL streams plus source-backed organization,
SCIM, audit-log, and access-token-metadata reads. It also declares 20 ordinary
mutations as typed reverse-ETL actions, including six deletes, and two
approval-gated SCIM direct writes. Every runnable endpoint is bound to the
pinned public OpenAPI description at
`sources/dockerhub-operation-source-lock.json`.

Service API documentation: https://docs.docker.com/reference/api/hub/latest/.

## Auth setup

Connection fields:

- `docker_username` (required) selects the namespace used by the public ETL
  streams.
- `repository` and `tag` select the detail streams when those streams are run.
- `base_url` defaults to `https://hub.docker.com/v2` and is only for an
  explicitly configured proxy or test endpoint.
- `page_size` controls the first request for the public paginated ETL streams.
- `access_token` is an optional stored secret. When present, the connector
  sends it as the pinned Docker Hub bearer scheme. Never pass it inline.
- `password`, `login_2fa_token`, `code`, and `secret` are optional `x-secret`
  fields that exactly name pinned credential-minting request material. The
  corresponding routes remain declared-but-disabled until their exact pinned
  response-secret field and encrypted-store declarations are completed.

Public repository and tag reads do not require a token. Organization and SCIM
commands require a bearer token with the provider's applicable role or scope;
the provider returns its authorization response (including 403) at runtime.

## Streams notes

The public ETL streams are `repositories`, `tags`, `repository_detail`, and
`tag_detail`. `repositories` and `tags` follow the source response's `next`
URL; the detail streams are single-object responses.

Direct reads expose the audit action catalog and paginated audit logs,
organization settings, members, invites, groups, group members, personal
access-token metadata, organization access-token metadata, and the declared
SCIM discovery, individual-resource, and paginated-user endpoints. Their
flags are derived from the pinned operation contract.
Paging stays on the direct-read `--page` / `--page-cursor` contract; no raw
source cursor flag is exposed.

## Write actions & risks

The following typed, source-backed reverse-ETL actions are runnable through
the required plan, preview, approval, and execute lifecycle:

- organization settings update; repository create; repository group create;
  immutable-tag update and verification;
- organization member update and delete; organization group create, replace,
  update, and delete; group-member add and delete;
- organization invite bulk-create, resend, and delete.
- personal access-token update and delete; organization access-token update
  and delete.

SCIM user create and update use the source-declared `application/scim+json`
media type through the same typed direct-write plan, preview, approval, and
execute lifecycle. They are direct writes, not reverse-ETL destination
bindings.

The six delete actions carry typed destructive confirmation. These ordinary
provider-authorized mutations are not disabled merely because a user's token
may lack the necessary role.

Access-token list, detail, update, and delete operations are runnable: their
pinned responses expose metadata, not a token secret. The two access-token
create operations remain `declaration-pending` because their pinned responses
return `token`. Login, 2FA login, and auth-token exchange likewise remain
`declaration-pending` until their exact response-secret field and encrypted
store declaration is complete. None is classified `unsafe-to-exercise`.

## Known limits

- The pinned document accounts for 54 operations. Forty-five are runnable
  (four ETL commands, 19 direct reads, two direct writes, and 20 reverse-ETL
  commands); nine have evidence-backed disabled dispositions in
  `sources/dockerhub-declaration-disposition.json`.
- PR #4297 made operation-scoped pagination and the closed
  `application/scim+json` write type executable, so the audit-log, SCIM-user,
  and SCIM-write contracts are now runnable. Three `HEAD` status checks and
  the bounded members CSV export remain disabled only because
  `internal/connectors/engine/bundle.go:2451,2676,2705,2733` omits the
  otherwise implemented `rest_status` and `text_export` kinds from its
  operation-kind loader/validation path. This is the recoverable
  `operation-kind-loader-registration` foundation integration gap, not a
  provider or connector schema limitation.
- `sync_transport.json` declares the definition-owned ETL source transport.
  Reverse-ETL eligibility remains blocked on the connector-neutral typed
  destination executor.
- `certification-sweep.json` records non-live declaration status. Live
  credentialed certification remains pending and is not attempted here.
