# Plan — exhaustive GitHub operation and command proof

## Goal

Close the captain-ordered proof contract on one frozen branch commit. The proof must derive its
inventory from the GitHub bundle, account for every documented REST/GraphQL endpoint, every stream,
write action, operation declaration, and CLI command, and keep provider-specific behavior in
GitHub-owned definitions/generator/hooks/tests. The one shared production adjustment is the
provider-neutral safe handling of slash-bearing `ref` path variables. Fixture replay and preflight remain supporting evidence; terminal live
records are limited to `PROVEN`, `UNTESTABLE` with a concrete reason, or `FAILED`.

## Current-ref — merged GitHub surface

The authoritative source-derived count and its provenance are in
[VERIFICATION.md](VERIFICATION.md). This frozen plan does not duplicate that
generated-surface measurement.

The frozen checkpoints below remain historical records preserved at base ref
`4df0b0416e46958d9acb1b02708464570c070e0f` on 2026-08-10.

## Scope

- Generate `.planning/phases/github-parity-extract-r1/OPERATION-PROOF-LEDGER.json` and
  `COMMAND-PROOF-LEDGER.json` from `api_surface.json`, `streams.json`, `writes.json`,
  `operations.json`, and `cli_surface.json`.
- Re-run every GitHub command through the current built binary and record dispatch/blocked
  outcomes, including aliases and the 38 write / 23 stream contracts with no dedicated command.
- Add deterministic provider-double accounting for every declared stream, write action, and
  operation, using existing typed engine/conformance contracts and the generic ETL/reverse-ETL
  routes where no connector command exists.
- Run the committed GitHub-only live harness against the current binary when the approved private
  test credential is available, write only to that repository under plan → preview → approval →
  execute, and record a machine-readable report plus a concise Markdown summary without secrets.
- Prove the declared limiter stops before provider rejection and that an independent request still
  has provider headroom; record only sanitized counts/headers.
- Audit the shared-core diff for provider-neutral behavior and record GitHub-only ownership.

## TDD / checkpoints

1. **Red — ledger contract and proof inventory.** Tests fail when a source endpoint/action/stream/
   operation/command is omitted, a `covered_by` target is unknown, an operation command or alias
   is unbound, a generic-only surface member lacks an exercise route, or a shared generated delta
   is not GitHub-only.
2. **Green — source-derived ledgers.** The generator/validator writes the two JSON ledgers and
   reports exact 1,224 endpoint, 37 stream, 574 write-action, 377 operation, and 1,179 command
   source counts without hand-maintained counts.
3. **Red — current-head binary proof.** A regression fixture with an unknown command and a blocked
   command must fail the sweep contract; the real sweep then records every final command and its
   rendered routing evidence from a freshly built binary.
4. **Green — deterministic contract proof.** Existing provider-double tests and the new exhaustive
   accounting run cover every stream/action/operation, including generic-only members. No raw
   request escape hatch, token-shaped fixture, or skipped safety gate is permitted.
5. **Green or explicit external blocker — live proof and limiter.** The live harness validates the
   dedicated credential/repository before starting `pm`, executes every case, and writes only
   terminal records. If the approved credential or provider window is unavailable, record the
   exact externally verifiable blocker rather than fabricating proof.
6. **Verify / review.** Run the scoped and repository gates, update `VERIFICATION.md`, execute the
   generated `verify-work` and `code-review` prompts inline because this runtime has no compatible
   isolated GSD roles, then run no-mistakes only after all evidence is committed.

## Required skills and lifecycle evidence

- GSD adapter path: `scripts/gsd doctor`, `scripts/gsd sources <command>`, generated
  `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts;
  inline/manual fallback is required by the single-worker contract.
- Loaded Go skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-documentation`, and `golang-lint`.
- CLI parity checks: runtime help, bare `pm github`, command help, generated manuals/goldens,
  website catalog confinement, and `--json`/approval/confirmation documentation.

## Verification commands

```bash
node --test scripts/tests/github-parity-proof.test.mjs
node scripts/github-parity-proof.mjs --check
node scripts/github-parity-proof.mjs --write
node scripts/github-command-reachability.mjs --pm ./pm --root <isolated-project>
go test -timeout 20m ./cmd/connectorgen/ ./internal/connectors/engine/ ./internal/connectors/commandrunner/ ./internal/connectors/conformance/ ./internal/connectors/certify/ ./internal/connectors/hooks/github/
go test -timeout 20m ./internal/app/ ./internal/cli/
go vet ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
make tidy-check lint docs-check smoke-no-build agent-contract-check connector-boundary release-workflow-check
bash scripts/verify-gsd-workflow
```

Live verification additionally uses the approved private repository and credential without printing
or serializing secret values; reverse ETL remains plan, preview, approval, execute.

## Follow-up / human gates

- Merging to `main` remains captain-only.
- The paused `fm/cli-top50-sweep-resume2-r1` branch must be rebased onto the resulting main after
  this PR lands; its copied GitHub commits and shared ledger will conflict otherwise.
- `issue delete` remains blocked because only a GraphQL mutation is documented and no GraphQL
  mutation executor exists. `repo delete` retains typed destructive confirmation; `repo create`,
  archive/unarchive, and secret-set are approval-gated non-destructive writes as decided by the
  captain.
