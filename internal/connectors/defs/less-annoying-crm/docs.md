# Less Annoying CRM Connector

## Overview

Reads Less Annoying CRM users, contacts, tasks, notes, and events through fixed v2 RPC requests.

Readable streams: `users`, `contacts`, `tasks`, `notes`, `events`.

Service API documentation: https://www.lessannoyingcrm.com/help/topic/API.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Less Annoying CRM Programmer API key.
- `start_date` (required, string); Initial ETL lower-bound configuration retained for connection compatibility.

Authentication uses declared mode(s): `api_key_header`.

## Execution contract

Connection check: `POST /`
Check JSON body: `Function`=GetUsers.

## Streams notes

- `users`: `POST /`; records `.`
  - JSON body: `Function`=GetUsers.
- `contacts`: `POST /`; records `Results`
  - JSON body: `Function`=GetContacts.
- `tasks`: `POST /`; records `Results`
  - JSON body: `Function`=GetTasks.
  - Incremental cursor: `DateCreated`.
- `notes`: `POST /`; records `Results`
  - JSON body: `Function`=GetNotes.
- `events`: `POST /`; records `Results`
  - JSON body: `Function`=GetEvents.
  - Incremental cursor: `DateUpdated`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
