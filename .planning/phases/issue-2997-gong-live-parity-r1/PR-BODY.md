## Draft: Gong release-0.3.0 parity certification

Refs #2997

This PR is intentionally a draft and is **not merge-ready**. It preserves the Gong certification
branch and its source-mapping evidence while the remaining generic source-import and full live
parity dependencies are resolved.

### Delivered evidence

- Re-audited Gong's public OpenAPI V2 source: a current 453,797-byte pinned artifact has 69 exact
  method/path/operation-ID/deprecation rows and matches the committed semantic fingerprint.
- After #4335 merged, converted that unchanged ledger into the importer’s v3 `gong-v2` document
  contract with the exact fixed `?version=` query declared as identity-bearing. The real importer
  now traverses the query and parses the official document, proving the source-query foundation is
  wired without a Gong-specific bypass.
- Merged `origin/main` tree `6410fe59c` (the final squashed shared-foundation rollup) without
  discarding the preserved branch.
- Promoted all 27 named write actions to implemented declaration-owned reverse-ETL commands;
  corrected array/object flag types and the required CRM schema flag. The three multipart uploads
  pass focused generic conformance with approval-digest binding.
- Made the required output policy explicit: Gong ordinary provider values—including values that
  happen to match configured credentials—must be retained. Only explicitly declared secret fields
  may be masked with an explicit marker. A strengthened focused declaration test covers all 30
  direct reads; the shared runtime correction is tracked in #4321.
- Regenerated manuals, skills, website data, and the Batch 2/3 map. All 69 source rows are enabled;
  the missing-foundation ledger has zero Gong rows.
- In a fresh initialized project with no credentials, the built binary ran every implemented Gong
  path—30 direct reads, 27 reverse-ETL writes, and 12 ETL streams—to `missing --credential`, with
  zero unknown, partial, or unbound results and no provider I/O.
- Reconciled three multipart parameter declarations using `connectorgen params-import`; its
  immediate check reports 17 scanned with zero remaining drift. Gong certification candidates and
  the generated 71-row sweep are current.
- Live, non-echoing persisted-App evidence now covers authentication, one bounded ETL read, one
  bounded typed direct read, required-input validation, and cursor-pagination validation. The
  repository `--direct-read-only --external-proof` harness report passed with zero leaks; only a
  non-secret proof fingerprint is retained in the phase ledger.
- The pending clean-run plan now uses the eight-surface contract: ETL, direct read, direct write,
  reverse ETL, binary download, binary upload, flow, and schedule. Gong's three named multipart
  actions map binary upload; the standard bounded `flow_roundtrip` and `schedule_roundtrip` map
  the two application workflows without inventing a Gong scheduler or flow-runtime endpoint.

### Current open blockers

- The #4335 source-query foundation is merged at `8127de418`, but source-import now stops at its
  next provider-neutral boundary: `GET /v2/all-permission-profiles` parameter 0 is required query
  `workspaceId`, whose official schema is `type: string` with no `maxLength`. The importer rejects
  it before it can write the canonical descriptor. The required generic behavior is to retain or
  type-gap that provider contract; no Gong max length or shared-code exception has been added.
- Foundation #4337 must prevent account-scoped non-secret configuration values from entering an
  external-proof process command. The current serializer retains that command verbatim, so the
  planned tenant endpoint cannot be passed to a proof-producing run until the provider-neutral
  privacy boundary is fixed. No account identifier is retained in this PR or its evidence.
- `GET /v2/targets` requires `workspaceId`, but the current direct CLI declaration has not yet
  received that required marker because `surface-sync` is blocked by the missing canonical source
  descriptor above. The live no-input call is recorded only as HTTP-400 classification; no manual
  flag or provider-specific runtime path was added.
- Shared provider-result projections currently collision-mask an ordinary scalar, header, raw
  body, or cursor when it equals configured credential material. This violates the captain's
  provider-output rule. Foundation issue #4321 owns the provider-neutral fix; this PR contains no
  Gong-specific escape hatch.
- Full live parity remains incomplete: the full harness observed 16 bounded ETL records and seven
  passing ETL append cells, while `calls`, `library_folders`, `flows`, `flow_folders`, and
  `permission_profiles` remain uncertified. There are no declaration-owned, self-cleaning Gong
  write/readback/cleanup pairings, so no mutation was sent. `get-brief` and `ask-entity` were not
  called because they are paid agentic endpoints and require a captain decision for any future
  certification cell.

### Verification run

- Focused Gong surface, commandrunner preflight, and multipart conformance tests with `-timeout 20m`.
- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs`
  → `verified 19 connectors / 5127 documented operations`.
- Built-binary 69-command credential-free sweep; docs validation; website generated-artifact tests;
  connector boundary/canon and repository workflow gates are recorded in the phase verification
  checklist as they complete.
- `git diff --check`.

The GSD/TDD plan, red/green ledger, verification checklist, and run state are under
`.planning/phases/issue-2997-gong-live-parity-r1/`. Required skills and the inline GSD fallback
are recorded there. No `no-mistakes` pipeline, merge, or release publication has been run.
