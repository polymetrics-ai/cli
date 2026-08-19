Refs #4015
Related: #4211

## Intent

Measure, rather than assume, end-to-end GitHub direct-read certification throughput: 10 operations, then a fresh 100, then two additional fresh 100-operation batches.

## What Changed

- Added a reproducible, serial live-read batch harness and a safe failure-replay helper under the issue phase evidence.
- Recorded real setup-to-teardown timing, produced-value assertions, checkpoint/resume behavior, safe rate-limit observations, the scaling projection, and agent rulebook rules.
- Staged a bounded evidence payload without a credential-scope claim.
- Recorded 14 distinct product defects; these are not provider refusals.

## Result

The three 100-operation runs average **154.831 s / 1.548 s per target operation** (150.351–158.521 s range). No limiter wait or not-sent event appeared in 310 timed target stages; the measured curve is flat, with fixed setup dominating the 10-operation sample.

The current ledger has 639 implemented direct reads, projecting to **16m30s** serial lifecycle time from the measured 100-operation curve. This is a measured-time projection, not a percentage estimate.

PR #4215 merged while the run was active, and this branch is rebased on it. The landed scope contract is now correct. Generic GitHub live-evidence import remains in open PR #4216, so this PR deliberately does not hand-author accepted evidence or claim full parity for a partial direct-read sample.

## Red / Green

- **Red:** certification refused a harness missing the declared non-secret `rate_limit_account` before any provider request.
- **Green:** after supplying the declared account, 10 and three fresh 100-operation read batches completed with assertion-bearing produced-value passes and distinct non-pass categories.
- **Refactor:** no runtime, connector bundle, or generated surface change; generated surfaces were rerun twice and remained byte-stable.

## GSD and skills

- Lifecycle resolved and executed inline: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`; Pi role spawning is forbidden by the delivery contract, and the fallback is recorded in phase artifacts.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`.

## Testing

- `go vet ./...`
- `go test -timeout 20m ./cmd/connectorgen`
- `go test -timeout 20m ./internal/connectors/certify`
- `go test -timeout 20m ./internal/cli`
- `go build ./cmd/pm`
- Individual `make verify` non-test gates, generated-surface checks, shell/JSON checks, and final workflow verification; details in phase `VERIFICATION.md`.

The full `make verify` / 550+ connector suite was not invoked as one worktree command, per repository guidance; CI owns that full-suite pass. `security/snyk` is a documented identical base-branch failure.

## Safety and follow-up

- Reads only; no GitHub mutations, repository writes, or fixture creation.
- The disposable credential was environment-only and never logged or committed.
- Follow-up: merge PR #4216, then import `STAGED-LIVE-REPORT.json` through its generic importer under PR #4215's verified scope contract. Do not create a manual accepted record.

## Automated review

- Route: `claude_auto` on PR open; no manual Claude/Copilot request.
- Status: pending creation. Any review findings will be dispositioned before merge.
