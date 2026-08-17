Refs #4015

## Intent

Deliver P1's scalable required-flag correction: a required REST path parameter
must produce a required CLI flag for every connector. This PR deliberately
changes the generator and regenerated artifacts, never GitHub's commands by
hand and never a connector-boundary allowlist.

## What Changed

- Derive and synchronize `flags[].required` from a matching required
  `rest.parameters[]` path declaration for REST operation-backed commands.
- Regenerate GitHub's surface and certification sweep: 104 required-flag
  fields across 92 commands change, taking product defects from 92 to 0.
- Add a repository-wide regression invariant across all 552 bundles, plus a
  focused generator test covering both a missing field and `required: false`.
- Return `commandrunner.MissingRequiredFlagError` so the CLI emits a typed
  `category=usage`, `code=usage_error` refusal without parsing message text.
- Record an audit of all 50 GitHub not-applicable declarations. It finds 26
  `unsupported_api` declarations whose provider operation is present in the
  pinned source lock; they are reported only, not reclassified. The single
  retained `unsupported_api` is `skill list`; all 23 `unsupported_local`
  declarations remain correct.

## Happy, Bad, and Edge Coverage

- **Happy (simulated transport):** the existing GitHub CLI direct-read test
  supplies `--pull-number`, receives and asserts the expected 100-record and
  then 20-record page bodies through the production CLI command path.
- **Bad (simulated transport):** omitting `--pull-number` now exits 2 with
  JSON `category=usage`, `code=usage_error`, and names the required CLI flag;
  the fake provider observes zero requests. The command-runner test also
  asserts the concrete `MissingRequiredFlagError` type before executor I/O.
- **Edge:** a required query parameter remains optional under this deliberately
  path-only derivation rule; the focused fixture asserts it is not changed.

No credentialed live provider call is needed or claimed here: generation and
pre-I/O validation are deterministic. The fake transport proves protocol and
refusal behavior, not a live GitHub response.

## TDD and GSD Evidence

- **Red:** the new all-bundle invariant reported 92 GitHub commands with an
  optional flag mapped to a required REST path; the focused generator fixture
  also failed before derivation.
- **Green:** generic derivation plus repository generation leaves the invariant
  at zero and GitHub's checked sweep at zero product defects.
- GSD `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` prompts were resolved and performed inline. The project
  contract forbids spawning the canonical GSD roles in this environment; the
  fallback and evidence are recorded in the phase artifacts.
- Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, and
  `golang-documentation`.

## CLI/Docs Parity

- `pm github pulls files view --help` renders `--pull-number` as required.
- `pm github releases assets view --help` renders `--asset-id` as required.
- `pm help github` and bare `pm github` both render successfully.
- Connector manuals/skills were regenerated; the docs and website generators
  each ran twice and the second full pass was byte-stable.

## Testing

All local commands passed:

```text
go test -timeout 20m ./cmd/connectorgen
go test -timeout 20m ./internal/connectors/commandrunner
go test -timeout 20m ./internal/cli
go vet ./...
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-runtime-preflight
make connector-canon-check
make connector-boundary
make release-workflow-check
go run ./cmd/connectorgen certification-sweep --connector github --check
```

`go run ./cmd/connectorgen surface-sync`, GitHub certification-sweep,
`pm docs generate`, and `pnpm --dir website run gen:docs` were each run twice;
`git diff --exit-code` passed after the second complete pass. Per repository
guidance, the changed packages plus their consumers were run locally rather
than the timeout-prone aggregate `go test ./...`; CI carries that full suite.

## Safety and Follow-ups

- No credentials, live provider writes, external state changes, or generic
  API/write tools were used.
- This change contains no connector-specific source identifier and no boundary
  allowlist change; `make connector-boundary` passes.
- The 26 contradicted `unsupported_api` declarations need a separately owned
  truthful availability contract and fixture/entitlement work; this PR leaves
  all classifications untouched as requested.

## Review and Pipeline

- Commit checkpoint: `7493b63ad` planning, then `eb9413221` green
  implementation/generation; both pushed to
  `fm/cli-required-flag-derivation-r1`.
- Manual code review recorded in
  `.planning/phases/cli-required-flag-derivation-r1/REVIEW.md`: no unresolved
  findings.
- Automatic Claude review is the primary route and is pending on PR open; no
  manual review request has been made. Copilot is not requested unless that
  primary route becomes unavailable.
