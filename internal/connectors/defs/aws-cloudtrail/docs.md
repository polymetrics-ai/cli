# Overview

AWS CloudTrail connector parity is implemented from the official AWS CloudTrail API Reference Actions page fetched on 2026-07-31. The bundle enumerates 60 official CloudTrail API actions exactly once: 19 ETL/read streams, 10 bounded provider query/lookup direct reads, and 31 typed reverse-ETL write actions. The official CloudTrail event-record contents page documents event record version 1.11 and 31 top-level event fields; those fields are schema payload fields for LookupEvents, not CDC/changefeed operations.

Readable streams: `describe_trails`, `get_channel`, `get_dashboard`, `get_event_configuration`, `get_event_data_store`, `get_event_selectors`, `get_import`, `get_insight_selectors`, `get_resource_policy`, `get_trail`, `get_trail_status`, `list_channels`, `list_dashboards`, `list_event_data_stores`, `list_import_failures`, `list_imports`, `list_public_keys`, `list_tags`, `list_trails`.

Bounded direct/provider query commands: `query cancel`, `query describe`, `query generate`, `query results`, `insights data`, `insights metric-data`, `query list`, `events lookup`, `sample-queries search`, `query start`.

Typed write actions: `add_tags`, `create_channel`, `create_dashboard`, `create_event_data_store`, `create_trail`, `delete_channel`, `delete_dashboard`, `delete_event_data_store`, `delete_resource_policy`, `delete_trail`, `deregister_organization_delegated_admin`, `disable_federation`, `enable_federation`, `put_event_configuration`, `put_event_selectors`, `put_insight_selectors`, `put_resource_policy`, `register_organization_delegated_admin`, `remove_tags`, `restore_event_data_store`, `start_dashboard_refresh`, `start_event_data_store_ingestion`, `start_import`, `start_logging`, `stop_event_data_store_ingestion`, `stop_import`, `stop_logging`, `update_channel`, `update_dashboard`, `update_event_data_store`, `update_trail`.

## Auth setup

Use `pm credentials add <name> --connector aws-cloudtrail` and provide secrets only from environment variables or stdin:

- `aws_key_id` (required secret)
- `aws_secret_key` (required secret)
- `aws_region_name` (required config, for example `us-east-1`)

Optional config fields are `base_url` for local fixture endpoints, `page_size`, `max_pages`, `start_date`, and `mode=fixture` for credential-free tests. Do not place AWS secret values in chat, command history, docs, fixtures, or issue comments.

## Streams notes

Every stream uses a fixed AWS CloudTrail JSON-RPC action with SigV4 authentication and no raw action/path/header/body escape hatch. Paginated actions pass bounded `MaxResults` and follow `NextToken` until it is absent or `max_pages` is reached. Streams whose official request schema has required fields expose those fields through definition-owned command flags and fixture `read_query` values. The connector keeps `query=false` at the metadata layer; CloudTrail provider query operations are modeled as typed direct reads, not as warehouse SQL.

## Write actions & risks

Reverse ETL writes are declared in `writes.json` as closed top-level AWS request schemas. Runtime execution remains the standard `pm reverse` flow: plan -> preview -> explicit approval -> execute. Destructive/admin operations such as delete, stop, disable, resource-policy, delegated-admin, and logging controls declare destructive confirmation metadata. The native executor maps each write action to one fixed CloudTrail `X-Amz-Target`; operators cannot supply arbitrary AWS action names or raw request bodies. Delete actions are treated idempotently for provider missing-resource HTTP 404 responses in fixture and local replay.

## Known limits

- This work is fixture-only and local-test verified; it does not certify live AWS provider behavior.
- CloudTrail event record fields are parsed as payload/schema fields where CloudTrail returns them, but they are not counted as CDC because AWS does not document a CloudTrail changefeed/subscription API in the audited sources.
- Provider query direct reads are bounded JSON responses with redaction. They do not expose a generic CloudTrail Lake SQL runner; only the documented `StartQuery` `QueryStatement` field exists as a typed, fixed-target operation.
- Nested AWS structures are typed as closed top-level fields with object/array payloads because the official action pages link many nested reusable shapes; this bundle does not add raw generic sub-body passthrough beyond those documented fields.
