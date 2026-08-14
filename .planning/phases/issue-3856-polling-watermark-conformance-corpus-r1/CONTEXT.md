# Context — #3856 immutable polling-watermark conformance corpus

**Primary issue:** #3856
**Parent:** #3855 / draft PR #4041
**Branch:** `feat/3856-polling-conformance-corpus` from
`origin/feat/3855-polling-apply-foundations@fa5eef681a4b06c09519574326a22683b26bd996`

## Locked decisions

- The child PR base is exactly `feat/3855-polling-apply-foundations`; it is a
  draft stacked PR and must never target `main`, #3862, or a combined branch.
- The corpus is separate from `internal/synccontract/testdata/conformance/v1.json`.
  The generic #3810 digest remains untouched.
- The corpus is embedded, versioned, defensive-copy only, and invoked through
  one no-skip runner. A lane factory cannot replace fixtures, pass filters, or
  omit a mandatory scenario.
- The runner validates a registered polling executor/descriptor plus exact
  corpus evidence before it invokes a lane. This is a conformance seam only;
  #3857 remains the owner of product preflight/admission.
- The first test is
  `TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture`. Its first
  transcript has equal-watermark identities across physical pages, durable
  acknowledgement followed by failed checkpoint persistence, and restart from
  the prior committed #3810 envelope without a lost stable identity.
- The reference lane is deterministic and fake-only. It is not a credentialed
  provider test, product executor, or certification claim. Credential-backed
  Transport remains deferred to the first executable provider-boundary child
  and final #4019 gate.
- Mandatory fixture cases are: equal-watermark split/recovery; empty page;
  non-advancing page; NULL/precision/coercion policy; unstable/non-unique
  keyset rejection; bounded overlap/commit-lag safety; source generation/schema
  mismatch; acknowledgement-before-checkpoint replay; tombstone/history versus
  hard-delete invisibility; and missing executor/evidence admission.

## Deferred work

- #3857: polling descriptor and real runtime preflight, including the public
  non-CDC capability decision.
- #3858: provider-bound page source and `RunETL` execution.
- #3859: apply strategies.
- All PostgreSQL adapters, live credentials, provider mutation, and CDC
  promotion.

## GSD execution mode

`scripts/gsd doctor`, all five command source resolutions, and
`go run ./cmd/agentcontractgen check` passed. The generated
`discuss-phase --auto`, `plan-phase --tdd --auto`, and `execute-phase --auto`
prompts were inspected. The supplied brief decides every material gray area,
and `--auto` is the explicit non-interactive authorization. The repository's
single-worker contract disallows GSD role spawning, so the generated workflow
is executed inline with durable artifacts.
