## Draft: Gong release-0.3.0 parity certification

Refs #2997

This PR is intentionally a draft and is **not merge-ready**. It preserves the Gong certification
branch and its source-mapping evidence while shared foundations and a disposable credential
reference remain external dependencies.

### Delivered evidence

- Re-audited Gong's public OpenAPI V2 source: 69 exact method/path/operation-ID rows match the
  Batch 2/3 immutable source lock.
- Reconciled the preserved branch with current main, typed destinations, and Batch 2/3 maps.
- Corrected Gong declaration-relative write paths and its CRM entity-schema fixture; the focused
  Gong generator, commandrunner, conformance, validation, and full surface-sync gates pass.
- Added the captain-required machine-readable missing-foundation ledger at
  `.planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/missing-foundation-gaps.json`.
  It has stable source-traced rows, source URL/revision/hash, runtime evidence, affected surfaces,
  closure commands, per-batch and portfolio rollups, and deduplicated shared-gap fan-out.

### Current open blockers

- `closed-operation-runtime-f4-binary-upload-approval-digest` is open on #4307. It blocks three
  exact Gong multipart operations (call media, CRM entities, and target assignments); each row is
  `enabled: false` and `merge_ready_eligible: false` in the ledger. No connector-specific bypass
  or unsafe fixture approval is present.
- The same ledger records the cross-connector F2/F4 fan-out: 51 source-traced Batch 2/3 operations
  remain non-enabled for merge-ready accounting (26 binary downloads; 25 binary/multipart uploads).
- No approved non-echoing disposable Gong credential reference exists. Live certification has not
  run, and no secret, customer payload, or browser-session substitute has been used.

### Verification run

- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/generate-missing-foundation-gaps.mjs --check`
- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs`
  → `verified 19 connectors / 5127 documented operations`
- `go run ./cmd/agentcontractgen check`
- `git diff --check`

The GSD/TDD plan, red/green ledger, verification checklist, and run state are under
`.planning/phases/issue-2997-gong-live-parity-r1/`. Required skills and the inline GSD fallback
are recorded there. No `no-mistakes` pipeline, merge, or release publication has been run.
