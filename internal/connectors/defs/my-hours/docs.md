# My Hours Connector

## Overview

Reads My Hours clients, projects, team members, tags, and bounded time-log windows through fixed REST routes.

Readable streams: `clients`, `projects`, `users`, `tags`, `time_logs`.

Service API documentation: https://developers.myhours.com/.

## Auth setup

Connection fields:

- `email` (required, string); My Hours login email.
- `end_date` (required, string); UTC date window terminal date, YYYY-MM-DD.
- `logs_batch_size` (optional, string); Declared time-log window width in days, 1 through 365.
- `password` (required, secret, string); My Hours login password.
- `start_date` (required, string); UTC date window start, YYYY-MM-DD.

Authentication uses declared mode(s): `declared_password_token`.

## Execution contract

Default stream pagination: `none`.

Connection check: `GET /Clients`

## Streams notes

- `clients`: `GET /Clients`; records ``
- `projects`: `GET /Projects/getAll`; records ``
- `users`: `GET /Users/getAll`; records ``
- `tags`: `GET /Tags`; records ``
- `time_logs`: `GET /Reports/activity`; records ``

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
