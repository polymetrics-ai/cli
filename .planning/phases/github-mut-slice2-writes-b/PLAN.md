## Task Delivery Header

- Issue: Refs #4015 — GitHub mutation certification, slice 2 writes-b.
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: A PR from `fm/cli-mut-slice2-writes-b` is open against the stated base, with one honest classification for each of the 146 assigned commands.
- Working branch: fm/cli-mut-slice2-writes-b
- Task: Execute the contained GitHub write slice serially, perform the enforced plan → preview → token-stdin run path, independently read back each claimed state change, direct-provider-delete every fixture, and retain only schema-v2 evidence that validates.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`; targeted certification/CLI tests; `git diff --check`; inspect the opened PR base through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Each assigned command has exactly one outcome | live | The committed batch ledger accounts for 146 unique command paths and the bucket total is 146. |
| A certified mutation changed provider state and was cleaned up | live | Independent provider read-back observes the produced state, then a direct DELETE followed by independent 404/empty read-back yields `verified_absent`. |
| No secret is persisted | live | Published proof contains only repository-salted fingerprints; credential values are supplied from Keychain via an environment variable at point of use. |
| Evidence artifacts validate | live | `connectorgen certification-matrix --check` accepts every retained record. |

## Execution plan

Inline/manual GSD fallback: the task explicitly prohibits role spawning, so the generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts are executed by this worker. Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

1. Create an isolated local pm project and credential using Keychain-to-environment import only. Read each command manual before execution and use the declared fixture scope.
2. For every command: create/read a `pm-cert-` fixture, issue a connector-command plan, preview it, supply the one-time token only through the bare `--approval-token-stdin` marker, and independently prove the provider-side result.
3. Destroy each owned fixture by direct `api.github.com` DELETE and independently prove absence. Record `certified` only after that proof; classify all non-passes separately and obtain a raw GitHub control for every product defect.
4. Write/validate accepted schema-v2 evidence after each certified result. Do not regenerate shared artifacts.
5. Run targeted checks, perform inline code review, commit, push, open the stacked PR, and read back its API-reported base.

## CLI help/docs/website parity

Not applicable: this certification lane does not alter the CLI surface, help, manual, generated documentation, or website; runtime command manuals are used as execution evidence only.
