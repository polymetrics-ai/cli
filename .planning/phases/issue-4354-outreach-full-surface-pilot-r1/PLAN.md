# Plan — #4354 Outreach full-surface pilot

## GSD plan (`plan-phase --tdd`, inline fallback)

1. **Evidence intake (no production code):** compare current `origin/main` and the candidate; mechanically import only final Outreach JSON/docs/source artifacts. Record their exact candidate source and a target-owned operation/lane summary. Verify no non-Outreach connector artifacts are imported.
2. **Red:** run the narrow source-evidence and preflight tests/checks against the imported artifacts. Capture any schema-v3 reader failure exactly; do not edit the shared reader. Add an Outreach-owned test only if it can demonstrate an unmet connector-specific contract without changing shared code.
3. **Green:** make the smallest Outreach-only artifact or fixture correction required by the observed failure; retain all operation rows and include exact citations. Regenerate only the connector artifacts mandated by existing generators.
4. **Usability proof:** build `pm` and in an isolated project use fixture transport with no credential to prove representative ETL, direct-write/reverse-ETL, and destructive/delete commands stop at `missing --credential`, not a mapping/certification/hash/live-cert policy. Test a source identity/method/path mismatch at the applicable preflight boundary.
5. **Verification and review:** run focused package tests plus relevant independent generator/evidence/certification/CLI checks; record CLI docs/help/website parity as pass, updated, or blocked. Run `git diff --check`, independent clean-worktree rerun, review the diff for scope/security, and open a draft PR when status is truthful.

## Commit checkpoints

1. Planning/GSD checkpoint (this artifact set).
2. Imported-artifact plus failing-checkpoint evidence, if a useful red state occurs.
3. Green connector/evidence slice.
4. Verification/review fixes, if any.

## CLI help/manual/website parity

- [x] `pm connectors` bare namespace remains successful and contextual (exit 0).
- [x] `pm help outreach`, `pm outreach prospects list --help`, `pm outreach create account note apply --help`, and `pm outreach delete account apply --help` render the imported surface and deletion confirmation.
- [x] `docs/cli/**` has no Outreach-specific page; `website/data/connectors.generated.json` has an Outreach record but `cliSurface: null`.
- [x] Website parity is **blocked**, not waived: generation of the global catalog must wait for source admission and the generator/evidence owner. The binary manual/help is current because it embeds the imported bundle.
