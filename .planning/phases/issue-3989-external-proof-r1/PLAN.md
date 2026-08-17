# #3989 external-binary certification proof plan

## Residual plan — live and OS-level proof

### Task Delivery Header

- Issue: Refs #3989 — Certification: add external-binary proof capture and ephemeral fingerprint-first credentials.
- Base branch: `integration/4015-mvp-flat-r1` at refreshed head `4a0289bcc`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: Commit the focused residual evidence and GSD records, push `fm/cli-3989-live-proof-residual-r1`, open a Conventional Commit PR to the exact base, then API-read the PR base.
- Working branch: `fm/cli-3989-live-proof-residual-r1`.
- Task: Execute the existing opt-in live GitHub proof with its disposable identity and extend only its test evidence for opaque body substitution, OS command-list/temporary paths, and same-run two-credential/root-salt semantics.
- Verification: `go test -timeout 20m ./internal/connectors/certify`, `go test -timeout 20m ./internal/cli`, the explicitly credentialed live smoke, `go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, each `make verify` constituent gate, website docs generation twice, `git diff --check`, and automated review after PR creation.

### Required GSD and skills

- Generated inline prompts: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` for `issue-3989-external-proof-r1`.
- Manual fallback: the phase is not registered in the GSD roadmap and the canonical contract forbids role spawning for this single-worker issue. The generated sequence is completed inline in these artifacts.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

### Residual TDD slices

1. **Opaque proof evidence:** Red—add the opaque request/response canary assertion and run it before its test support exists. Green—serialize an accepted complete opaque exchange and assert raw-canary absence plus exact fingerprint presence for both bodies. Refactor—keep the original bounded-capture refusal as its own negative test.
2. **Fingerprint semantics:** Red—add one two-credential proof test expecting repeated A marker equality, B marker difference, and no serialized salt. Green—assert the existing normalized prepared-value boundary produces those facts in one proof. Refactor—retain the separate cross-root replay assertion as its independent salt-rotation contract.
3. **Real OS boundary:** Red—add a blocked external-child test whose helper is initially absent. Green—after the fresh child resolves the credential into memory, have that child record a secret-safe snapshot of its own process-list entry, argv, project root, runner workdir, and fresh-binary directory; after ordinary completion the parent verifies the artifact reports no raw credential in any location. The one-route Recurly TLS fixture deliberately expects its strict incomplete-full-parity exit after the snapshot: it has authenticated only `/accounts`, not Recurly's declared write surface. Refactor—do not hold/release the child or poll parent-side lifecycle state; stream scanned files, redact any detected value before serialization, and keep the evidence request unconditional in CI.
4. **Live provider evidence:** Run `TestExternalProofGitHubSmoke` only with the disposable identity named by the brief and retain sanitized command/result metadata (never secret material) in the ledger and verification record. The smoke itself independently asserts observable GitHub-backed proof creation and read-back.

### CLI parity assessment

The residual changes no command, flag, output schema, help string, docs source, website source, generated manual, or completion metadata. `pm connectors`, `pm help connectors`, and `pm connectors certify --help` are nevertheless re-run and recorded as unchanged-surface checks; docs/website edits are explicitly not applicable.

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
