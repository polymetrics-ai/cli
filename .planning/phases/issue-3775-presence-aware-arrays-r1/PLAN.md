# PLAN — issue #3775 presence-aware required string arrays

Issue: #3775. Integrated sub-issues: #3778, #3780, #3781, #3783.
Branch: `fm/cli-found-presence-aware-arrays-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase 3775`; executed inline, with fixed decisions
  captured in `CONTEXT.md` and `DISCUSSION-LOG.md`.
- Plan prompt: `scripts/gsd prompt plan-phase 3775 --tdd`; executed inline in this plan.
- Execute prompt: `scripts/gsd prompt execute-phase 3775`; to be executed inline through the
  ordered RED/GREEN/REFACTOR slices below.
- Verify prompt: `scripts/gsd prompt verify-work 3775`; to be executed inline after targeted
  validation. If it finds a real gap, use `plan-phase 3775 --gaps` then
  `execute-phase 3775 --gaps-only` before rerunning verification.
- Review prompt: `scripts/gsd prompt code-review 3775`; to be executed inline with findings
  recorded in `REVIEW.md` before final handoff.
- Inline/manual fallback: compatible isolated GSD roles are unavailable and this issue's canonical
  single-worker/no-spawn contract forbids spawning them. The fallback changes no TDD, verification,
  review, or human gate.

## Required skills loaded

- `golang-how-to` — routed this Go runtime/test task.
- `golang-testing` — table-driven, observable red/green tests using isolated fakes.
- `golang-error-handling` — preserve established error shape and return paths.
- `golang-safety` — retain raw-presence guards and avoid slice/nil mistakes.
- `golang-security` — preserve identifier/dangerous-character validation and avoid exposing secrets.
- `golang-design-patterns` — keep one shared validator rather than create executor/provider forks.
- `golang-structs-interfaces` — retain the existing fake connector/public interface test boundary.

## Slice A — #3778: contract and red checkpoint

1. RED: add a named, table-driven focused test for the common required-flag contract. It must
   assert absent map key and raw `[]string{}` fail as missing, a required scalar blank fails, a
   zero-minimum required `string_array` accepts `[]string{""}` and blank-only CSV, and
   `min_items: 1` rejects the same materialized empty array. Retain a `max_items` control.
2. Run only that new test before production edits and record the failure caused by the current
   post-coercion `[]string{}`-is-absent classification in `TDD-LEDGER.md`.
3. Keep the test as the permanent contract; do not hide the behavior in a test-only helper.

## Slice B — #3780: shared presence mechanism

1. GREEN: change only the owned shared validation behavior in
   `internal/connectors/commandrunner/runner.go` so an empty `[]string` no longer erases a
   previously established raw presence. Preserve blank scalar rejection, coercion validation,
   unknown-flag handling, unsafe-character rejection, and `min_items`/`max_items` errors.
2. Run the focused contract test and `go test ./internal/connectors/commandrunner`.
3. REFACTOR: keep the rule centralized in the existing validator; do not add a schema field,
   `connectorgen` mirror, provider branch, raw-body escape hatch, retry, or redaction path.

## Slice C — #3781: public execution-path integration

1. RED/GREEN: add a public `Run` operation-direct-read test whose required `string_array` maps to
   `body.items`; an explicit blank must reach the fake connector as a typed, literal `[]string{}`.
2. RED/GREEN: add a public `BuildWriteCommand` reverse-ETL test whose required `string_array` maps
   to `record.items`; an explicit blank must reach the planned record as `[]string{}`, retain
   `ApprovalRequired`, and leave the fake provider `Write` uncalled.
3. Add omitted/raw-empty/min-items-one negative controls proving neither connector operation nor
   write dispatch occurs. No remote write is executed.

## Slice D — #3783: durable regression enforcement

1. Fold the five semantic states into compact named table cases that pass through the actual
   materialization paths, covering both `body.*` and `record.*` targets.
2. Assert exact typed empty-array representation, not merely absence of an error; marshal or
   compare values without redaction/stripping.
3. Run the existing `TestEveryImplementedCommandPassesRuntimePreflight` alongside the focused
   package suite. This is a runtime guard only; no `connectorgen` rule is added.

## Verification plan

- Focused red/green: `go test ./internal/connectors/commandrunner -run 'Test.*Required.*StringArray|Test.*ExplicitEmpty.*'`.
- Package regression: `go test ./internal/connectors/commandrunner`.
- Runtime preflight: `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`.
- Formatting/static: `gofmt -w internal/connectors/commandrunner/runner.go internal/connectors/commandrunner/runner_test.go`; `go vet ./internal/connectors/commandrunner`; `go build ./cmd/pm`.
- Applicable repository gates individually: `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`.
- Do not run the timeout-prone whole `go test ./...` or `make verify` monolith locally; CI owns the
  550+ connector suite. Record any unavailable/full-pipeline step as such rather than estimating.
- CLI parity checklist: all help/manual/docs/website/generated-surface entries are N/A because no
  command declaration or rendered surface changes; verify with changed-path inspection.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. RED test checkpoint, retaining its observed failure in `TDD-LEDGER.md`.
3. GREEN mechanism and public-path regression checkpoint.
4. Verification/review documentation checkpoint.

## Safety and non-goals

- No Front bundle/docs change, provider call, credential, schema field, capability claim, output
  policy, redaction/masking, retry, generic write tool, or reverse-ETL execution.
- `internal/connectors/commandrunner/runner.go` changes stay within #3775-owned functions; any
  need to edit another lane's function is a stop-and-report condition.
- The existing executable-surface preflight sweep remains the guard against implementation claims
  without matching API-surface evidence; this foundation does not create declarations.
