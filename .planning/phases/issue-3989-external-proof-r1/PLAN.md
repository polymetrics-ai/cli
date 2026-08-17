# #3989 external-binary certification proof plan

## Residual plan — live and OS-level proof

## Bounded credential-scope gap — 2026-08-17

The integration base now includes #4215, whose accepted-evidence schema v2
allows the narrower `observed_operations` credential claim only when the
record contains a matching `protocol_exchanges` proof discriminator. The
existing #3989 external-proof command still rejects `--external-proof` unless
`--full-parity` is present in both its parent and fresh-child paths, and its
writer rejects a completed non-parity run before it can persist the exchanges
it actually observed. That gate is obsolete for an external proof whose claim
is explicitly bounded; it must not be worked around by pretending that a
failed schedule, resume, or write sweep passed.

**Scoped change:** external-proof artifacts advance to schema v2 and derive
their credential scope inside the writer. A passing completed full-parity
report may retain `full_parity` / `full_parity_stage`; every other completed
run may publish only `observed_operations` / `protocol_exchanges`, and only
after a complete, bounded HTTPS transcript contains an observable successful
provider response. The artifact preserves the actual process exit code and
does not convert a non-passing certification report into a pass. Full-parity
flow references remain required only for the full-parity claim.

**Separate live findings (not scope-expansion targets):** prior authorized
disposable-identity runs reached `schedule_create` with typed CLI error exit
3 and, on a later run, `resume` with typed CLI error exit 1. These are retained
as distinct redacted findings in verification and the PR body. They are not
evidence of parity, are not retried blindly, and are not patched around in
this issue.

### Bounded-scope TDD slice

1. **Red:** add a writer test for a complete external child transcript with a
   provider-success exchange but a nonzero certification exit; it expects a
   schema-v2 `observed_operations` / `protocol_exchanges` proof and raw
   credential absence, so the current full-parity refusal fails. Add a fresh
   TLS-child CLI test without `--full-parity`; it expects the parent to relay
   the child’s honest nonzero exit while the child writes that bounded proof.
2. **Green:** remove only the parent/child `--full-parity` requirement; derive
   the proof scope from the completed report and its captured exchanges rather
   than from a caller-provided scope string. Keep the nonzero report exit,
   full-parity flow-reference check, truncation refusal, safe argv, and secret
   scans intact. Update the opt-in GitHub smoke to assert the v2 bounded claim,
   an observed provider success, exact transcript verification, and secret
   absence.
3. **Refactor/review:** retain separate redacted stage diagnostics for
   `schedule_create` and `resume`, update help/manual/website text and golden
   expectations to say external proof can make a bounded claim, regenerate
   generated docs/data, and inspect the diff for any raw credential path.

### Bounded-scope execution record

The planned writer red failed with `external proof requires a completed
successful process`; the fresh-child red failed with the obsolete
`--external-proof requires --full-parity` usage gate. The green implementation
derives the v2 claim in the writer and removes only that parent/child gate.
`TestWriteExternalProofRefusesBoundedClaimWithoutSuccessfulProviderResponse`
keeps a zero-write refusal for an all-unsuccessful transcript.

The live GitHub smoke then ran with the named disposable identity and passed:
the real fresh child produced one secret-free `observed_operations` /
`protocol_exchanges` artifact, observed a GitHub 2xx, and verified its exact
child transcript. This is a bounded claim, not a substitute for full parity.
Earlier full-parity `schedule_create` (typed error, exit 3) and `resume`
(typed error, exit 1) failures are recorded separately rather than retried or
changed here. A local one-route Recurly fixture continues to demonstrate the
OS boundary; its exact incomplete-certification exit is now 2 under the
current failure-code contract, with a concise fingerprint-redacted diagnostic
showing its intentionally unserved route.

### Task Delivery Header

- Issue: Refs #3989 — Certification: add external-binary proof capture and ephemeral fingerprint-first credentials.
- Base branch: `integration/4015-mvp-flat-r1` at refreshed head `c9791db4d`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Commit the focused residual evidence and GSD records, push `fm/cli-3989-live-proof-residual-r1`, open a Conventional Commit PR to the exact base, then API-read the PR base.
- Working branch: `fm/cli-3989-live-proof-residual-r1`.
- Task: Execute the existing opt-in live GitHub proof with its disposable identity and extend only its test evidence for opaque body substitution, OS command-list/temporary paths, same-run two-credential/root-salt semantics, and a readable credential-fingerprinted diagnostic when the live smoke fails. Maintain the aggregate `internal/cli` package below its fixed 20-minute capacity by parallelizing only independent dynamic-help assertions; no test coverage is sampled, deleted, or moved behind an opt-in.
- Verification: `go test -timeout 20m ./internal/connectors/certify`, `go test -timeout 20m ./internal/cli`, the explicitly credentialed live smoke, `go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, each `make verify` constituent gate, website docs generation twice, `git diff --check`, and automated review after PR creation. The live-smoke failure diagnostic must contain no exact, base64, or URL-escaped prepared secret while retaining its non-secret provider reason.

### Required GSD and skills

- Generated inline prompts: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` for `issue-3989-external-proof-r1`; gap closure uses `plan-phase --gaps` then `execute-phase --gaps-only`.
- Manual fallback: the phase is not registered in the GSD roadmap and the canonical contract forbids role spawning for this single-worker issue. The generated sequence is completed inline in these artifacts.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-benchmark`, `golang-performance`, and `golang-troubleshooting`.

### Residual TDD slices

1. **Opaque proof evidence:** Red—add the opaque request/response canary assertion and run it before its test support exists. Green—serialize an accepted complete opaque exchange and assert raw-canary absence plus exact fingerprint presence for both bodies. Refactor—keep the original bounded-capture refusal as its own negative test.
2. **Fingerprint semantics:** Red—add one two-credential proof test expecting repeated A marker equality, B marker difference, and no serialized salt. Green—assert the existing normalized prepared-value boundary produces those facts in one proof. Refactor—retain the separate cross-root replay assertion as its independent salt-rotation contract.
3. **Real OS boundary:** Red—add a blocked external-child test whose helper is initially absent. Green—after the fresh child resolves the credential into memory, have that child record a secret-safe snapshot of its own process-list entry, argv, project root, runner workdir, and fresh-binary directory; after ordinary completion the parent verifies the artifact reports no raw credential in any location. The one-route Recurly TLS fixture deliberately expects its strict incomplete-full-parity exit after the snapshot: it has authenticated only `/accounts`, not Recurly's declared write surface. Refactor—do not hold/release the child or poll parent-side lifecycle state; stream scanned files, redact any detected value before serialization, and keep the evidence request unconditional in CI.
4. **Live provider evidence:** Run `TestExternalProofGitHubSmoke` only with the disposable identity named by the brief and retain sanitized command/result metadata (never secret material) in the ledger and verification record. The smoke itself independently asserts observable GitHub-backed proof creation and read-back.
5. **CLI package capacity:** Red—run the complete `internal/cli` package with verbose timings and establish that the fixed 20-minute deadline is aggregate capacity, not a Bahmni-specific hang; the 706.417s baseline ranks the external-child proofs first (118.770s and 118.210s), Bahmni at 39.140s, and the independent 17,800-case dynamic-help sweep at 22.500s. Green—make each connector-command/`--help` or `-h` assertion an explicitly parallel subtest while retaining the same generated command set, both flag variants, all three assertions, and the nonzero total. The focused verbose run still reports 17,800 variants, the focused race run passes, and the unchanged full-package deadline passes in 694.432s. Refactor—share only the immutable loaded registry and per-command immutable surface; every subtest owns its arguments and output, with no test timeout, sampling, or coverage reduction.
6. **Live failure diagnostic:** Red—add planted-secret tests for the child-stream, failed-flow, missing-flow, failed-report, and typed-stage-error paths before their fingerprinted diagnostic support exists. Green—use an in-memory random HMAC salt and the existing `{{pmcertfp:v1:<hash>}}` format to replace exact, base64, and URL-escaped credential forms while preserving the non-secret category/code/message and stage status/exit/kind. Refactor—route every `assertKind` mismatch through this safe envelope path; an external child may render only the checked diagnostic and never its raw captured stream. A redaction error reports no captured output; diagnostics create no proof, root salt, artifact, or log file.

### CLI parity assessment

This gap changes `--external-proof` semantics and its persisted proof schema.
Update runtime help, the embedded manual source, `docs/cli/connectors.md`, the
website command reference, and their golden/help assertions. Verify `pm
connectors`, `pm help connectors`, and `pm connectors certify --help`; the bare
namespace is unchanged but remains an explicit parity check. No completion
metadata or direct-read candidate surface changes.

## Original merged-slice header — historical

- Issue: Closes #3989 — Certification: add external-binary proof capture and ephemeral fingerprint-first credentials.
- Base / head: `integration/4015-mvp-flat-r1` → `fm/cli-3989-external-proof-r1`.
- Required GSD route: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.
- Manual fallback: phase 3989 is absent from the roadmap (`gsd-sdk query init.phase-op 3989` returned `phase_found:false`), and compatible role spawning is unavailable. This plan, ledger, run-state, verification, and review record execute that route inline.

## Skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-design-patterns`, and `golang-structs-interfaces`.

## TDD slices

1. **Red:** test an external child invocation with canary values in authorization, query, JSON/opaque request, response header, JSON/opaque response, and output. It must fail before the production observer exists. **Green:** provide an ephemeral credential/process boundary and assert no vault/profile/argv/artifact/log contains a canary while every proof location holds its deterministic fingerprint.
2. **Red:** test that evidence rejects absent exchanges, child failure, incomplete substitution, non-full-parity input, and truncated required body; assert no artifact is written. **Green:** wire the all-or-nothing proof model to the completed external run and make each refusal leave zero evidence writes.
3. **Red:** exercise HTTPS redirects, retryable error then success, binary/opaque payload, and provider error body against the real built binary; assert missing observations or unbounded data fails. **Green:** collect bounded full-fidelity transcript observations with explicit truncation metadata while preserving child-visible bytes.
4. **Red:** external-binary smoke acceptance must prove its stdout/stderr equals observed process output and a filesystem audit before cleanup finds no certification vault credential/key. **Green:** persist only sanitized accepted evidence after the full child run succeeds.
5. **Refactor/review:** retain in-process harnesses for fixtures, remove them from accepted live evidence, add the explicit `--external-proof --full-parity` help/manual/website contract, and run focused plus repository gates.

## Acceptance evidence

| Criterion | Evidence | Observable assertion / required side effect |
| --- | --- | --- |
| Ephemeral credential intake | fake HTTPS provider + filesystem audit | canary is absent from project tree, vault/key/profile, argv, transcript, stdout/stderr, and temporary artifacts; fingerprint appears at every observed proof location. |
| Stable/local fingerprints | deterministic unit test | same salt and value produce same fingerprint; different salt produces a different fingerprint. |
| No persisted credential | fake + filesystem audit before cleanup | zero certification credential/key/profile writes; refusal paths write zero artifacts. |
| Fresh external full transcript | real freshly built `pm` + HTTPS test provider | observed child stdout/stderr match stored result and transcript contains method/query/body/status/response; incomplete run is rejected. |
| Bounded redirects/retries/binaries/errors | HTTPS provider fake | explicit exchange count, redirect/retry/error observations, byte ceiling and truncation fields; a proof requiring truncation is refused. |
| Safe argv and flow read-back | fresh external binary + refusal fake | artifact retains the exact credential-free child argv and references successful `flow_plan`, `flow_preview`, `flow_run`, and `flow_status`; a missing reference produces zero proof writes. |
| Live GitHub smoke | opt-in live | accepted sanitized evidence only when already-configured credential reference exists; no secret is emitted or persisted. |

## Verification

- `go test -timeout 20m ./internal/connectors/certify/...`
- targeted build/external-process acceptance test
- `go vet ./...` and relevant `make verify` component gates individually
- provided PostgreSQL container command, kept separate from #3989 source changes, to validate the live environment as requested
- final `git diff --check`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and fresh rebase before push.

## CLI parity

`--external-proof --full-parity` is public certification CLI behavior. Its help text,
generated manual source, `docs/cli/connectors.md`, website CLI reference, and golden
transcript are updated together. Verify `pm connectors`, `pm help connectors`, and
`pm connectors certify --help` alongside the docs check.
