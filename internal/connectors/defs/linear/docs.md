# Overview

Linear is modeled from the official Linear GraphQL schema, pinned to blob `e92dc40c31e3b6e3962f93fa1d8cbe91f3e83034` from master commit `7ef4c5024f88667b2c85057ff4c905676c4a93c2` (`packages/sdk/src/schema.graphql`).

This connector-local bundle now inventories every parsed root GraphQL operation in `api_surface.json` operation-ledger mode. It implements fixed GraphQL ETL streams for list/connection Query fields and fixed GraphQL reverse-ETL write actions for scalar-required mutations that can be represented safely by the current declarative engine. It does not expose a raw GraphQL query/mutation/body escape hatch.

Connector-local generated counts from the pinned blob:

| query root fields | mutation root fields | subscription root fields | surface rows | streams | write actions | blocked operation rows |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 165 | 372 | 80 | 617 | 64 | 116 | 437 |

The GitHub parent/subissue r2 dispatch counts remain preserved in those issue bodies. This connector-local evidence does not claim live certification and does not fabricate implemented counts beyond the concrete stream/write rows present in this bundle.

## Auth setup

Connection fields:

- `access_token` (secret): Linear OAuth access token. When present, it is sent as a Bearer Authorization header and takes priority.
- `api_key` (secret): Linear personal API key. By default it is sent as the bare Authorization header value; set `auth_type` to `oauth` or `oauth2.0` to send it as Bearer.
- `auth_type` (optional): ``, `oauth`, or `oauth2.0`; default ``.
- `base_url` (optional): full GraphQL endpoint, default `https://api.linear.app/graphql`.
- Fixed stream documents request 50 records per GraphQL connection page.

Never place secret values in command text, fixtures, docs, or issue bodies. Add credentials from environment variables or stdin through the credential manager.

## Streams notes

Streams are fixed GraphQL Query documents generated from root fields returning connection/list object data. Connection streams use `first` and `after` variables with cursor pagination capped at 100 pages per read; list streams are single-request full refreshes. Runtime request bodies are fixed in `streams.json` and only declared variables are populated.

Representative streams include `issues`, `teams`, `projects`, and `users`; additional generated streams cover other documented list/connection root Query fields from the pinned schema. Paginated stream fixtures under `fixtures/streams/*/{page_1,page_2}.json` are sanitized synthetic GraphQL response shapes with `hasNextPage`/`endCursor` transitions for local conformance only.

## Write actions & risks

`writes.json` contains 116 fixed GraphQL reverse-ETL actions whose complete argument list is required, scalar, non-secret-shaped, and not provider-internal in the schema. The action document is connector-owned metadata; callers provide only typed record fields declared in `record_schema`. `cli_surface.json` declares one canonical provider command for each action.

93 write action(s) carry `confirm: "destructive"` for delete/archive/remove/revoke/rotate/cancel/disconnect/merge-style mutations. These operations are in scope under the captain policy, but they execute only through reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation. Their canonical commands include the same typed destructive approval language. Blocked mutation rows in `api_surface.json` are not excluded as unsafe; they name the missing shared foundation or provider-internal policy instead.

Fixture write captures under `fixtures/writes/*.json` are synthetic replay examples. They do not perform live Linear writes.

## Known limits

- No live Linear provider calls, credentials, writes, subscriptions, or certification were run.
- `api_surface.json` uses operation-ledger blocked rows instead of legacy `excluded` rows for direct/binary/CDC and unsupported mutation shapes.
- GraphQL direct-read/query/search and binary execution remains blocked by the provider search/query foundation (#2985); `operations.json` and `cli_surface.json` keep planned commands bounded and fixed-document only.
- Linear subscriptions/changefeeds remain blocked by CDC foundations (#2986/#2988); no connector-local stream pretends to execute a live subscription.
- Input-object, optional-argument, list, JSON, secret-sensitive, deprecated, and provider-[INTERNAL] mutation shapes remain blocked until shared fixed GraphQL variable/object support can omit absent optional `record.*` fields safely without a raw GraphQL escape hatch or provider-specific hook.
- Declarative `Check()` cannot send a fixed GraphQL body, so `capabilities.check` is false in this bundle; fixture-only stream/write replay is not live health certification.
