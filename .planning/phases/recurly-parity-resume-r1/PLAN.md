# Recurly parity-resume r1 plan

## Scope and ownership

Own `internal/connectors/defs/recurly/**`, Recurly-only generated documentation/data, this phase's
GSD artifacts, and required generated CLI transcript coverage. The recovered commit `6b3224e7f`
is a source snapshot, not a branch to resurrect. Review fixes may also change the smallest shared
runtime boundaries required for provider-neutral decimal flags, declarative-write body/retry
safety, and fixture request/response replay. Do not modify other connector bundles or legacy
connector Go.

## Manual-GSD fallback

`scripts/gsd doctor` passed, but `scripts/gsd prompt programming-loop init --phase
recurly-parity-resume-r1 --dry-run` returned `unknown GSD command: programming-loop`. This phase
therefore follows the repository's permitted manual-GSD fallback: this plan, TDD ledger,
verification checklist, prompt snapshot, summary, and run state are maintained directly.

## Required skills and references

- Skills: `gsd-programming-loop`, `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, `golang-cli`, `golang-documentation`, `golang-lint`, and `no-mistakes`.
- References: `AGENTS.md`, shared connector-parity resume contract, migration handoff and
  conventions, architecture design, GSD adapter/runtime-loop docs, issue delivery contract, and
  CLI help/docs/website parity guidance.

## Delivery slices

1. Record the red recovery preflight, then restore only the Recurly bundle from `6b3224e7f` on top
   of current `origin/main`. Do not restore historical shared or cross-connector changes.
2. Run the current validator to reproduce and fix the six required CLI flags: `currency` for
   invoice preview; `currency` and `account.code` for purchase preview; `product_code`,
   `unit_amount`, and `currency` for gift-card preview.
3. Use Recurly's provider-owned v2021-02-25 OpenAPI YAML as the field research source. Build a raw
   provider-field research matrix for every path/query/body/form/file input, preserving operation
   URL/section, evidence tier, confidence, requiredness source, and Tier-5 deferrals. Preserve
   those files as raw research and do not invent, modify, or integrate the shared citation
   convention in this phase.
4. Promote and execute-test the three bounded binary download operations now supported by current
   commandrunner. Retire stale blocker wording only after execution evidence exists.
5. Reconcile every operation in the provider OpenAPI with the bundle. Map record-shaped mutations
   to `writes.json`; do not mark a one-shot `rest_write` command implemented. Establish and report
   the actual operation total from the provider spec rather than relying on the recovered count.
6. Run the connector gates, build `pm`, execute representative help and Recurly commands, regenerate
   Recurly-owned docs/catalog/website output late, commit coherent green slices, and then run
   `no-mistakes` on the committed feature branch.
7. Before the `no-mistakes` run, compare every Recurly artifact that existed on current
   `origin/main` with the recovered definition. Preserve the five legacy streams' query defaults,
   schema primary/cursor/required metadata, typed fields, computed projections, and representative
   fixture records; add a connector-local regression test for that recovery invariant.
8. Reconcile review findings against the pinned provider OAS, then regenerate affected Recurly
   request schemas, write fixtures, non-legacy response schemas, and stream fixtures from that
   source. Preserve the five current-main legacy stream contracts and the raw provider-research
   artifacts byte-for-byte.
9. Add provider-neutral decimal flag coercion, required-empty-JSON body support, stable per-record
   idempotency headers for actions that declare provider support, fail-safe single-attempt execution
   otherwise, expected-query matching for write fixtures, and response-header replay for
   link-pagination fixtures. Document the billing-write retry policy explicitly.
10. Remove Recurly's newly introduced hard runtime limiter because the provider documents different
    sandbox and production/GET limits; retain only evidence-backed informational metadata.
11. Preserve automatic retries for actions that explicitly declare idempotent delete semantics,
    while keeping every other unkeyed mutation single-attempt. Regenerate the complete Recurly
    mutation-query boundary from the pinned OAS, require callers to choose refund and account
    redaction behavior, and record connector-local raw retry evidence without changing the shared
    citation convention.
12. Treat subscription termination `charge` as a required caller decision at the Recurly generator
    boundary, alongside `refund`. Regenerate its closed record schema, plain query mapping, CLI
    flag, fixture, raw evidence, manual, and website data without changing shared query omission
    behavior.

## TDD checkpoints

- Red: Recurly bundle absent on current main; `connectorgen validate` must fail before recovery.
- Red: recovered preview commands fail current required-flag validation; record the six findings.
- Green: validator, conformance, runtime preflight, and targeted CLI tests pass after flag fixes.
- Red/green: each binary command must fail before its metadata promotion and execute successfully
  against a bounded fixture/replay path afterward.
- Red/green: a recovery-preservation test must demonstrate the legacy-stream metadata loss before
  restoring it from current `origin/main`, then prove the restored metadata stays present.
- Red/green: focused regression tests must fail before decimal coercion, required-empty-body
  emission, idempotency-bound retries, expected-query matching, response-header replay, and the
  selected Recurly OAS-shape corrections, then pass together after the complete review-fix round.
- Red/green: reasoned pre-fix inspection proves explicitly idempotent deletes are forced to one
  attempt and all three Recurly mutation query controls are absent. Add focused regressions before
  production edits, apply the whole fix round, then run one focused test command at the end.
- Red/green: call-path inspection proves optional `record.charge` passes schema and preview but
  cannot trigger shared `omit_when_absent` handling. Require it in the generated contract and add
  a focused preview-validation regression before the single end-of-round verification.

## Verification plan

Run the shared-resume contract gates on current main: surface sync (write and check), Recurly
validation, Recurly conformance, global implemented-command preflight, targeted CLI tests, vet and
build for changed packages, `pm` help/command execution, and `website` data generation. Do not
claim full GSD verification: `make verify` is intentionally excluded by the shared contract because
its full test suite exceeds bounded command windows.

## Orchestration decision

`local_critical_path`: the task has a single connector-owned write scope and the current task
instructions prohibit proactive subagent delegation.
