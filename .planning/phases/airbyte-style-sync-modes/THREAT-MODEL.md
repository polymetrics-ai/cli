# Threat Model

## Assets

- Credentials in the encrypted vault.
- Local warehouse raw and final JSONL files.
- Per-stream cursor state.
- Agent-facing JSON output.

## Risks

- Secret leakage through logs or generated docs.
- Destructive overwrite on failed sync.
- Incorrect checkpoint advancement causing data loss.
- Agent-supplied unsafe stream, table, or mode values.

## Mitigations

- Never resolve credentials for docs or skills.
- Use temp files and atomic rename for overwrite finalization.
- Commit stream state only after successful finalization.
- Validate sync modes, cursor, primary key, and destination names before ETL side effects.

