# Context — PR #4308 status-check result preservation

## Decided scope

- Repair the common CLI output boundary for the already-complete `commandrunner.Result.StatusCheck` result.
- Keep the existing `BinaryDownload` and `DirectRead` branches byte-for-byte behaviorally unchanged.
- Use a dedicated, typed v1 envelope kind following the existing result-envelope naming convention; do not classify a status check as an ETL read.
- Test only shared engine/CLI code and source-locked synthetic declarations. No production connector bundle or command spelling is added.
- Re-run the harmless credential-free GitHub Pages HEAD/GET proof after the fix. The delivery evidence must omit credentials and scratch data.

## Scope fence

This is a shared foundation continuation on existing PR #4308, not a connector implementation lane: it changes no `internal/connectors/defs/**` declaration and contains no connector-name branch. Its only production responsibility is preserving a typed result the ordinary loader, runner, and provider executor already produce.
