---
name: pm-file
description: File connector knowledge and safe action guide.
---

# pm-file

## Purpose

Reads local JSONL or CSV files as source streams.

## Icon

- asset: icons/pm-file.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/karthik-sivadas/polymetrics-cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: file

## Authentication

- No secret authentication is required for this connector.

## Configuration

- path (required): Local JSONL or CSV file path.
- stream: Optional stream name override.

## ETL Streams

- file: Local file stream from configured path.

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
- Source modes: full_refresh, incremental
- Destination modes: append, overwrite, append_dedup, overwrite_dedup

## Security

- read risk: local file read
- write risk: unsupported
- mutation risk: none
- approval: not required for reads
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect file
```

### Inspect as structured JSON

```bash
pm connectors inspect file --json
```

### File ETL

```bash
pm credentials add file-local --connector file --config path=/path/to/records.jsonl
pm connections create file_to_warehouse --source file:file-local --destination warehouse:warehouse-local --stream file --table imported_records
pm etl run --connection file_to_warehouse --stream file --json
```

## Agent Rules

- Run pm connectors inspect file before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
