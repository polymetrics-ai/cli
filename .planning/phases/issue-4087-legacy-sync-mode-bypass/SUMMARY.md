# Summary: issue #4087 legacy sync-mode bypass

The two public deduped compatibility aliases now parse to their closed typed contracts through both normal and persisted-legacy routes:

- `full_refresh_overwrite_deduped` → `full_overwrite`
- `incremental_append_deduped` → `incremental_dedupe`

Without an admitted transport, `RunETL` returns `ModeNotExecutableError` before source I/O. The compatibility names remain public, and generic help, generated documentation, website docs, and connector certification report the same result.

`internal/synccontract/public_modes.go` is the connector-neutral authority for their public names, typed contracts, and generic capability projection.

The change is connector-neutral. Required GSD/TDD, verification, and inline manual code-review evidence are recorded in this phase directory; PR creation and external automated review remain Firstmate-owned next steps.
