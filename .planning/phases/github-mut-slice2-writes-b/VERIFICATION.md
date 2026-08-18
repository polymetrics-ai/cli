## Verification Checklist

- [ ] All 146 assigned paths attempted serially and classified exactly once. (Current: commands 1–141 have been attempted; command 142 is at the real-money captain gate, and 143–146 remain behind it. Final classifications remain under the mandated fixture/read-back re-audit.)
- [ ] Every certified result has plan, preview, token-stdin execution, independent produced-value read-back, direct provider DELETE, and independent absence proof.
- [ ] Every product defect includes a raw GitHub API control.
- [ ] Every retained schema-v2 record passes `go run ./cmd/connectorgen certification-matrix --check`. (Current validator run: passed; 17 records are added over the integration base and each comes from captured live traffic.)
- [ ] `git diff --check` passes.
- [ ] Targeted Go tests pass.
- [ ] PR is opened against `integration/4015-mvp-flat-r1`, and its API-reported base is recorded.

## Batch 1 checkpoint — 2026-08-18

The per-command receipt and corrected bucket split are in `BATCH-1.md`. The empty-collection audit converted commands 11, 19, and 20 from `no_object` to certified by creating contained fixtures and proving both the mutation and cleanup. Command 58 was subsequently attempted against a nonexistent enterprise scope under supervisor direction and classified as entitlement.

## Batches 2 and 3 checkpoint — 2026-08-18

`BATCH-2.md` records the completed 50-command execution batch and names every row that still needs a real fixture before final classification. `BATCH-3.md` records commands 109–141, including raw provider controls for the migration and webhook-config defects. No pending row is banked cheaply as `no_object` or as a control-free `product_defect`.

## Live attempt receipt — 2026-08-18

`issue create` used the contained fixture title `pm-cert-mut2-issue-create-20260818-0320`.

- Produced-value assertion (`agent_derived`): the independent issue collection contained exactly one issue with that unique title; a missing title or a different title is rejected.
- The command completed the connector-command plan → preview → bare `--approval-token-stdin` run lifecycle.
- The mandated REST provider DELETE returned HTTP 404 while the independent item read-back returned HTTP 200: this reproduces the stated cleanup defect.
- A direct GitHub GraphQL `deleteIssue` cleanup mutation then succeeded; subsequent item read-back returned HTTP 410 and collection read-back contained zero title matches. The cleanup is therefore proven absent, but it did not meet the task's literal REST-DELETE requirement.
- The importer limitation is a surface finding only. The fleet ruling authorizes direct schema-v2 evidence authoring from the captured real command and provider traffic; it never authorizes invented exchanges.
