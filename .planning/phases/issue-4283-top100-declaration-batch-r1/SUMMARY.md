# Increment 001 summary

Ten ranked daily-use API connectors have reproducible public-source locks and
exact source-to-`api_surface` declaration records. This remains a
declaration-only, non-live-certification checkpoint.

- Public operations found and mapped: 4,378 / 4,378; every Batch-1 source is
  a complete provider-published OpenAPI document with high input confidence.
- Docker Hub runnable retrofit: 41 commands, 20 typed write actions, and six
  typed delete actions across 54 pinned operations.
- Docker Hub disabled dispositions: 13 (ten named foundation gaps and three
  schema/media incompatibilities; `unsafe-to-exercise` is zero).
- All ten connectors now declare the reusable ETL source transport. Reverse
  ETL eligibility is the recoverable `generic-typed-destination-executor` gap;
  no generic writer or destination binding was invented.
- No credential was requested or used; all live certification remains pending.

## Docker Hub corrected deliverable

Docker Hub is the first runnable-parity proof slice. Its 50 operation contracts
(23 `rest_read`, 27 `rest_write`) are backed by four existing ETL commands, 17
operation-bound direct reads, and 20 reverse-ETL commands linked to typed
writes. The reverse-ETL surface preserves plan, preview, approval, and execute;
it does not expose a generic HTTP mutation path.

The public OpenAPI is the only schema source. Required path-item parameters are
derived exactly from the pinned document. The `params-import` limitation that
does not merge those inherited parameters is recorded as a recoverable
foundation/generator gap rather than concealed by manual flag invention. The
token creates and the three credential exchanges remain source-declared
foundation gaps because the current engine cannot safely execute
`sensitive_policy` writes or return a secret response.

Verification evidence is recorded in `VERIFICATION.md`, source pin evidence in
`SOURCE-LOCK-VERIFICATION.json`, and exact operation rejections in
`REJECTION-LIST.json`. Complete the current local gates and commit this Docker
Hub slice before selecting another connector. Gitea is not a definition bundle
on `main` and must be created or substituted before a later increment.

## Classification correction

Every provider mutation is now classified as a direct-write endpoint. The
cohort has 2,370 direct-write endpoints and 118 enabled direct-write bindings;
the former reverse-ETL primary-class count is zero. Reverse ETL is an
eligibility attribute on those direct writes and stays zero-eligible until a
connector-neutral typed destination factory supplies the required transport
binding, apply strategies, and acknowledgement contract.

## Provider surface reconciliation

The full-provider-source audit found no Batch-1 understatements. The old/new
`api_surface.json` counts are unchanged (Docker Hub 54, Notion 55, Stripe 589,
Bitbucket 331, GitLab 1,755, CircleCI 111, Sentry 224, Vercel 422, Asana 249,
Jira 617), while their settled provider OpenAPI counts are 54, 49, 589, 331,
1,755, 111, 223, 400, 249 and 617. The 6/1/22 positive deltas are documented
bounded variants or legacy entries, not a source-completeness substitute.

## CI verify repair

The batch's source-only transport declarations exposed an order-dependent test
assumption in `TestDefinitionTransportFactoriesSelectDeclaredEvidence`: a
shared factory may hold a declaration's evidence as its primary record or an
additional accepted record. The test now asserts the actual accepted-evidence
contract, and both the targeted test and complete `internal/app` package pass
locally. This is test-only; it does not change transport composition,
connector declarations, credentials, or live-certification status.

## Reconciliation relaunch — first connector-local proof slice

The first bounded reconciliation slice declares four exact, fixture-backed
destination mappings—Notion views, Stripe customers, CircleCI schedules, and
Vercel projects—and supplies the missing CircleCI/Vercel ETL and reverse-ETL
commands. The installed binary reaches credential preflight for those four
commands without a provider request. The new seven-surface ledger gives all
491 typed actions a named destination eligibility disposition (four static
fixture mappings and 487 exact pending dependencies), while retaining the
source action-set hashes for anti-drift verification.

This is deliberately not connector completion: source-operation command
coverage remains well below the requested full-reachability bar for several
providers. The generic destination descriptor also maps by executor and stream,
not selected action, so `action-scoped-source-binding` remains a precise
foundation dependency for differing action field maps. Persisted App/CLI
destination dispatch is additionally pending in the #4304 foundation branch.
No provider credential was used; live certification remains pending.
