# Verification — GitHub mutation certification slice 4 writes-d

## Pending live run

- [ ] The original Batch 1 count is superseded and must be recomputed before PR handoff: it incorrectly retained `no_object=19` without fixture recovery and did not include subsequently retained certifications.
- [x] Certified evidence currently retained: users add social account, create public SSH key, create SSH signing key, archive repository (direct restore/read-back; contained state), create private disposable repository, create repository ruleset, create private gist, and create label. Each executed mutation had an independent produced-value read-back; cleanup is recorded separately as containment proof and does not alter certification.
- [x] Surface finding resolved by fleet ruling: connector commands use their authorized local `--approve` token path; stdin-token support is transport-only. See `TDD-LEDGER.md`.
- [x] Live continuation covered paths 51–145 one command at a time. Safe fixture writes produced direct provider cleanup/read-back evidence for ruleset creation, private gist creation, and label creation. Numeric path-ID attempts from paths 98 onward reproduced the fleet-wide `integer_id_scientific_notation` malformed-path class; the raw connector URLs contain scientific notation, so these were not classified as provider `no_object` results. The final six retry plans were fenced by the connector authentication cohort before any provider request and are retained as honest product-defect outcomes pending the next batch's evidence tally.
- [x] `no_object` recovery audit started: the branch parent collection was independently enumerated (`main`), then two disposable `pm-cert-` branches were created. Connector branch mutations fenced before GitHub; a raw GitHub branch-protection PUT and independent GET both returned 200, followed by direct protection DELETE (204), direct branch DELETE (204), and branch GET (404). These branch-path outcomes are therefore connector `product_defect`, not `no_object`; the original batch tally must be recomputed before PR handoff.
- [x] Label-delete recovery confirms the same rule: a real `pm-cert-` label fixture was created (201), but the connector fenced before the request; direct provider deletion and 404 read-back contained the fixture. This is a connector product defect, not `no_object` and not a failed-cleanup downgrade.
- [ ] Each manifest command has exactly one terminal bucket: `certified`, `no_object`, `wrong_credential`, `entitlement`, `not_implemented`, `product_defect`, or `escape_needs_captain`.
- [ ] Every `certified` record contains an observable effect assertion and an independently proven cleanup absence assertion.
- [ ] No credential value is present in the worktree, evidence, command output, PR body, or status line.
- [ ] `go run ./cmd/connectorgen certification-matrix --check` is run after each eligible evidence record and at handoff.
- [ ] Repository verification and `scripts/verify-gsd-workflow` are run before the PR.
- [ ] The opened PR base is read from GitHub's API and equals `integration/4015-mvp-flat-r1`.
