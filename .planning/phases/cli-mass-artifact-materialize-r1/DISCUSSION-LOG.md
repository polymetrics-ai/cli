# Discussion record — 426 connector artifact sweep

The captain order `CAPTAIN-ORDER-fast-426-terra-20260809.md` fixes the scope: reconcile and materialize the complete 426-target JSON surface, with static gates only. No unresolved product decision remains. This is an inline/manual `discuss-phase` fallback because the single-worker delivery contract forbids role spawning and this runner has no compatible Pi worker.

## Promoted-native command-surface recovery (2026-08-10)

The captain directed an immediate repair after the built binary returned `unknown command` (exit 2) for `apify-dataset`, `basecamp`, `copper`, `google-classroom`, `google-pagespeed-insights`, `metabase`, and `rootly`. No product choice is open: their materialized `cli_surface.json` files already declare 25 implemented ETL commands, and each has a Tier-2 hook adapter that delegates read execution to its legacy native connector.

Read-only diagnosis found that `bundleregistry.New` loads the hook-backed bundle and then `nativeset.RegisterInto` replaces it with a promoted `definitionConnector`. That wrapper forwards the bundle definition and configuration constraints, but not its `CommandSurface`; consequently the hand-rolled dynamic CLI rejects the connector before it can render help or invoke the native read path. The minimal repair is to forward the already-declared bundle command surface from `definitionConnector`, with a CLI-level regression covering exactly the seven affected connectors. It performs help-only, credential-free checks; no provider request, command contract, docs wording, or website data changes are needed.

## PR #3957 CI remediation decision (2026-08-10)

The captain directed four ordered, non-product decisions: make `govulncheck ./...` actually load
and scan by separating the two task-scoped script mains; replace the batch reference dangerous
scheme blocklist with an HTTP(S) allowlist that retains relative references; repair the
connector-boundary scanner rather than renaming `.rss`/`/rss.xml` static-asset entries; and make
the website generated-catalog search test assert the intended connector is present rather than
depending on a stale rank bound. The `data:`/`vbscript:` issue is defense in depth today because
the same-host and HTTPS gates reject it downstream, but the early filter must still be correctly
closed. The parked no-mistakes gate is explicitly out of scope for this remediation.
