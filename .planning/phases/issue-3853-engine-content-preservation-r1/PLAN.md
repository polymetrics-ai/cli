# PLAN — issue #3853 engine content preservation

Issue: #3853.
Branch: `fm/cli-found-engine-content-preservation-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase 3853`; executed inline, with fixed decisions
  captured in `CONTEXT.md` and `DISCUSSION-LOG.md`.
- Plan prompt: `scripts/gsd prompt plan-phase 3853 --tdd`; executed inline in this plan.
- Execute prompt: `scripts/gsd prompt execute-phase 3853`; execute inline through the ordered
  RED/GREEN/REFACTOR slices below.
- Verify prompt: `scripts/gsd prompt verify-work 3853`; execute inline after targeted validation.
  If it finds a real gap, use `plan-phase 3853 --gaps`, then `execute-phase 3853 --gaps-only`.
- Review prompt: `scripts/gsd prompt code-review 3853`; execute inline with the result recorded in
  `REVIEW.md` before handoff.
- Inline/manual fallback: compatible isolated GSD roles are unavailable and the canonical
  single-worker/no-spawn task contract forbids role spawning. The fallback changes no TDD,
  verification, review, or human gate.

## Required skills loaded

- `golang-how-to` — route this shared Go engine/test task.
- `golang-design-patterns` and `golang-structs-interfaces` — retain the existing engine/transport
  boundary instead of creating a parallel error path or API.
- `golang-error-handling`, `golang-security`, and `golang-safety` — preserve wrapped error context,
  bounded transport behavior, request safety, and no unintended data mutation.
- `golang-testing` — retain red-to-green, observable `httptest` regression coverage.
- `golang-cli` and `golang-documentation` — keep the existing help/manual/golden/website contract
  synchronized.
- `vercel-react-best-practices` and `vercel-composition-patterns` — reviewed for the website-doc
  surface; no React component or composition code changes are expected.
- `frontend-design` and `web-design-guidelines` are not installed in this environment; the
  available Vercel guidance is the fallback for the documentation-page change.

## Slice A — test-first engine preview contract

1. RED: reverse, never delete, the existing preview tests that assert masking. Retain actions with
   `RedactFields`, use plainly synthetic fixture values, and assert that resolved preview warnings
   retain declared fields, nested fields, and configuration-secret substitutions exactly.
2. Run only the named preview tests before production edits and record the expected red failure in
   `TDD-LEDGER.md`.
3. GREEN: change only preview interpolation in `write.go` so it resolves the unmodified runtime
   configuration and record. Remove preview-only masking helpers once unused, while retaining any
   separate write-action error behavior outside this issue.
4. REFACTOR: prove `redact_fields` declarations still load and are simply no longer interpreted by
   preview resolution; do not rewrite any bundle.

## Slice B — test-first engine error-content contract

1. RED: add focused `httptest` cases for direct read, operation direct read, and binary download.
   Each returns a non-success response with a harmless query value and JSON diagnostic body that
   the current `safety.RedactErrorText` flow strips. Assert the final engine error preserves both.
2. Run only these named tests before production edits and record the observed red failure.
3. GREEN: remove the engine calls to `safety.RedactErrorText`. When an error is a bounded
   `connsdk.HTTPError`, format from its raw captured URL/body fields so the shared transport's
   `Error()` rendering cannot re-mask the content. Preserve the existing class/hint ordering and
   generic non-HTTP error path.
4. REFACTOR: centralize only the shared engine rendering helper; do not touch `connsdk`, runner,
   output-policy behavior, or binary record field redaction.

## Slice C — operator contract parity

1. RED/GREEN: replace the current masking assertions in `internal/cli/docs.go`, generated
   `docs/cli/reverse.md`, golden transcripts, and `website/content/docs/reverse-etl.mdx` with the
   truthful complete-content boundary. Preserve the distinction between connector-engine content
   and the separately-owned generic source-table plan sample path.
2. Keep approval tokens described as time-bounded, single-use authorization capabilities omitted
   from JSON output. Keep destructive confirmation, digest revalidation, and non-idempotent
   no-retry claims intact.
3. Regenerate only the prescribed checked-in manual/golden artifacts through existing project
   commands; do not manually manufacture broad generated catalog changes.

## Verification plan

- Red/green engine tests:
  `go test ./internal/connectors/engine -run 'TestDryRunWritePreviewResolved(MethodPath|Path).*Preserves|Test(DirectRead|OperationDirectRead|BinaryDownload).*Preserves.*Error' -count=1`.
- Package regression: `go test ./internal/connectors/engine`.
- CLI/docs regression: targeted `go test ./internal/cli -run 'TestGoldenTranscripts|Test.*Docs'`;
  then separate `go test ./internal/cli` under the repository's timeout guidance.
- Runtime help/manual parity after building `pm`: `./pm reverse`, `./pm help reverse`, and
  `./pm reverse --help`; all must exit successfully and agree with `docs/cli/reverse.md`.
- Website/docs parity: `rg -n 'complete.*content|redact_fields|mask' internal/cli/docs.go
  docs/cli/reverse.md website/content/docs/reverse-etl.mdx`; run `./pm docs validate
  --connectors-dir docs/connectors` and the appropriate existing generator/check.
- Formatting/static: `gofmt -w` only changed Go files; `go vet ./internal/connectors/engine`;
  `go vet ./internal/cli`; `go build ./cmd/pm`; `git diff --check`.
- Applicable repository gates individually: `make tidy-check`, `make lint`, `make docs-check`,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check`.
- Do not run whole `go test ./...` or `make verify` as a single local command; CI owns the
  timeout-prone 550+ connector suite. Do not run live-provider, credentialed, or reverse-ETL
  execution checks.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. RED test checkpoint, retaining observed failures in the ledger.
3. GREEN engine mechanism and documentation parity checkpoint.
4. Verification/review documentation checkpoint.

## Safety and non-goals

- No provider call, credential, bundle rewrite, schema/policy enum change, capability claim,
  command-runner change, generic source-table masking change, generic shell/HTTP/SQL write tool,
  dependency, or reverse-ETL execution.
- Existing `redact_fields` values stay declaration-load-compatible. The behavior changes solely at
  this engine boundary.
- Do not claim a command implemented without matching `api_surface`; this foundation creates no
  operation or command declaration.
