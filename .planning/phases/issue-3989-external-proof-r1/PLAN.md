# #3989 external-binary certification proof plan

## Delivery header

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
