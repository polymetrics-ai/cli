# Inline code review — seven connector extraction r1

## Method

The canonical delivery contract forbids spawning review roles for this single-worker captain task,
so review was performed inline after the required `golang-lint` guidance and local lint gate. The
reviewed range is `origin/main...HEAD` plus the staged seven-bundle/generated-artifact delta.

## Reviewed risks and disposition

| Area | Review result |
| --- | --- |
| Shared `covered_by.writes` foundation | Additive and backward-compatible: singular `write` remains accepted; validator checks every plural target; API schema stays closed to undeclared fields. |
| Existing connector behavior | `connectorgen validate` loaded all 551 connectors with 0 findings; focused validator tests cover accepted and unknown plural targets. |
| Unrelated engine regression | Current-main REST operation parameter support was retained; only the bounded plural coverage capability was imported. |
| Bundle provenance and exclusions | Seven source trees exactly match `c28bc75a3`; no `github` or `zendesk-support` bundle/test input is present. |
| Generated artifacts | Surface sync, docs, website data, and golden transcript use documented generators; endpoint-ledger delta is within the seven. |
| Executability | The real compiled binary reached each of the 1,984 implemented command paths and their own `NAME` headers. |
| Credential/live-service safety | No credentials were requested, held, or used; no live connector operation ran. |

## Findings

The table above records the review as it stood at first pass. It was wrong on three rows, and those
rows are corrected here rather than deleted, because the way they read clean is the useful evidence.

| First-pass claim | What review actually found |
| --- | --- |
| "Shared `covered_by.writes` foundation … validator checks every plural target" | True of the validator, and of nothing else. Three further consumers — `conformance.checkSurfaceComplete`, `certify.hasSurfaceCoverage`/`addSurfaceCoverageCounts`, and `connectorgen`'s `batchSurfaceSplit`/`ensureMaterializedCoverage`/`materializeCLISurface` — still read the singular `.Write`, so all 292 jira and 252 workday-rest plural-only rows were invisible to them. `TestConformance` sweeps the real `../defs` tree and would have failed on jira and workday-rest; that package was not in the verified set. |
| "`connectorgen validate` loaded all 551 connectors with 0 findings" | Accurate and insufficient. `engine.ResolveCheck` scans `{{ }}` matches only, so twelve help-scout stream paths carrying a bare `{conversationId}`-style placeholder validated clean while the engine would send those braces to the wire. A zero-finding validate run is only as strong as the rules the validator has. |
| "Seven source trees exactly match `c28bc75a3`" | True at import, and no longer true of the shipped tree. Review found defects in the imported definitions, so 17 files were deliberately corrected afterwards — see TDD-LEDGER.md. |

Additionally, help-scout, jira and workday-rest each shipped `capabilities.write: true` alongside a
risk block and `docs.md` still declaring the connector read-only with no write actions. help-scout's
read "none; read-only, no obviously-safe reverse-ETL writes" for a connector with 65 write actions
including 18 permanent DELETEs, and that text was already published through `pm connectors inspect`,
the generated MANUAL/SKILL files and the website bundle.

Every finding above is fixed, regression-tested, and re-verified on the post-fix tree; the
enforcement now lives at a shared boundary so the same class cannot be reintroduced silently.

One item is deliberately left open and is called out in VERIFICATION.md rather than closed: the
1,984-command real-binary NAME sweep was not re-run after the bundle corrections, on the reasoning
that none of them adds, removes or renames a command.

Automated GitHub review is intentionally deferred until firstmate opens the PR; then the PR must
receive the repository's normal Claude review coverage or its documented fallback before human merge.
