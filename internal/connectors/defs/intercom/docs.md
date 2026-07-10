# Intercom connector

## Overview

Intercom is implemented as a declarative connector bundle backed by the official Intercom REST API 2.14 OpenAPI description. The bundle accounts for all 149 official operations with executable stream reads, bounded direct reads, bounded text/binary metadata reads, or approval-gated reverse ETL write actions.

## Auth setup

Store the Intercom token as the secret field `access_token` using `pm credentials add intercom-local --connector intercom --from-env access_token=<env-var-name>`. Optional non-secret config includes `base_url`, `api_version`, and `page_size`.

## Streams notes

Run `pm intercom` to inspect the provider-style command surface. Collection/list/search commands use declared ETL streams with bounded `--limit` handling. Direct object reads use fixed Intercom endpoints and the `json_response`, `text_response`, or `binary_metadata` output policies.

## Write actions & risks

Reverse ETL write commands create plans first. Live Intercom mutations require plan, preview, approval token, and execute. Destructive/admin actions declare `confirm: destructive`, so approved execution must include the typed confirmation challenge.

## Known limits

No credentialed live Intercom checks are run by static validation. Binary/file endpoints return metadata or bounded text rather than writing arbitrary files. The connector does not expose raw generic HTTP, SQL, shell, or arbitrary GraphQL write surfaces.
