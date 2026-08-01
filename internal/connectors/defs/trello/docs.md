# Trello connector docs

## Overview

The Trello connector is generated from Atlassian's official Trello OpenAPI document and covers the supportable REST API surface without live provider calls in conformance. This parity wave audited `https://developer.atlassian.com/cloud/trello/swagger.v3.json` at SHA-256 `b50fca38c5ea62025f9778482f89f11ae3da0dd983d31ba49401c4422e450b19` and partitions all 261 official HTTP operations into 219 executable connector operations (3 fixture-backed streams, 95 fixed JSON direct reads, and 121 fixture-backed typed writes) plus 42 blocked operation-ledger rows.

Executable streams cover the high-volume board/list/checklist ETL paths. Parameterized GET endpoints are exposed as fixed direct-read commands with closed path/query flags and `json_redacted` output policy. Executable writes are typed reverse ETL actions using closed record schemas for supportable Trello POST, PUT, and DELETE operations.

## Auth setup

Create a Trello API key and token out of band, then add credentials without placing secret values in prompts or command history. Use environment variables or stdin only:

```bash
pm credentials add trello-prod --connector trello --from-env key=TRELLO_API_KEY --from-env token=TRELLO_API_TOKEN
```

The connector injects both `key` and `token` through connector-local query authentication so direct reads, streams, and writes carry the required Trello credentials without exposing secret values in command flags. Both fields are marked `x-secret` and are redacted from previews and errors.

## Streams notes

The fixture-backed ETL streams are `boards`, `lists`, and `checklists`. The `boards` stream reads `/members/me/boards`; `lists` and `checklists` read board-scoped endpoints and use `--config id=<board-id>` for the board id. The remaining supportable GET endpoints are fixed direct-read commands under `pm trello read ...` with explicit path/query flags.

```bash
pm trello read boards --credential trello-prod --json
pm trello read lists --credential trello-prod --config id=<board-id> --json
pm trello read get-cards-id --credential trello-prod --id <card-id> --json
pm trello read get-search --credential trello-prod --query <search-text> --json
```

Fixture pages under `fixtures/streams/**` are sanitized synthetic pages for local conformance replay only; they are not live Trello captures and contain no real board, member, organization, card, webhook, or token data.

## Write actions & risks

All supportable Trello mutations are exposed as reverse ETL write actions, never as raw HTTP. CLI write commands create a plan first, support preview, require explicit approval, and only execute after approval. DELETE actions declare destructive confirmation plus idempotent missing-resource handling for `404`.

Representative examples:

```bash
pm trello write create-card --credential trello-prod --id-list <list-id> --name "New card" --preview --json
pm trello write comment-card --credential trello-prod --id <card-id> --text "Fixture-safe comment" --preview --json
pm trello write delete-webhook --credential trello-prod --id <webhook-id> --preview --json
```

The declarative engine builds Trello write requests from typed records. Form and bounded multipart bodies carry supportable Trello parameters, including file upload fields for card attachments, card file sources, board backgrounds, custom emoji, custom stickers, avatars, and organization logos. Destructive or notification-producing effects remain gated by reverse ETL plan → preview → approval → execute.

## Known limits

The official operation ledger intentionally blocks 42 endpoints:

- Trello Enterprise administration endpoints remain blocked because they require elevated enterprise authority and mutate organization/member/admin state.
- Token-management endpoints remain blocked because they expose or mutate credential/token state.
- Application compliance is blocked as an elevated compliance direct-read surface.
- `/batch` is blocked because it accepts arbitrary sub-request URLs and would be a raw generic HTTP escape hatch.
- Field/filter accessor endpoints such as `/cards/{id}/{field}` are tracked as duplicates of covered object or collection endpoints.

No credentialed Trello checks, live provider calls, or provider writes are part of local conformance. Binary response payloads were not found in the audited OpenAPI content types; OpenAPI file upload request parameters are modeled as bounded multipart reverse ETL actions rather than raw HTTP uploads.
