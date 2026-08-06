# PLAN — issue #3628 Email IMAP/SMTP native connector

Issue: #3628. Parent: #3620. Integrated children: #3660, #3662, #3664, #3666, #3668.
Branch: `fm/cli-email-imap-smtp-connector-r1`.

## GSD path

- `scripts/gsd doctor`, command source resolution, and `go run ./cmd/agentcontractgen check`:
  passed before planning.
- Discuss and plan prompts were resolved with `scripts/gsd prompt discuss-phase 3628 --auto` and
  `scripts/gsd prompt plan-phase 3628 --tdd`. Since this issue is absent from the generic ROADMAP,
  execute their required workflow inline and record outputs here rather than mutating ROADMAP/STATE.
- After implementation, resolve and execute `verify-work 3628` and `code-review 3628` inline. Any
  verified gap uses the corresponding `plan-phase --gaps` / `execute-phase --gaps-only` trace.

## Required skills loaded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, `golang-documentation`, `golang-dependency-management`, and
  `golang-pkg-go-dev`.

## Slice A — #3660: contract, evidence, and red checkpoint

1. Create the Email defs bundle with `x-secret` password, closed IMAP/SMTP security enums, and
   port enums. Add the actual IMAP `LIST`/`UID FETCH` and SMTP `MAIL/RCPT/DATA` protocol ledger,
   matching `cli_surface.json` rows for `pm email mailboxes list`, `pm email messages list`, and
   `pm email message send`.
2. RED: add focused bundle/runtime-preflight tests that name the three expected commands and prove
   invalid ports/security reject before credential vault persistence without echoing a supplied
   value. The test must fail before the Email bundle/native registration exists.
3. Add native wiring (and only generated imports through `connectorgen gen`) so `pm email ...`
   resolves to the native executor, never a generic REST path.

## Slice B — #3662: IMAP polled reads and cursor truth

1. RED: add local fake-IMAP tests for mailbox list and bounded message fetch. Assert envelope,
   flags, internal date, RFC822 size, bounded body parts, UID, UIDVALIDITY, and no request beyond
   the requested record/body byte bound.
2. RED: test cursor encode/parse and the read behavior for an existing `(UIDVALIDITY, UID)` lower
   bound, a UIDVALIDITY change/reset, malformed/cross-mailbox state, and the hard-delete limit.
3. GREEN: implement a narrow go-imap v2 adapter plus native reader. Fetch only fixed, declared
   IMAP data; do not expose raw commands, arbitrary search, arbitrary mailbox paths, or IDLE.
   Body parts are bounded and represented safely without hidden truncation.
4. Implement `Check` via authenticated protocol no-op(s), `Catalog`, and `StatefulReader` shape;
   no fake CDC/changefeed capability is declared.

## Slice C — #3664: SMTP typed write and approval binding

1. RED: add local SMTP fake tests for typed recipient/subject/body/attachment validation,
   root-confined attachment reads, deterministic MIME construction, recipient envelope including
   BCC, unmasked preview content, destructive typed confirmation, and no send during preview.
2. GREEN: use standard-library SMTP with explicit TLS/STARTTLS/none selection, strict TLS
   verification, bounded context-aware dial/connection deadline, and no retry. Use the approved
   `go-imap` module only for IMAP; no SMTP module is added.
3. Build `engine.PreparedWrite` from the actual SMTP envelope and full MIME `DATA` bytes. Use the
   established preview warnings/output for that unmasked material and bind it in the preview digest.
   Re-read attachments at execution so preview drift blocks before SMTP sends.
4. Make `send_message` non-batchable, `confirm: destructive`, and reachable only through the
   existing plan → preview → approval → execute path. Never mask payload content or output.

## Slice D — #3666/#3668: docs, parity, and validation

1. Add connector docs with RFC 9051/6409 citations, security/credential setup, cursor semantics,
   hard-delete limitation, IMAP IDLE/#3614 seam, send irreversibility, attachment limits, and fake
   examples that never contain a password.
2. Update applicable CLI and website documentation/generator output following the parity guide.
   Exercise bare `pm email`, topic/help, and each individual command path.
3. Run formatting, focused package tests, runtime implemented-command preflight, bundle validation,
   surface sync, native registration tests, connector boundary, docs/help checks, module integrity,
   and a `pm` build. Measure the `pm` binary-size delta against the pre-dependency baseline.

## Verification plan

- RED/GREEN test commands recorded verbatim in `TDD-LEDGER.md`.
- Focused: `go test ./internal/connectors/native/email`, `go test ./internal/connectors/commandrunner
  -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`, relevant registry/app/CLI packages.
- Contract: `go run ./cmd/connectorgen validate internal/connectors/defs/email`,
  `go run ./cmd/connectorgen surface-sync --check`, and `go run ./cmd/connectorgen gen` followed by
  a clean generated diff.
- Formatting/static: `gofmt`, `go vet` on changed packages, `go build ./cmd/pm`, `go mod verify`,
  `git diff --check`.
- Individual applicable repository gates rather than timeout-prone `go test ./...`/`make verify`:
  tidy-check, lint, docs-check, smoke-no-build only if it does not create/execute a mail write,
  agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, and
  release-workflow-check. CI owns the full suite.

## Safety and non-goals

- No live mail server or actual send; no credential printing/storage in source, fixtures, docs,
  shell output, preview filtering, logs, or tests.
- No `redact_fields`, redacting output policy, generic HTTP/shell/SQL/mail command, IMAP IDLE,
  JMAP, webhook/subscription behavior, generic search, fake REST endpoint, or SMTP read claim.
- If fulfilling a discovered need requires a new dependency, shared engine/schema/runner change, or
  another concurrent lane's owned function, stop and report rather than widening this connector lane.
