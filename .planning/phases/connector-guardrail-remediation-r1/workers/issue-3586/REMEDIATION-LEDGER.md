# Remediation Ledger — issue 3586

## Scope and source evidence

- Sub-issue: #3586 — generated/unrelated connector remediation.
- Parent: #3579; parent PR #3580.
- Source audit report: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-eight-connector-guardrail-audit-r1/report.md` sections for PR #3532 and PR #3535.
- PR #3532: https://github.com/polymetrics-ai/cli/pull/3532; merge SHA `e99d6f1193814d00bd1b0c09fc092639d4fd8c54`; target connector `zendesk-support`.
- PR #3535: https://github.com/polymetrics-ai/cli/pull/3535; merge SHA `5d61794f76c46ca256280eb4166c5285c4b68731`; target connector `google-ads`.

## Remediation summary

- Forward-only remediation: no history rewrite, no force-push, no blanket revert.
- No audited docs/manual/website/generated output was removed in this slice. Evidence showed current state is either valid target output, a valid shared generated index/golden, or valid unrelated connector safety/generated metadata that would fail current validation if reverted.
- The path ownership defect is remediated by recording these dispositions plus the guard fixture/proof in this worker directory. Executable target-aware guard enforcement is owned by #3581; this worker does not edit shared `cmd/connectorgen/**` or `internal/connectors/boundary/**` code.

## Validation evidence used for dispositions

```bash
for p in internal/connectors/defs/zendesk-support internal/connectors/defs/google-ads internal/connectors/defs/gong; do
  go run ./cmd/connectorgen validate "$p"
done
```

Result captured locally: all three connector definitions checked with `0 findings`.

Additional Gong revert probe, without editing the worktree:

```bash
tmp=$(mktemp -d); cp -R internal/connectors/defs/gong "$tmp/gong"
git show 5d61794f76c46ca256280eb4166c5285c4b68731^:internal/connectors/defs/gong/cli_surface.json > "$tmp/gong/cli_surface.json"
go run ./cmd/connectorgen validate "$tmp/gong"
```

Result: failed with 19 `cli_surface_safety` findings for missing required CLI flags. Disposition: preserve current Gong `required:true` safety metadata even though #3535 should not have carried unrelated Gong churn.

## PR #3532 — Zendesk Support dispositions

| Path | Path class | Disposition | Evidence |
| --- | --- | --- | --- |
| `docs/connectors/bahmni/MANUAL.md` | unrelated connector manual | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds `operation=bahmni.patient_search`; valid operation-mapping docs, not stale. |
| `docs/connectors/bahmni/SKILL.md` | unrelated connector skill | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds `operation=bahmni.patient_search`; valid operation-mapping docs, not stale. |
| `docs/connectors/bitbucket/MANUAL.md` | unrelated connector manual | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds planned operation IDs for Bitbucket search/download commands; valid operation metadata, not stale. |
| `docs/connectors/bitbucket/SKILL.md` | unrelated connector skill | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds planned operation IDs for Bitbucket search/download commands; valid operation metadata, not stale. |
| `docs/connectors/gong/MANUAL.md` | unrelated connector manual | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds operation IDs for Gong direct-read commands; current Gong definition validates. |
| `docs/connectors/gong/SKILL.md` | unrelated connector skill | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds operation IDs for Gong direct-read commands; current Gong definition validates. |
| `docs/connectors/hubspot/MANUAL.md` | unrelated connector manual | Preserve current generated docs; mark as lane violation for future target-aware guard. | Audit classified unrelated to Zendesk; no stale output evidence found. Removing would desync generated docs from current CLI-surface metadata. |
| `docs/connectors/hubspot/SKILL.md` | unrelated connector skill | Preserve current generated docs; mark as lane violation for future target-aware guard. | Audit classified unrelated to Zendesk; no stale output evidence found. Removing would desync generated docs from current CLI-surface metadata. |
| `docs/connectors/xero/MANUAL.md` | unrelated connector manual | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds operation IDs for Xero report/attachment commands; valid operation-mapping docs, not stale. |
| `docs/connectors/xero/SKILL.md` | unrelated connector skill | Preserve current generated docs; mark as lane violation for future target-aware guard. | Diff adds operation IDs for Xero report/attachment commands; valid operation-mapping docs, not stale. |
| `docs/connectors/catalog/all-connectors.json` | shared generated index | Preserve as valid shared index; future guard should allow only narrow generated index updates when scoped/generated evidence exists. | Audit says generated output; target Zendesk docs/metadata also changed in same PR. |
| `docs/connectors/README.md` | shared generated index | Preserve as valid generated docs index; future guard should allow only narrow generated index updates when scoped/generated evidence exists. | Audit says generated output; no stale output evidence. |
| `docs/connectors/zendesk-support/MANUAL.md` | target generated manual | Preserve. | Target connector generated manual for #3532; `connectorgen validate internal/connectors/defs/zendesk-support` passes. |
| `docs/connectors/zendesk-support/SKILL.md` | target generated skill | Preserve. | Target connector generated skill for #3532; `connectorgen validate internal/connectors/defs/zendesk-support` passes. |
| `internal/cli/testdata/golden_transcripts.json` | shared generated golden | Preserve as valid generated/help golden; future guard should allow only narrow shared goldens with target evidence. | Audit says generated output; no stale output evidence. |
| `internal/connectors/guide.go` | shared generated/docs renderer output path | Preserve; not edited by this worker. | Audit says generated output; #3532 added operation rendering coverage in `internal/connectors/guide_test.go`. |
| `website/data/connectors.generated.json` | website generated data | Preserve as valid generated website data; future guard should allow only target/narrow generated website outputs. | Audit says generated output; target Zendesk operation metadata needed website data exposure. |
| `website/lib/connectors.catalog.data.generated.json` | website generated data | Preserve as valid generated website catalog data. | Audit says generated output; no stale output evidence. |
| `website/lib/connectors.types.ts` | website generated type artifact | Preserve as valid generated website type artifact. | Audit says generated output; no stale output evidence. |
| `website/app/api/raw/[...slug]/route.ts` | website shared source | Preserve as shared website foundation; future connector lane should reject or require separate foundation PR. | Diff adds operation mapping rendering for connector raw docs API. Not stale, but not target-owned. |
| `website/app/api/search/route.ts` | website shared source | Preserve as shared website foundation; future connector lane should reject or require separate foundation PR. | Diff indexes `command.operation` for search. Not stale, but not target-owned. |
| `website/app/docs/connectors/[slug]/page.tsx` | website shared source | Preserve as shared website foundation; future connector lane should reject or require separate foundation PR. | Diff renders `operation:<id>` mapping. Not stale, but not target-owned. |
| `website/scripts/lib/cli-surface.mjs` | website shared tooling | Preserve as shared website foundation; future connector lane should reject or require separate foundation PR. | Diff maps `command.operation` into generated website data. Not stale, but not target-owned. |
| `website/tests/api/connector-data.test.ts` | website shared test | Preserve as shared website foundation test; future connector lane should reject or require separate foundation PR. | Test asserts Zendesk operation-backed CLI surface metadata. Not stale, but not target-owned. |

## PR #3535 — Google Ads dispositions

| Path | Path class | Disposition | Evidence |
| --- | --- | --- | --- |
| `internal/connectors/defs/gong/cli_surface.json` | unrelated connector definition | Preserve current safety metadata; mark as lane violation for future target-aware guard. | Current Gong validation passes. Reverting to pre-#3535 file causes 19 `cli_surface_safety` findings for missing required body-mapped flags. |
| `docs/connectors/catalog/all-connectors.json` | shared generated index | Preserve as valid shared generated index; future guard should allow only narrow generated index updates with target evidence. | Audit says generated output; Google Ads target docs/metadata changed in same PR. |
| `docs/connectors/catalog/all-connectors.md` | shared generated index | Preserve as valid shared generated index. | Audit says generated output; no stale output evidence. |
| `docs/connectors/google-ads/MANUAL.md` | target generated manual | Preserve. | Target connector generated manual for #3535; `connectorgen validate internal/connectors/defs/google-ads` passes. |
| `docs/connectors/google-ads/SKILL.md` | target generated skill | Preserve. | Target connector generated skill for #3535; `connectorgen validate internal/connectors/defs/google-ads` passes. |
| `docs/connectors/README.md` | shared generated index | Preserve as valid generated docs index; future guard should allow only narrow generated index updates with target evidence. | Audit says generated output; no stale output evidence. |
| `internal/cli/testdata/golden_transcripts.json` | shared generated golden | Preserve as valid generated/help golden; future guard should allow only narrow shared goldens with target evidence. | Audit says generated output; no stale output evidence. |
| `internal/connectors/guide.go` | shared generated/docs renderer output path | Preserve; not edited by this worker. | Audit says generated output; no stale output evidence. |
| `website/data/connectors.generated.json` | website generated data | Preserve as valid generated website data; future guard should allow only target/narrow generated website outputs. | Audit says generated output; Google Ads target metadata validates. |
| `website/lib/connectors.catalog.data.generated.json` | website generated data | Preserve as valid generated website catalog data. | Audit says generated output; no stale output evidence. |
| `website/lib/connectors.catalog.generated.ts` | website generated data | Preserve as valid generated website catalog artifact. | Audit says generated output; no stale output evidence. |
| `website/lib/connectors.types.ts` | website generated type artifact | Preserve as valid generated website type artifact. | Audit says generated output; no stale output evidence. |
| `website/scripts/lib/cli-surface.mjs` | website shared tooling | Preserve as shared website foundation; future connector lane should reject or require separate foundation PR. | Diff maps `flag.required` to website data; needed by current CLI-surface safety metadata, not stale. |

## Out-of-scope audited shared paths

The audit also listed shared Go/runtime/tooling and migration guidance paths for #3532/#3535. This sub-issue does not edit or remediate those paths because #3584/#3585/#3583 own shared runtime/tooling and PM/guidance remediation. The guard fixture still marks shared website/tooling and unrelated connector paths as future rejections for connector lanes.
