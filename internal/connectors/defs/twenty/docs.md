# Twenty CRM connector

## Overview

Twenty CRM is exposed as a declarative connector for the 28 documented REST object collections:
companies, people, opportunities, notes, tasks, messages, calendar objects, workflow objects,
workspace members, and supporting association objects. The connector supports ETL list streams and
reverse ETL create, update, batch-create, and delete actions for those objects.

Official source inventory: Twenty's API overview at https://docs.twenty.com/developers/extend/api
states that each workspace has schema-per-tenant REST and GraphQL APIs under `/rest/` and
`/graphql/`, with bearer API-key auth, cloud base URL `https://api.twenty.com/`, batch limits of 60
records, and CRUD support for records. Because Twenty does not publish one static workspace API
reference, this bundle uses the standard-object metadata from the open-source Twenty repository and
keeps GraphQL as documented source context rather than duplicating the same CRUD operations as a
second unsafe/raw surface.

## Auth setup

Twenty uses bearer authentication with an `api_key` secret and an optional `base_url` config value.
Provide the secret from an environment variable or stdin; do not paste it into prompts, commit it,
or print it in logs. Inspecting the connector manual or command surface does not read credentials.

## Streams notes

Each stream maps to `GET /rest/<TwentyObject>` and reads records from the Twenty `data.<object>`
envelope with cursor pagination using `pageInfo.endCursor` and `pageInfo.hasNextPage`. The CLI
surface exposes those streams through commands such as `pm twenty companies list --json`.
Get-by-id endpoints are implemented as bounded, operation-backed direct reads. Each declared
endpoint uses its provider-cased REST path (for example `/rest/calendarEvents/{id}`), requires an
ID, and returns the runtime's redacted JSON policy.

## Write actions & risks

For every Twenty object, reverse ETL declares:

- `create_<object>` for `POST /rest/<TwentyObject>`;
- `update_<object>` for `PATCH /rest/<TwentyObject>/{id}`;
- `batch_<object>` for `POST /rest/batch/<TwentyObject>`;
- `delete_<object>` for `DELETE /rest/<TwentyObject>/{id}`.

Reverse ETL remains plan, preview, approval, then execute. Delete actions are typed destructive
actions and require `--confirm destructive`. Batch actions accept one declared `--records` JSON
array, validated against the action's closed record schema and capped at Twenty's documented 60
records; this is not a generic raw request-body escape hatch.

## Command Surface

`pm help twenty`, `pm twenty`, and `pm twenty --help` render the Twenty connector manual and command
surface without credentials. Implemented list commands read ETL streams; get commands use bounded
direct reads; create, update, batch, and delete commands remain subject to reverse-ETL plan,
preview, approval, and execute safeguards.

The documented batch commands stay implemented and user-reachable. Their
application batch-dispatch certification is explicitly deferred to 0.3.1:
`write_eligibility.json` source-traces every batch `POST /rest/batch/...`
route to the published Twenty API, locks the provider-owned `records` envelope
and 60-record limit, and names the foundation block. This is neither a partial
command classification nor a claim that batch delivery was certified.

## Fixture conformance and certification

S7 adds synthetic, credential-free replay fixtures for all 28 read streams and all 112 write actions.
Stream fixtures mirror Twenty's `data.<object>` envelope and cursor-style `pageInfo` response shape.
Write fixtures exercise create, update, batch, and delete request construction against the replay
capture server only; they do not execute live reverse ETL writes.

Local conformance is fixture-backed, runs without secrets, and is included in credential-free
certification when the Twenty bundle fixtures are available. Live `pm connectors certify twenty`
remains credential-gated for real Twenty tenants; use placeholder env values only with localhost
fixture replay. Reverse ETL still must follow plan, preview, approval, and execute before any live
mutation.

## Parity-deviation ledger

None for the declarative REST object surface covered by this bundle: 28 streams and 112 write
actions have request-shape fixtures. Twenty's GraphQL API remains documented source context rather
than a duplicate, raw parallel command surface for the same REST CRUD operations.

## Known limits

- No live Twenty credentials are required for connector inspection, help rendering, docs generation,
  validation, or fixture conformance.
- Destructive delete actions are declarative only until a user follows reverse ETL plan, preview,
  approval, and execute.
- The provider source used by this bundle does not document an idempotency-key
  header for `POST /rest/companies`. Twenty therefore does not declare the
  generic typed-destination application transport: doing so would invent a
  retry/delivery guarantee. This does not alter the implemented, approval-gated
  direct typed command.
