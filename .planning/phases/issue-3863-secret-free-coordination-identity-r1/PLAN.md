# PLAN — issue #3863 secret-free credential coordination identity

Issue: #3863. Parent: #3862. Branch: `fm/cli-found-credential-coordination-identity-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `go run ./cmd/agentcontractgen check`: passed.
- Discuss prompt: `scripts/gsd prompt discuss-phase 3863`; executed inline and recorded in
  `CONTEXT.md` / `DISCUSSION-LOG.md`.
- Plan prompt: `scripts/gsd prompt plan-phase 3863 --tdd`; executed inline in this plan.
- Execute prompt: `scripts/gsd prompt execute-phase 3863`; executed inline under this plan.
  Verify prompt: `scripts/gsd prompt verify-work 3863`; executed inline with automated coverage.
  Review prompt: `scripts/gsd prompt code-review 3863 --files=...`; executed inline at standard
  depth and recorded in `REVIEW.md`. If verification finds a real gap, use `plan-phase 3863 --gaps` followed by
  `execute-phase 3863 --gaps-only` before rerunning verification.
- Inline/manual fallback: compatible isolated GSD roles are unavailable and the canonical
  single-worker contract forbids role spawning. No TDD, verification, review, or human gate is
  weakened.

## Required skills loaded

- `golang-how-to` — routed the Go identity/CLI/test work.
- `golang-design-patterns` and `golang-structs-interfaces` — one small typed builder and runtime
  contract rather than two ad-hoc keys.
- `golang-error-handling`, `golang-safety`, and `golang-security` — fail closed, avoid secret
  inputs/output, validate metadata without echoing values, and preserve safe state behavior.
- `golang-testing` — red/green observable contracts with no live provider or credential.
- `golang-cli` and `golang-documentation` — safe credential-link flags, help, manual, website, and
  machine-readable output parity.

## Slice A — identity contract and RED evidence

1. RED: add focused table-driven tests for a typed `CoordinationIdentity` builder. Prove an
   explicitly shared binding yields one opaque auth-cohort projection; rate keys match only when
   policy, kind, and subject match; distinct non-secret subjects split the budget even under the
   same binding; and no rate scope appears without a declaration.
2. RED: prove malformed/incomplete policy scopes and unsupported kinds are rejected without an
   accidental fallback. Assert identity serialization and error paths cannot contain the binding
   preimage. The builder input contains no secret/revision field, so derivation can be proven pure
   and vault-free.
3. Run the new package test before implementation, retain the failure, and record it in
   `TDD-LEDGER.md`.

## Slice B — protected credential metadata and runtime handoff

1. GREEN: add a project-salted, domain-separated identity implementation to `internal/connectors`
   and add its opaque value to `RuntimeConfig`. Do not alter `CredentialRevision`, connector bundle
   schemas, command runner, requester, or registry behavior.
2. RED/GREEN: add protected state storage for binding IDs plus public non-secret
   provider-family/auth-profile credential metadata. Generate an isolated binding at creation;
   migrate old records on open without vault reads; construct runtime identity before reading the
   vault; pass only the opaque identity onward.
3. RED/GREEN: add programmatic explicit linking with exact compatibility checks. Test linked
   cross-connector credentials, unlinked copies, mismatched family/profile, and an invalid link
   target. Rotation evidence stays independent: the same binding retains cohort/rate identity while
   approval revision remains its existing separate mechanism.

## Slice C — safe credential CLI and rendered parity

1. RED/GREEN: expose non-secret `--provider-family`, `--auth-profile`, and safe named-link inputs;
   add an existing-credential link action. Validate every new metadata value before vault/state
   persistence with errors naming only the field/constraint, never supplied values.
2. Test `pm credentials`, `pm help credentials`, and `pm credentials --help`, valid linking,
   incompatible cross-connector rejection, and JSON/inspect output. Confirm no binding preimage,
   opaque projection, secret/revision material, or rate key is rendered.
3. Update embedded help, generated CLI manual, website CLI reference, and generated website data;
   test bare `pm credentials` namespace help remains successful.

## Slice D — regression, verification, and review

1. Run focused identity, app, and CLI tests, then their package suites. Run the existing runtime
   executable-command preflight sweep as a no-regression guard; no declaration is added here.
2. Run format, vet/build, docs/website generation/validation, all applicable individual repository
   gates, and `git diff --check`. Do not use a live provider, raw credential, reverse-ETL execution,
   `go test ./...`, or the timeout-prone `make verify` monolith locally.
3. Execute GSD verify and code-review prompts inline. Record every review finding/disposition in
   `REVIEW.md`; use gap planning only for real uncovered requirements.

## Verification plan

- RED/GREEN: targeted identity and app tests, then `go test ./internal/connectors` and
  `go test ./internal/app`.
- CLI: `go test ./internal/cli`; `go build ./cmd/pm`; built binary `pm credentials`,
  `pm help credentials`, and `pm credentials --help` against a temporary empty root only.
- Formatting/static: `gofmt -w` only changed Go files; `go vet ./internal/connectors`,
  `go vet ./internal/app`, `go vet ./internal/cli`; `git diff --check`.
- Docs/website: generated `docs/cli/credentials.md`; website data regeneration; targeted website
  docs/typecheck or project helper if available; grep the three safe flags across help/manual/site.
- Individual repository gates: `make tidy-check`, `make lint`, `make docs-check`,
  `make smoke-no-build` only if it does not require a credentialed/provider or reverse-write run,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check`.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. RED identity/credential/CLI tests checkpoint, retaining observed failure.
3. GREEN implementation and rendered-parity checkpoint.
4. Verification/review documentation checkpoint.

## Safety and non-goals

- No secret value, secret-derived equality, binding preimage, credential revision, or raw policy
  subject is emitted to ordinary output, logs, or coordination state. No generic HTTP/SQL/shell
  surface, provider call, registry, fence, parking state,
  scheduler, or capability declaration is added.
- A rate-scope projection is impossible without an explicit supported scope declaration. Rotation
  invalidates approval evidence only; it cannot reset an account rate budget or join/leave a cohort.
- `internal/connectors/commandrunner/runner.go`, shared connector JSON schema, engine/requester
  policy logic, and downstream coordination work are immutable for this slice. Report rather than
  absorb a need to change them.
