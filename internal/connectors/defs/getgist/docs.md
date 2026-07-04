# Overview

GetGist (Gist, getgist.com) exposes contacts, marketing, conversations, knowledge base, team, workspace, and e-commerce APIs. This bundle expands beyond the legacy six read streams to cover the documented JSON REST surface with declarative streams and write actions.

## Auth setup

Provide a Gist API key via the `api_key` secret; it is used only for Bearer auth and is never logged. `base_url` defaults to `https://api.getgist.com` and may be overridden for tests or proxies.

## Streams notes

Legacy streams keep their projected schemas: `contacts`, `tags`, `segments`, `campaigns`, `forms`, and `teammates`. New list/detail streams use the wrapper keys documented by Gist, for example `articles`, `article`, `collections`, `conversation`, `messages`, and e-commerce singular resource wrappers. Config-scoped detail streams require the matching id config key when read.

## Write actions & risks

Write actions create, update, delete, tag, subscribe, assign, reply to, and otherwise mutate Gist resources. Destructive delete actions are marked with destructive confirmation and idempotent 404 handling where the target is already absent. Reverse ETL writes require plan preview and approval.

## Known limits

- `/conversations/search` is excluded because it is a read-like POST endpoint with request-body search criteria; the current engine read path does not send `stream.body` payloads.
- Webhook setup and webhook data sections are excluded from ETL coverage because they describe application callback configuration and inbound notification payloads, not resources read from Gist.
- `contacts` still carries `x-cursor-field: updated_at` as legacy catalog metadata, but no stream sends an incremental lower-bound parameter because legacy did not filter server-side or client-side.
