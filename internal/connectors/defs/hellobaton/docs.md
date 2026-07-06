# Overview

Reads Hellobaton projects, milestones, tasks, phases, companies, and users through the Hellobaton
REST API.

Readable streams: `projects`, `milestones`, `tasks`, `phases`, `companies`, `users`.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Hellobaton API key, sent as the api_key query parameter on
  every request. Never logged.
- `base_url` (required, string); format `uri`; Hellobaton API base URL, e.g.
  https://<company>.hellobaton.com/api.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `page_size`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `projects`: GET `/projects` - records path `results`; query `page_size`=`100`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `milestones`: GET `/milestones` - records path `results`; query `page_size`=`100`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `tasks`: GET `/tasks` - records path `results`; query `page_size`=`100`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host.
- `phases`: GET `/phases` - records path `results`; query `page_size`=`100`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host.
- `companies`: GET `/companies` - records path `results`; query `page_size`=`100`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `users`: GET `/users` - records path `results`; query `page_size`=`100`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Hellobaton API read of project, milestone,
task, phase, company, and user data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
