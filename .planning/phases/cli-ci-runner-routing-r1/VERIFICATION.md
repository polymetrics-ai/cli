# Verification — CI runner routing

Status: passed locally; full CI and the no-mistakes shipping pipeline are
reserved for a later firstmate instruction.

## Passed

- `./scripts/tests/verify-ci-runner-routing.sh` — proves the single selector,
  exact trust conditions, GitHub-hosted fallback, 19 dynamic Linux consumers,
  persistent GitHub configuration instructions, Claude selector eligibility,
  Windows exception, and website deployment exception.
- Ruby YAML parse of every `.github/workflows/*.yml` file.
- `make release-workflow-check` — passed after its release dependency assertion
  was strengthened for the selector plus release-assets dependency.
- `go run ./cmd/agentcontractgen check` — passed.
- `git diff --check` — passed.

`actionlint` is not installed locally; no dependency was installed for this
workflow-only change. The repository's GitHub Actions validation remains part
of the later CI gate.

## Review

Manual source review found no secret material or server/deployment paths. The
selector's self-hosted result is the known online `polymetrics-cli` label; it
checks same-repository and explicit-author conditions before returning that
label. Those expressions are routing policy, not a durable fork-security
boundary, so `docs/security/self-hosted-ci-runner-policy.md` requires GitHub
fork approval and dedicated runner-group scoping before routing is safe. All
other event types return `ubuntu-latest` deliberately. The Windows workflow is
untouched, and the website deployment job retains its separate trusted-ref
runner exception.
