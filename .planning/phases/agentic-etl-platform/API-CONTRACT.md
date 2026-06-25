# API CONTRACT: Agentic ETL Platform

## CLI JSON Error

```json
{
  "api_version": "polymetrics.ai/v1",
  "kind": "Error",
  "error": {
    "category": "validation",
    "code": "validation_error",
    "message": "safe, sanitized message"
  }
}
```

## Connector Manifest

Connector inspection returns metadata plus manifest fields:

- `config_fields`
- `secret_fields`
- `streams`
- `pagination`
- `sync_modes`
- `risk`

## Skills Generation

`poly skills generate --dir <path>` writes skill directories containing `SKILL.md` and an index document.
