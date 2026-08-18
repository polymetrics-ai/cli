# Verification — GitHub slice 4 live certification

## Commands and results

- `scripts/gsd doctor` — passed; repository-local GSD adapter and command
  registry are available.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`,
  `verify-work`, and `code-review` — resolved to the installed GSD command
  sources.
- `go run ./cmd/agentcontractgen check` — passed; canonical contract and
  registered projections are current.
- `go run ./cmd/connectorgen certification-matrix --check` — passed:
  certification shards are current.
- `go test -timeout 20m ./cmd/connectorgen` — passed.
- `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` — passed;
  the evidence-record changes have the required GSD/TDD planning artifacts.

## Live proof and cleanup

Live read commands used the disposable classic credential and, when an endpoint
returned an access-shaped 403/404, one bounded fine-grained retry before honest
classification. Each passing response was asserted by structure or an
identifier-bearing produced value; exit status alone was not evidence.

Direct provider read-backs proved cleanup by returning zero for:
`/user/keys`, `/user/ssh_signing_keys`, `/user/gpg_keys`,
`/user/social_accounts`, `/user/blocks`, and `/user/following`.

## Review

The changed production paths are evidence records only. No connector runtime,
CLI surface, generated artifact, or mutation path was changed.
