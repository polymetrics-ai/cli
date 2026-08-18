---
status: passed
phase: issue-4015-github-slice5-repo-content
verified: 2026-08-18
---

# Issue #4015 — GitHub slice 5 verification

## Outcome

- Branch-owned certified records: 25 (23 direct reads and 2 ETL reads).
- Shared-base records are not counted as this branch's contribution.
- Other slice outcomes: 19 shared-base-covered, 33 `no_object`, 8 `entitlement`, 7 `product_defect`, and 82 mutations deferred by captain direction; total 174.
- Provider cleanup proof: zero remaining `pm-cert-slice5` labels, environments, deployments, and codespaces.

## Exact verification commands and results

| Command | Result |
| --- | --- |
| `git diff --name-only origin/integration/4015-mvp-flat-r1...HEAD -- internal/connectors/certifications/evidence` | PASS — exactly 25 branch-owned evidence files |
| `for f in $(git diff --name-only origin/integration/4015-mvp-flat-r1...HEAD -- internal/connectors/certifications/evidence); do jq -e '.status == "passed"' "$f"; done` | PASS — 25/25 records report `status: passed` |
| `jq empty internal/connectors/certifications/evidence/github-manual-slice5-*-agent-derived-20260818.json` | PASS |
| `go run ./cmd/connectorgen certification-matrix --check` | PASS — `certification shards are current` |
| `go run ./cmd/agentcontractgen check` | PASS |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | PASS |
| `go run ./cmd/connectorgen surface-sync --check` | PASS |
| `go run ./cmd/connectorgen boundary . --json` | PASS |
| `go test -timeout 20m ./cmd/connectorgen ./internal/agentcontract ./internal/connectors/certify` | PASS |
| `go test -timeout 20m ./internal/cli` | PASS |
| `go vet ./...` | PASS |
| `go mod tidy && git diff --exit-code -- go.mod go.sum` | PASS — no module drift |
| `go build -o /tmp/pm-cert-slice5-verified ./cmd/pm` | PASS |
| `scripts/tests/pinned-build-dependencies.sh` | PASS |
| `scripts/tests/homebrew-release-notify.sh` | PASS |
| `scripts/tests/release-target-parity.sh` | PASS |
| `git diff --check origin/integration/4015-mvp-flat-r1...HEAD` | PASS |

Thirteen agent-derived assertion negative controls passed: each rejected a plausible wrong produced value. A credential/private-key marker scan over the 25 branch records found no secret material.

## Cleanup read-backs

The live run used direct GitHub collection GETs and counted only objects with the `pm-cert-slice5` prefix. Authorization was supplied from Keychain at point of use and is intentionally represented as `<keychain-token>` below rather than stored in this artifact.

```bash
curl --silent --show-error --fail-with-body \
  -H 'Authorization: Bearer <keychain-token>' \
  -H 'Accept: application/vnd.github+json' \
  'https://api.github.com/repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru/labels?per_page=100' \
  | jq '[.[] | select(.name | startswith("pm-cert-slice5"))] | length'
# 0

curl --silent --show-error --fail-with-body \
  -H 'Authorization: Bearer <keychain-token>' \
  -H 'Accept: application/vnd.github+json' \
  'https://api.github.com/repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru/environments?per_page=100' \
  | jq '[.environments[] | select(.name | startswith("pm-cert-slice5"))] | length'
# 0

curl --silent --show-error --fail-with-body \
  -H 'Authorization: Bearer <keychain-token>' \
  -H 'Accept: application/vnd.github+json' \
  'https://api.github.com/repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru/deployments?per_page=100' \
  | jq '[.[] | select((.environment // "") | startswith("pm-cert-slice5"))] | length'
# 0

curl --silent --show-error --fail-with-body \
  -H 'Authorization: Bearer <keychain-token>' \
  -H 'Accept: application/vnd.github+json' \
  'https://api.github.com/user/codespaces?per_page=100' \
  | jq '[.codespaces[] | select(.name | startswith("pm-cert-slice5"))] | length'
# 0
```

These provider read-backs are the cleanup proof. No connector delete exit status was treated as proof of absence.

## Product-defect controls

Seven product defects were retained with raw `api.github.com` controls: unsupported invented `page=1` on two Dependabot alert endpoints, two valid HTTP 204 responses rejected as JSON EOF, valid `main...main` rejected as path traversal, and incorrect completeness metadata for the participation and punch-card statistics endpoints.

## Deliberately deferred aggregate gates

Captain integration decision: `go test -timeout 20m ./...` and `make verify` are deferred, not skipped, to one integration pass after all six lanes land. Running shared generated-artifact gates independently in every lane would create six-way conflicts. This lane ran the targeted package tests and non-regenerating gates listed above.

## GSD workflow evidence

- Manual-GSD/inline fallback: the canonical single-worker lane prohibited role spawning; the adapter sources were resolved and the lifecycle was executed inline.
- Red: `bash scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` exited 1 because this branch lacked a planning artifact.
- Green: the same command passes with this phase record present.
