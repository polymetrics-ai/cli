# Overview

Reads Less Annoying CRM users, contacts, tasks, notes, and events through fixed v2 RPC requests.

Readable streams: `users`, `contacts`, `tasks`, `notes`, `events`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.lessannoyingcrm.com/help/topic/API.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Programmer API key sent in the declared `Authorization` header.
- `start_date` (required, string); retained initial ETL lower-bound configuration.

Secret fields are redacted in logs and write previews: `api_key`.

The runtime uses only the fixed `https://api.lessannoyingcrm.com/v2` origin and the declared API-key header authentication. It does not accept caller-provided origins or fixture modes.

Connection checks send a bounded `GetUsers` request.

## Streams notes

Default pagination: numbered `Page` requests with `MaxNumberOfResults=500`.

Incremental streams use their declared cursor fields and client-side lower-bound filtering when a lower bound is available.

- `users`: POST `GetUsers`; records at the response root.
- `contacts`: POST `GetContacts`; records at `Results`.
- `tasks`: POST `GetTasks`; records at `Results`; incremental cursor `DateCreated`.
- `notes`: POST `GetNotes`; records at `Results`.
- `events`: POST `GetEvents`; records at `Results`; incremental cursor `DateUpdated`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint groups.
