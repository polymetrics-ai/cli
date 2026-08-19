# Overview

Docker Hub exposes four public ETL streams plus source-backed organization,
SCIM, and access-token-metadata reads. It also declares 20 ordinary mutations
as typed reverse-ETL actions, including six deletes. Every runnable endpoint is bound
to the pinned public OpenAPI description at
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
  corresponding routes remain declared-but-disabled until the engine can
  execute a `sensitive_policy` and safely handle a returned credential.

Public repository and tag reads do not require a token. Organization and SCIM
commands require a bearer token with the provider's applicable role or scope;
the provider returns its authorization response (including 403) at runtime.

## Streams notes

The public ETL streams are `repositories`, `tags`, `repository_detail`, and
`tag_detail`. `repositories` and `tags` follow the source response's `next`
URL; the detail streams are single-object responses.

Direct reads expose audit action catalog, organization settings, members,
invites, groups, group members, personal access-token metadata, organization
access-token metadata, and the declared SCIM discovery and individual resource
endpoints. Their flags are derived from the pinned operation contract.
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

The six delete actions carry typed destructive confirmation. These ordinary
provider-authorized mutations are not disabled merely because a user's token
may lack the necessary role.

Access-token list, detail, update, and delete operations are runnable: their
pinned responses expose metadata, not a token secret. The two access-token
create operations remain declared as `foundation-gap` because their pinned
responses return `token`. Login, 2FA login, and auth-token exchange likewise
remain declared as `foundation-gap`: they require the named `sensitive_policy`
live path and safe secret response handling. None is classified
`unsafe-to-exercise`.

## Known limits

- The pinned document accounts for 54 operations. Forty-one are runnable
  (four ETL commands, 17 direct reads, and 20 reverse-ETL commands); 13 have
  evidence-backed disabled dispositions in
  `sources/dockerhub-declaration-disposition.json`.
- Three pinned `HEAD` operations need a response-less operation executor.
  Audit-log and SCIM-user collection paging need operation-scoped pagination.
  The SCIM writes use `application/scim+json`, which the typed JSON write
  executor does not accept. These are recoverable foundation/schema gaps.
- `sync_transport.json` is absent under recoverable foundation issue #4093.
- `certification-sweep.json` records non-live declaration status. Live
  credentialed certification remains pending and is not attempted here.
