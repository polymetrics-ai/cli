# Run state — issue #4291

- Status: paused on PR #4296, targeting `main` (API verified)
- Scope: only the 20 issue-listed connector definition directories plus this issue's planning evidence
- GSD fallback: no roadmap phase exists for #4291 and the canonical worker prohibits role spawning; prompts were generated and their required process was executed manually.
- Safety: public documentation only, no credentials, no live provider operation calls, no engine edits, and no destination action fabrication.
- Current citation reconciliation: `origin/main` at `3c66c33b8` is merged locally and #4332's
  discriminated rendered-reference contract is present. Twenty source locks were copied and
  byte-compared before the merge because main deletes their paths, then restored. Iterable's 29
  duplicate route identities were corrected and operation-evidence passes at 5,457 rows. A v3 trial
  for Close, Salesloft, Copper, Zoho Bigin, and Klaviyo reaches the descriptor boundary, where
  Close's hash-pinned public-spec import finds a valid provider number parameter without a declared
  `minimum`, rejected by `validateBoundedRequestSchemaType` / `sourceValidateNumericBounds` in
  `cmd/connectorgen/sourceimport.go:6989-6991,7277-7281`. Separately, valid v3 locks are invisible
  to `readOperationEvidenceSourceLock` (`cmd/connectorgen/operationevidence.go:483-498`), which
  only reads legacy `rest.operations` despite accepting schema v3. The trial locks were restored
  from the byte-compared preservation copy while both shared decisions remain open. Outreach
  separately requires a second immutable `developers.outreach.io` capture for its six cross-origin
  custom-object citations and must not reuse the `api.outreach.io` capture.
- Captain defect scope in this branch: all 20 owned locks are under comprehensive provider-surface audit, not only the eight named by the first magnitude spot-check. Any source that understates its provider inventory receives a full source-lock/API-surface/disposition remap; a verified no-change source is recorded explicitly. Salesloft is complete at 211 provider operations (formerly 12). The fleet-wide 58 locks with missing counts are not present in this branch and remain owned by their respective mapping lanes.
- Relaunch baseline: `READINESS-BASELINE.json` records 3,932 provider operations, 1,342 direct reads without a current exact command, 1,745 direct writes without an exact typed action, ten declared source/destination proofs, and three source inventories still requiring proof. Foundation `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` is an ancestor and supplies persisted App/CLI dispatch, exact action selection, and schema-validated camelCase fields. Help Scout's `update_conversation` is fixture/preflight proven; `update_customer` remains not enabled behind the separately registered action-specific source-binding foundation gap. Its five source-locked v3 direct reads are also not enabled behind the assigned `cli-operation-route-override-foundation-r1` common foundation. These are planning deficits, not safety or certification exclusions.
- Current green slice: repaired the Gorgias operation-ledger metadata that made CI `Verify` fail; regenerated the two stale website data projections; then completed Gorgias (`tickets → update_ticket`), Chatwoot (`contacts → update_contact`), Customer.io (`snippets → update_snippet`), Close (`leads → update_lead`), Outreach (`sequences → activate_sequence`), Zoho Bigin (`records → delete_record`), Chargebee (`customers → update_customer`), ServiceNow (`incidents → update_incident`), and Help Scout (`conversations → update_conversation`) source/destination declaration proofs. Unauthored mutations in Customer.io, Close, Zoho Bigin, and Chargebee correctly read `declaration-pending`; ServiceNow preserves its public fixed-template/dynamic-schema boundary. The `FOUNDATION-GAPS.json` register preserves resolved #4304 fan-out plus the two open Help Scout gaps: source-binding multiplicity and the five-operation route-override foundation. No credentials or live provider operations were used.
- Operation-evidence increments 1–2: `OPERATION-SURFACE-EVIDENCE.json` records all 3,932 provider operations with source URL/version/hash and seven named surface cells. The stale generic-destination gap was removed from all 1,111 direct-write rows across Salesloft, Copper, Klaviyo, Intercom, Freshdesk, Segment, ActiveCampaign, Iterable, Square, and Braintree; each correctly remains connector-local declaration-pending because no typed action exists. The five Help Scout v3 route-gap rows and one action-specific source-binding gap remain open and make the portfolio non-merge-ready.
- Latest stacked proof: `origin/fm/cli-reverse-etl-destination-r1` is still
  `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` and an ancestor. A fresh installed binary exposed
  the Help Scout declarative destination, exact `update_conversation` selection, all modes, and
  durable acknowledgement; the real connection-create preflight stopped at a deliberately absent
  credential before provider I/O. The route-override foundation is not yet published, so this PR
  remains non-merge-ready and must not claim deployed reverse ETL.
