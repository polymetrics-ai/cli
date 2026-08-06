# PLAN — issue #3614 webhook receiver exposure modes

Branch: `fm/cli-found-webhook-exposure-modes-r1`. Stack base:
`fm/cli-found-database-sync-contract-r1` / issue #3810.

## GSD path and required skills

- Passed: `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`,
  `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; and
  `go run ./cmd/agentcontractgen check`.
- Prompts generated with `scripts/gsd prompt <command> 3614`; executed inline
  because canonical single-worker delivery does not permit role spawning here.
- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, and `golang-documentation`.

## Slice A — exposure declaration and lifecycle (TDD)

1. RED: specify the closed three-mode configuration. Prove HTTPS-only opaque
   endpoint generation, named-Tailscale-only external tunnel input,
   no-listener operator/pull-stream modes, and loopback-only tunnel listener
   behavior.
2. GREEN: add a small receiver package with mode validation, URL
   fingerprinting that retains no callback URL, lifecycle state, heartbeat
   expiry, endpoint-generation rotation degradation, and typed #3810 recovery
   requests. Consume `connectors.CoordinationIdentity`; never use a credential
   revision or raw credential.
3. RED/GREEN: persist the minimum mode/subscription state in project state and
   expose a safe operator status view. Explicit re-registration/reconciliation
   actions may clear only their request markers, never claim that a provider
   mutation occurred.

## Slice B — bounded, durable HTTP ingress (TDD)

1. RED: cover raw-body signature verifier ordering, failed verification,
   durable-receipt failure, duplicate receipt, out-of-order arrival, overload,
   and oversized bodies.
2. GREEN: implement an injectable receiver bound to `127.0.0.1`/`::1`, with
   method/path matching, `http.MaxBytesReader`, deadlines, and an explicit
   in-flight bound. Verify raw bounded bytes before JSON/event interpretation.
3. GREEN: atomically persist a minimized durable receipt before HTTP success.
   Existing durable duplicates acknowledge without redispatch; new writes that
   cannot persist or enter bounded hand-off return retryable failure.

## Slice C — CLI/help/docs truth surface (TDD)

1. RED/GREEN: add a typed `pm` receiver namespace/configuration interface with
   `--json` safe status. It must state active mode, generation (opaque),
   listener behavior, degradation, and no provider-registration claim. It
   cannot print a callback URL, signing secret, or raw body.
2. Update runtime help, bare namespace behavior, CLI manual and website docs;
   document `tailscale_funnel` as external operator setup and
   `provider_pull_or_stream` as non-webhook.
3. Record the local Funnel version/DNS/port evidence without publishing a
   secret path. If safe constraints permit, prove a real port-10000 Funnel to
   a `pm` loopback receiver without touching the existing mappings.

## Slice D — verification and review

1. Run targeted receiver/app/CLI tests plus package suites, race checks for
   receiver concurrency, build, vet, formatting, diff check, and the individual
   repository gates named by `AGENTS.md`.
2. Verify `pm help`, bare namespace, command help, docs and website parity.
   Do not run a provider call, credentialed test, reverse ETL, generic poller,
   `go test ./...`, or monolithic `make verify` locally.
3. Execute verify-work and standard-depth code review inline. If an acceptance
   gap remains, run `plan-phase 3614 --gaps`, then
   `execute-phase 3614 --gaps-only` before re-verifying.

## Safety and immutable boundaries

- No new Go module or dependency. No Tailscale invocation from `pm`, no
  generic shell/HTTP/SQL write surface, and no provider subscription mutation.
- #3810 owns checkpoint/recovery semantics; #3855 owns polling; #3862 owns
  transport seams; provider lanes own registration, verifier algorithms, and
  provider event identities.
- No callback URL with a possible secret path, signing secret, raw credential,
  signature header, or raw event body appears in state output, logs, tests, or
  docs. The persisted public state uses only opaque generations/digests.

## Commit checkpoints

1. Planning artifacts and dependency stack base.
2. RED receiver/mode tests.
3. GREEN receiver/lifecycle/durable receipt implementation.
4. CLI/docs/Funnel evidence and final review/verification artifacts.
