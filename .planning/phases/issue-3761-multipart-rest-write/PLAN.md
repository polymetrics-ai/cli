# PLAN — typed multipart execution for declared `rest_write`

Issues: #3761 (parent), #3763, #3768, #3772, #3774, #3777.
Branch: `fm/cli-found-multipart-writes-r1`.

## GSD path

- `scripts/gsd doctor`: passed; 69 registered commands.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase 3761`.
- Plan prompt: `scripts/gsd prompt plan-phase issue-3761-multipart-rest-write --tdd --skip-research`.
- Execute, verify, and review prompts were generated with `scripts/gsd prompt
  ...` and completed through the documented inline/manual fallback. This
  issue-scoped phase is not a numbered ROADMAP phase and the worker brief
  prohibits role spawning, so no GSD/specialist role is spawned; `SUMMARY.md`,
  `VERIFICATION.md`, and `REVIEW.md` retain the equivalent execution,
  coverage, and disposition evidence.
- `go run ./cmd/agentcontractgen check` is included in the final non-aggregate
  verification gates.

## Required skills loaded

- `golang-how-to` — selected the applicable Go skill set.
- `golang-design-patterns` and `golang-structs-interfaces` — extend the
  existing declared executor with one closed optional type rather than a new
  transport abstraction.
- `golang-error-handling`, `golang-security`, and `golang-safety` — preserve
  contextual failures, strict root confinement, no generic write escape hatch,
  bounded IO, and no response masking.
- `golang-testing`, `golang-context`, and `golang-concurrency` — test the
  no-network preview, loopback request, cancellation-aware snapshot, single
  attempt, and pipe cleanup paths deterministically.
- `golang-cli` and `golang-documentation` — guard command materialization and
  record the deliberately inapplicable CLI/help/manual/website update.
- `no-mistakes` v1.41.2 — governs the later validation/shipping stage; never
  pass `--yes` and never edit during an active pipeline run.

## Dependency-ordered TDD slices

### 1. #3763 — closed contract and loader red evidence

1. RED: add operation-bundle tests that demonstrate existing
   `multipart/form-data` rest writes are unexecutable and that missing caps,
   body schema, declared fields, file requirements, or non-`rest_write` use
   are rejected.
2. GREEN: add `RESTOperationSpec.Multipart`, its closed meta-schema object,
   and semantic validation. Require literal `multipart/form-data`, a
   connector-relative path, positive aggregate/file and response caps, a
   recursively closed typed body schema, required string file source fields,
   and only declared parts.
3. REFACTOR: share the existing part media-type validator without changing
   `writes.json` behavior or provider bundles.

### 2. #3768 — canonical preview and approval binding

1. RED: add multipart operation preview tests for no network calls, changing
   digest on field/path changes, missing approved file digest, and stale
   approval rejection.
2. GREEN: factor the existing canonical multipart representation so a
   `rest_write` and a `writes.json` action serialize identical ordered field,
   source-path identity, cap, media-type, and SHA-256 material.
3. REFACTOR: keep actual file paths private to transport construction while the
   prepared request carries only canonical hashes. Re-preparation remains
   fail-closed before dispatch.

### 3. #3772 — bounded single-attempt multipart dispatch

1. RED: add a loopback multipart direct-write test that currently fails before
   any server hit because multipart is unsupported.
2. GREEN: add a limited multipart requester path, use it from the existing
   direct-write switch with `DisableRetries=true`, and reuse root-contained
   multipart transport construction. Assert boundary, field, filename, media
   type, bytes, response bound, and complete response/error content.
3. REFACTOR: add negative tests for changed/missing files, per-file and
   aggregate overflow, disallowed media, root escape/symlink swap, 429/5xx
   single attempt, and 307/308 non-replay.

### 4. #3774 — connector-author documentation

1. Document the stable operation-level contract in the migration convention
   and only the necessary architecture overview.
2. State that reverse-ETL multipart remains proven separately, legacy
   `file_upload` stays blocked, and adoption owns any new provider command and
   its CLI/manual/website parity.
3. Make no capability claim for GitLab, Freshchat, Gong, or any other
   connector in this foundation.

### 5. #3777 — executable-claim end-to-end guard

1. RED: create a minimal in-memory bundle plus command surface and show its
   multipart direct-write preflight/plan/preview/approval/loopback execution
   fails before the dispatcher exists.
2. GREEN: prove the real `commandrunner.Preflight` and App/engine/requester
   path succeeds only with the closed contract and approved payload digest.
3. REFACTOR: retain deliberate mutations proving legacy `file_upload`, a missing
   endpoint declaration, unbounded response, missing source/cap, or disabled
   dispatch cannot advertise `availability: implemented`.

## Verification checklist

- Focused RED/GREEN: `go test ./internal/connectors/engine -run
  'Multipart|OperationDirectWrite|Bundle|Schema' -count=1` and focused
  `connsdk`, `commandrunner`, and `app` loopback tests.
- Changed packages: `go test ./internal/connectors/engine
  ./internal/connectors/connsdk ./internal/connectors/commandrunner
  ./internal/app -count=1`; `go vet` on those packages; `go build ./cmd/pm`.
- Non-aggregate repository gates: `make tidy-check`, `make lint`, `make
  docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make
  connectorgen-validate`, `go run ./cmd/connectorgen surface-sync --check`,
  `make connector-boundary`, and `make release-workflow-check`.
- `go run ./cmd/agentcontractgen check`; `go run ./cmd/connectorgen validate`;
  grep checks confirming no definition adoption and no new redaction policy.
- CLI parity exemption: inspect no new command/help/manual/website surface;
  do not invoke a credentialed or live `pm` command.

## Commit checkpoints

1. Planning/context/TDD checkpoint.
2. Red contract and engine tests, preserving observed failure output in the
   ledger.
3. Green contract/preview/dispatch slices and focused gates.
4. Docs plus end-to-end guard, then review/verification fixes.

No no-mistakes run, push, PR creation, or merge occurs before firstmate asks
for that explicitly, per the worker brief's Definition of done.
