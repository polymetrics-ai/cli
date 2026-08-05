# Recurly parity-resume r1 plan

## Scope and ownership

Own only `internal/connectors/defs/recurly/**`, Recurly-only generated documentation/data, this
phase's GSD artifacts, and required generated CLI transcript coverage. The recovered commit
`6b3224e7f` is a source snapshot, not a branch to resurrect. Do not modify shared
engine/schema/validator files, other connector bundles, or legacy connector Go.

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
   URL/section, evidence tier, confidence, requiredness source, and Tier-5 deferrals. Do not
   invent the in-flight citation convention; rebase onto its landed form before final validation.
4. Promote and execute-test the three bounded binary download operations now supported by current
   commandrunner. Retire stale blocker wording only after execution evidence exists.
5. Reconcile every operation in the provider OpenAPI with the bundle. Map record-shaped mutations
   to `writes.json`; do not mark a one-shot `rest_write` command implemented. Establish and report
   the actual operation total from the provider spec rather than relying on the recovered count.
6. Run the connector gates, build `pm`, execute representative help and Recurly commands, regenerate
   Recurly-owned docs/catalog/website output late, commit coherent green slices, and then run
   `no-mistakes` on the committed feature branch.

## TDD checkpoints

- Red: Recurly bundle absent on current main; `connectorgen validate` must fail before recovery.
- Red: recovered preview commands fail current required-flag validation; record the six findings.
- Green: validator, conformance, runtime preflight, and targeted CLI tests pass after flag fixes.
- Red/green: each binary command must fail before its metadata promotion and execute successfully
  against a bounded fixture/replay path afterward.

## Verification plan

Run the shared-resume contract gates on current main: surface sync (write and check), Recurly
validation, Recurly conformance, global implemented-command preflight, targeted CLI tests, vet and
build for changed packages, `pm` help/command execution, and `website` data generation. Do not
claim full GSD verification: `make verify` is intentionally excluded by the shared contract because
its full test suite exceeds bounded command windows.

## Orchestration decision

`local_critical_path`: the task has a single connector-owned write scope and the current task
instructions prohibit proactive subagent delegation.
