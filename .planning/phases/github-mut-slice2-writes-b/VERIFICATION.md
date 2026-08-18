## Verification Checklist

- [ ] All 146 assigned paths attempted serially and classified exactly once. (Current: 1/146 attempted; `issue create` completed plan → preview → stdin-token run.)
- [ ] Every certified result has plan, preview, token-stdin execution, independent produced-value read-back, direct provider DELETE, and independent absence proof.
- [ ] Every product defect includes a raw GitHub API control.
- [ ] Every retained schema-v2 record passes `go run ./cmd/connectorgen certification-matrix --check`. (Current validator run: passed; evidence is written directly from captured real traffic under the fleet ruling.)
- [ ] `git diff --check` passes.
- [ ] Targeted Go tests pass.
- [ ] PR is opened against `integration/4015-mvp-flat-r1`, and its API-reported base is recorded.

## Live attempt receipt — 2026-08-18

`issue create` used the contained fixture title `pm-cert-mut2-issue-create-20260818-0320`.

- Produced-value assertion (`agent_derived`): the independent issue collection contained exactly one issue with that unique title; a missing title or a different title is rejected.
- The command completed the connector-command plan → preview → bare `--approval-token-stdin` run lifecycle.
- The mandated REST provider DELETE returned HTTP 404 while the independent item read-back returned HTTP 200: this reproduces the stated cleanup defect.
- A direct GitHub GraphQL `deleteIssue` cleanup mutation then succeeded; subsequent item read-back returned HTTP 410 and collection read-back contained zero title matches. The cleanup is therefore proven absent, but it did not meet the task's literal REST-DELETE requirement.
- The importer limitation is a surface finding only. The fleet ruling authorizes direct schema-v2 evidence authoring from the captured real command and provider traffic; it never authorizes invented exchanges.
