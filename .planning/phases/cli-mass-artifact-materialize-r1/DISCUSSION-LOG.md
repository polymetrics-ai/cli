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

## Certify-timing budget decision (2026-08-10)

The captain approved raising the certify cap only after the requested main-versus-branch
measurement: the endpoint-ledger corpus grew from 2,060 to 13,534 entries (6.57x), while actual
certify test elapsed time grew from 159.268s to 295.117s (1.85x). This is strongly sub-linear;
raising the budget does not conceal a performance regression. The approved bound is 8m rather
than 3m30s, and the exact figures plus rationale must be included in PR #3957's body. No product
or connector-runtime decision is open.

## Existing command-surface preservation and redaction decision (2026-08-10)

The captain confirmed that Amazon SQS's underscored generated reverse-ETL flags and Google Search
Console's missing command redactions are real regressions and that this slice must not damage an
already working connector. The implementation choice is therefore fixed: an existing reverse-ETL
command is retained as its complete executable public contract while materialization refreshes only
its cited provider endpoints; action `redact_fields` are then unioned into the retained command
surface. This covers flag spelling, optional fields, examples, approval metadata, and secrets
rather than addressing only the two visible artifacts. A read-only corpus audit must report any
remaining spelling or redaction drift before merge; it does not authorize mutating those other
connectors in this narrow repair.

## Full audited-corpus regeneration decision (2026-08-10)

The captain superseded the narrow-repair boundary after the audit established six further
flag-contract regressions and six further action-to-command redaction gaps. The generator is now
the only authorized repair point: regenerate every affected command surface from its preserved
pre-sweep command contract where one exists, refresh provider references from the current bundle,
and union every declared write-action redaction. Do not hand-edit the fourteen observed artifact
instances or accept any remaining audit finding. Completion requires both whole-corpus counts to
be exactly zero: changed established flag spellings and action redactions absent from their
matching command surfaces.

## Complete existing command-contract preservation (2026-08-10)

The post-regeneration comparison against the accepted `f96a47e80` baseline exposed a third
instance of the same materializer loss: eleven existing connector surfaces had their complete
`global_flags` contract removed. These flags include credential, output, pagination, and
reverse-ETL approval controls, so leaving them absent would violate the standing no-regression
constraint even though the requested two audits were already clean. The decision is fixed: use
the repaired generator with the preserved pre-sweep command surface for all eleven, then replace
only its generated output. Completion requires three zero-count audits: flag spelling,
write-action redaction propagation, and baseline global-flag contract equality.
