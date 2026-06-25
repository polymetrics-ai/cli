# DATA MODEL: Agentic ETL Platform

## Run Progress

ETL runs store:

- run id
- connection
- stream
- status
- records read/transformed/loaded/failed
- batch count
- checkpoint cursor/page metadata
- timestamps
- sanitized error

## Connector Manifest

Connector manifests are static Go structs exposed by connector implementations and serialized for docs, skills, and agent inspection.

## Secrets

Secret values remain only in the encrypted vault and in process memory while resolving a credential. State, docs, skills, logs, and JSON output contain only secret field names.
