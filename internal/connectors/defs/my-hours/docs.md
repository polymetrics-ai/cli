# Overview

Reads My Hours clients, projects, users, tags, and bounded time-log activity through fixed REST routes.

Readable streams: `clients`, `projects`, `users`, `tags`, `time_logs`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.myhours.com/.

## Auth setup

Connection fields:

- `email` (required, string); My Hours login email.
- `password` (required, secret, string); password sent only to the fixed declared login route.
- `start_date`, `end_date` (required, string); UTC time-log window bounds in `YYYY-MM-DD`.
- `logs_batch_size` (optional, string); window width in days, default `30`, maximum `365`.

Secret fields are redacted in logs and write previews: `password`.

The runtime uses only the fixed My Hours login and API origins. It does not accept caller-provided origins, credential controls, fixture modes, refresh, or replay.

## Streams notes

- `clients`: GET `/Clients`.
- `projects`: GET `/Projects/getAll`.
- `users`: GET `/Users/getAll`.
- `tags`: GET `/Tags`.
- `time_logs`: GET `/Reports/activity` in contiguous, non-overlapping UTC `DateFrom`/`DateTo` windows, capped at 600 windows.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint groups.
