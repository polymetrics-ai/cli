# Code Review — Issue #4166 certification proof gaps

**Mode:** inline standard review; the canonical issue-worker contract forbids spawning the GSD reviewer role.

## Verdict

No open critical or warning findings remain in the validation diff. The mandatory real-GitHub Gap 3 proof passes with independent provider read-back and zero residue.

## Review scope

- Full write-sweep probe and both sabotage tests.
- Connector-neutral certification stage plus App-owned issue-label transport proof fixture and absence/unregistration inverses.
- Fresh-binary faithful/live flow round trip, provider client, cleanup guard, refusal matrix, checkpoint assertions, and secret handling.
- GSD context, plan, TDD ledger, verification, UAT, summary, and run state.

## Findings and dispositions

1. **Fixed — warning:** the first write probe loaded the connector certification profile for every action. The immutable config is now resolved once beside the already-cached bundle; 607/607 real engine preparations complete in 1.31 seconds in the focused coverage test.
2. **Fixed — warning:** the first non-applicable transport test assumed a skipped stage has `Passed=true`; repository convention represents skips as `Passed=false` plus `Error=skipped:` and excludes them when computing the overall verdict. The test now asserts the actual explicit-skip contract.
3. **Fixed — warning:** local proof-root and live HTTP response cleanup ignored errors. The proof propagates cleanup failure when no earlier error exists; response close is explicitly bounded/ignored after the body is consumed.
4. **Tracked externally — product defect, no fix here:** provider-verified GitHub 401 is currently classified as typed `internal/internal_error`. The test asserts that observed value, zero provider writes, and unchanged checkpoint with a #4169 comment. Captain assigned the product correction and adjacent HTTP classes to #4169.
5. **Closed — live evidence:** the dedicated private-repository proof passes under `Polymetrics-Cert`. The test owns only `pm-cert-flow-4166-<10 hex>` repositories, cleanup is idempotent, and deletion requires a post-delete 404. The pre-existing #3993 repository is not referenced.
6. **Recorded — runbook drift:** resource-owner labels `polymetrics-cli-cert` and `polymetrics-cli-cert-1` are stale; the live owner is `Polymetrics-Cert`. The first credential remains valid when only its token value is extracted from the labeled line. No credential value is recorded.

## Security and scope review

- Credential values enter only environment-backed credential commands or bounded Authorization headers in the live test client; command output is captured, exact secrets are scanned, and failure output is withheld.
- Approval tokens use stdin, are scanned from the persisted project tree, and are absent from evidence.
- The primary provider mutation is only `comment_issue` in a uniquely created disposable repository. `merge_pull_request` and `delete_file` never execute.
- The chosen action uses the declarative engine fallback, not a GitHub write hook. GitHub-specific adapters are enumerated in `VERIFICATION.md` and `SUMMARY.md`.
- No product fix, connector definition, CLI surface, help/manual, website file, dependency, #4125 code, #4158 code, or #4169 code changed.

## Verification reviewed

Focused Gap 1, Gap 2, App transport, hermetic fresh-binary CLI, and live-GitHub fresh-binary CLI tests pass. Vet, changed-package lint, build, smoke, docs validation, module tidiness, agent contracts, connector validation/surface sync/certification shards/boundary/canon, GitHub parity, and release workflow gates pass. Aggregate `go test ./...` and `make verify` were intentionally not run under the machine-load constraint; CI owns them.
