Refs #4015

## Summary

- Adds `connectorgen certification-evidence report`, a connector-agnostic
  importer from a completed certification report plus its redacted external
  proof into accepted evidence. Connector-owned `evidence_import` metadata
  supplies the provider and matrix identity; shared Go contains no
  connector-name branch or boundary allowlist change.
- Replaces the former PostgreSQL-only provider constant with definition-owned
  metadata. All 14 accepted PostgreSQL records are byte-identical to rebased
  integration head `c9791db4d`.
- Adopts accepted-evidence schema v2 from #4215: a full-parity proof still
  requires the verified full-parity flow, while a direct-read-only proof is
  explicitly `observed_operations` / `protocol_exchanges` and makes no broader
  credential-scope claim.
- Stages GitHub's declaration-owned REST-read binding and a generic declared
  stage selector. Once a fresh run identifies its passing stages, their exact
  names can be selected in the connector definition with no new
  `connectorgen` branch. The same importer is proven with independently
  declared Xero metadata.

## Honest GitHub publication scope

GitHub publishes **0** live-tested capability cells in this PR. This is not a
claim that the 1,571-command gate has closed.

| Status | Count | Reason |
| --- | ---: | --- |
| Published/certified by this PR | 0 | No fresh report-plus-proof artifact is available to import. |
| Staged but unpublished, previously proven live reads | 34 | The generic v2 observed-operation mechanism and declared stage selector are ready, but the previous passes have no recoverable report/proof pair. A fresh bounded read run and its exact definition-owned stage list are deferred while the exposed classic PAT is rotated. |
| Other declared GitHub commands | 1,537 | Outside this PR's publication slice; no execution or evidence is inferred. |
| Declared total | 1,571 | |

The old classic PAT is revoked / do-not-use after an external runbook exposed
it to local terminal output. It is not present in this branch, staged changes,
PR text, or committed evidence, and this PR makes no live call with it. A
credential owner must rotate it before the fresh run. The rule captured here is
that certification runbooks may contain only a secret-store reference,
environment-variable name, or protected key path — never a raw secret.

## Negative controls

- The importer test plants one known value in a request authorization header,
  URL query, and response body; emitted evidence contains none of it and its
  response body is fully fingerprinted.
- A synthetic secret-shaped value in a valid certification definition makes
  `connectorgen validate` fail with `[secret_literal]`; restoring it makes the
  same check pass.
- A capability with no accepted evidence stays not live-tested.
- Deliberately breaking a PostgreSQL CDC evidence key regenerates that cell as
  `live_tested=false, records=0`; restoring it regenerates
  `live_tested=true, records=1`.

## Verification

- `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/certify ./internal/cli`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go run ./cmd/connectorgen certification-candidates --connector github` twice
  (byte-stable)
- `go run ./cmd/connectorgen certification-candidates --connector github --check`
- `go run ./cmd/connectorgen certification-matrix --check`
- `git diff --exit-code c9791db4d -- internal/connectors/certifications/evidence`
- `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` twice
- `pnpm run gen:docs` twice and `pnpm run gen:website-data` twice from `website/`

`make verify` passed after the rebase, including the full Go suite, build,
docs, generation, certification, boundary, canon, and release-workflow gates.
`security/snyk` remains CI-only and has the known identical base-branch
failure; this PR does not mask it.
