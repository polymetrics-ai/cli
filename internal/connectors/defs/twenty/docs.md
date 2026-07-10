# Twenty CRM connector

## Overview

S1 ships the Twenty CRM connector as a foundation skeleton. Metadata, connection spec,
API surface, streams skeleton, and this documentation load through the declarative engine,
but executable streams and writes are introduced by later slices.

## Auth setup

Twenty uses bearer authentication with an `api_key` secret. Provide the value from an
environment variable or stdin; do not paste it into prompts, commit it, or print it in logs.

## Streams notes

No streams are declared in S1. Read streams for researched GET endpoints are added by S3
#280 alongside their `api_surface.json` coverage entries.

## Write actions & risks

No write actions are declared in S1. Reverse-ETL writes for POST/PATCH endpoints land in S4
#281, and destructive DELETE coverage lands in S5 #282. Reverse ETL remains plan, preview,
approval, then execute.

## Known limits

This skeleton intentionally has an empty `api_surface.json` endpoint list so conformance stays
bidirectional: every covered endpoint must resolve to a declared stream or write, and every
declared stream or write must be covered. The researched 168-operation manifest remains in
`.planning/auto-loop/RESEARCH/twenty/RESEARCH.json` until later slices materialize it
incrementally with the matching streams and writes.
