# Twenty CRM connector

## Overview

Twenty CRM is exposed as a declarative connector for the 28 documented REST object collections:
companies, people, opportunities, notes, tasks, messages, calendar objects, workflow objects,
workspace members, and supporting association objects. The connector supports ETL list streams and
reverse ETL create, update, batch-create, and delete actions for those objects.

## Auth setup

Twenty uses bearer authentication with an `api_key` secret and an optional `base_url` config value.
Provide the secret from an environment variable or stdin; do not paste it into prompts, commit it,
or print it in logs. Inspecting the connector manual or command surface does not read credentials.

## Streams notes

Each stream maps to `GET /rest/<TwentyObject>` and reads records from the Twenty `data.<object>`
envelope with cursor pagination using `pageInfo.endCursor` and `pageInfo.hasNextPage`. The CLI
surface exposes those streams through commands such as `pm twenty companies list --json`.
Get-by-id endpoints (`GET /rest/<TwentyObject>/{id}`) are documented in the command surface as
planned direct-read commands because generic JSON direct-read output policies are outside S6.

## Write actions & risks

For every Twenty object, reverse ETL declares:

- `create_<object>` for `POST /rest/<TwentyObject>`;
- `update_<object>` for `PATCH /rest/<TwentyObject>/{id}`;
- `batch_<object>` for `POST /rest/batch/<TwentyObject>`;
- `delete_<object>` for `DELETE /rest/<TwentyObject>/{id}`.

Reverse ETL remains plan, preview, approval, then execute. Delete actions are destructive and require
typed `--confirm destructive` approval. Batch actions send a validated top-level `records` array
through reverse ETL records; the S6 CLI surface intentionally does not expose a generic JSON/raw
write flag for batch payloads.

## Command Surface

`pm help twenty`, `pm twenty`, and `pm twenty --help` render the Twenty connector manual and command
surface without credentials. Implemented list commands read ETL streams. Implemented create, update,
and delete commands plan reverse ETL writes from commandrunner-safe scalar flags. Batch commands are
marked partial until scalar CLI coercion for `records: []object` is approved.

## Fixture conformance and certification

S7 adds synthetic, credential-free replay fixtures for all 28 read streams and all 112 write actions.
Stream fixtures mirror Twenty's `data.<object>` envelope and cursor-style `pageInfo` response shape.
Write fixtures exercise create, update, batch, and delete request construction against the replay
capture server only; they do not execute live reverse ETL writes.

Local conformance is fixture-backed and safe to run without secrets. Live `pm connectors certify
twenty` remains credential-gated and currently needs certify-harness follow-up for Twenty's camelCase
`updatedAt` cursor fields and longest stream names before it can be treated as a green live certificate.
Reverse ETL still must follow plan, preview, approval, and execute before any live mutation.

## Parity-deviation ledger

- None for the declarative REST object surface covered by this bundle: 28 streams and 112 write
  actions have request-shape fixtures.
- Generic JSON direct-read execution for get-by-id commands remains outside S6/S7 and is documented
  as `planned` in `cli_surface.json` rather than exposed as an executable raw HTTP surface.
- Batch write scalar CLI coercion remains intentionally deferred; use reverse ETL records with a
  top-level `records` array instead of a raw JSON flag.

## Known limits

- Generic JSON direct-read execution for get-by-id commands is not implemented in S6/S7.
- Batch write execution through scalar CLI flags is deferred; use reverse ETL records with a
  top-level `records` array instead of a raw JSON flag.
- No live Twenty credentials are required for connector inspection, help rendering, docs generation,
  validation, or fixture conformance.
- Destructive delete actions are declarative only until a user follows reverse ETL plan, preview,
  approval, and execute.
