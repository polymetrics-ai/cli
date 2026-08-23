# Verification checklist — issue #4299 definition embed slim

Status: Verify CI toolchain remediation locally complete; post-push latest-run
monitoring and the documented independent-review handoff remain required.

## Verify CI toolchain remediation — completed local execution

GitHub Verify run `32321756934` failed after reaching the real
`release-installed-github-certification.sh` gate. The test builds and assembles
the host `linux/amd64` artifact; Linux target selection deliberately also
builds its deb/rpm packages, so the assembler correctly refuses to proceed
without `nfpm` on `PATH`. The existing Verify job ran `make verify` without
provisioning that pinned tool. The repair must be a red workflow-contract test
followed by the repository-pinned `nfpm@v2.43.0` setup in that owning job, not a
target substitution, synthetic archive, or relaxed gate. Results will be
recorded below after execution.

Completed focused red/green evidence:

```text
Red:   bash scripts/tests/verify-release-tooling.sh
       verify release tooling check failed: Verify job must provision pinned nfpm before make verify
Green: bash scripts/tests/verify-release-tooling.sh
       verify release tooling: nfpm is provisioned in the owning Verify job
Green: make release-workflow-check
       release size budget guard passed
       release production layout passed
       installed GitHub certification archive proof passed
```

The workflow green step installs exactly `nfpm@v2.43.0`, exposes `GOPATH/bin`
through `GITHUB_PATH` before `make verify`, and directly executes that installed
binary. The workflow-contract test is included in `release-workflow-check`, so
a future removal, unpinned version, missing PATH export, or reordering after
`make verify` fails before the production release proof can silently rely on an
ambient runner tool.

Post-fix repository validation on 2026-08-20:

```text
go test -timeout 20m ./...                         PASS
go vet ./...                                       PASS
make release-workflow-check                        PASS
make lint                                           PASS (0 issues)
make docs-check                                     PASS
make tidy-check                                     PASS
make smoke-no-build                                 PASS
make agent-contract-check                           PASS
make connectorgen-validate                          PASS (552 connectors, 0 findings)
make connectorgen-surface-sync                      PASS (552 connectors, 0 changes)
make github-parity-artifacts-check                  PASS
make connectorgen-certification-matrix              PASS
make connectorgen-certification-candidates          PASS
make connectorgen-certification-sweep               PASS
make connector-runtime-preflight                    PASS
make connector-boundary                             PASS (clean; 293 files, 552 connectors)
make connector-canon-check                          PASS
scripts/verify-gsd-workflow origin/main             PASS
```

Changed-workflow validation is layered: the new workflow-contract test parses
the actual `verify.yml` job and is run both directly and through
`release-workflow-check`; the existing pinned-build-dependencies validator also
passed from that gate. No credentials, live provider request, altered source
lock, connector definition, relaxed release target, fake archive, or ambient
`nfpm` dependency was used.

## Inline GSD code review — 2026-08-20

The generated `code-review` prompt was executed inline because the repository
single-worker contract forbids reviewer-role spawning. Reviewed the workflow,
Makefile wiring, new workflow-contract test, and evidence records together:

- The job-local `nfpm@v2.43.0` installation precedes the only `make verify`
  step, exports `GOPATH/bin` through GitHub's supported `GITHUB_PATH` channel,
  and directly invokes the installed binary before the downstream step can use
  it.
- The regression test fails closed for missing setup, an unpinned install,
  missing PATH publication, absent direct executable confirmation, or setup
  ordered after `make verify`; it is itself wired into the release gate.
- The actual host Linux archive continues to assemble packages and execute the
  extracted binary outside the checkout. No failure is hidden by a fallback,
  skip, allow-failure setting, synthetic asset, or ambient tool.
- The raw GitHub exception, connector definitions, operation ledgers, command
  surface, result behavior, and source locks are not in this repair diff.

Result: no actionable finding.

Remediation gates completed locally on 2026-08-20:

```text
go test -timeout 20m ./...                         PASS
go vet ./...                                       PASS
go build ./cmd/pm                                  PASS
make release-workflow-check                        PASS
make tidy-check; make lint; make docs-check         PASS
make smoke-no-build; make agent-contract-check      PASS
make connectorgen-validate                          PASS (552 connectors, 0 findings)
make connectorgen-surface-sync                      PASS (552 connectors, 0 changes)
make github-parity-artifacts-check                  PASS
make connectorgen-certification-{matrix,candidates,sweep}  PASS
make connector-boundary                             PASS (clean; 552 connectors)
make connector-canon-check                          PASS
scripts/verify-gsd-workflow origin/main             PASS
```

## Required behavior

- [x] Real `defs.FS` inventory is sorted, attributed, and rejects `api_surface.json`, `fixtures/**`, and every source lock except the explicit GitHub exception.
- [x] GitHub source-lock literal bytes and SHA-256 match the committed raw file.
- [x] `go test -timeout 20m ./internal/connectors/certify` proves GitHub GraphQL schema certification remains available offline.
- [x] A real assembled archive is extracted to a directory outside the checkout; its installed `pm` initializes a new project and reaches the GitHub full-certification GraphQL boundary with 29 schema-conformant, 2 live-required, and 274 fixture-bound commands. The credential-free full certificate remains non-passing, as expected.
- [x] Current rebased release-like before/after measurements report identical commands, build metadata, byte sizes, archive sizes, and deltas in [MEASUREMENTS.md](MEASUREMENTS.md).

Focused proof:

```text
go test -timeout 20m ./internal/connectors/defs ./internal/connectors/certify -count=1  PASS
scripts/tests/release-size-budget.sh                                                   PASS
```

`TestGithubFullGraphQLInventoryUsesEmbeddedSourceLockOutsideCheckout` is an
internal-function test only: it changes working directory, then compiles
GitHub's full GraphQL inventory from `defs.FS`. It does not by itself prove an
installed binary or archive. That runtime claim is now covered separately by
`scripts/tests/release-installed-github-certification.sh`, which extracts a
real assembler-produced host archive, invokes its `pm` from an initialized
temporary project, accepts the expected nonzero credential-free certificate,
and parses the JSON report for the passed GraphQL stage and `29/2/274` counts.
It uses no credentials and makes no provider request.

## Review remediation focused proof

- [x] Red: `bash scripts/tests/release-production-layout.sh` against the
  reviewed head failed because the actual GNU-tar producer listed `./pm` rather
  than `pm`.
- [x] Green: `bash scripts/tests/release-production-layout.sh` passed after the
  assembler began archiving the explicit root entries. The test invokes the
  real selected-target verifier, which in turn invokes the real size-budget
  guard, and rejects root binaries accompanied by `nested/pm` and `../pm`.
- [x] `bash scripts/tests/release-size-budget.sh` passed. It covers deterministic
  reports, quiet success/failure behavior, oversized archive/installed binary,
  missing/duplicate root binaries, and nested/traversal impostors.
- [x] `scripts/tests/release-installed-github-certification.sh` passed. Its
  installed binary certificate had the expected nonzero credential-free exit,
  emitted no stderr, and its JSON proof asserted GraphQL `29/2/274` plus a
  passed `graphql_schema_conformance` stage.
- [x] `scripts/tests/release-target-parity.sh` passed; the four supported
  targets remain darwin/amd64, darwin/arm64, linux/amd64, and linux/arm64.
- [x] `scripts/verify-release-assets.sh --release-version 0.0.0-snapshot
  --print-expected-release-assets --targets linux/amd64,linux/arm64` produced
  only the selected Linux archive/package subjects and their bundles; an
  unsupported Windows target failed closed.

The PR `package-check` job now runs the real release verifier against the two
Linux archives and packages it assembled before QEMU installation tests. The
job installs `rpm` first because package metadata verification is part of that
real verifier; this is not a reduced or synthetic check.

## Repository gates

- [x] `go test -timeout 20m ./internal/connectors/defs ./internal/connectors/certify ./internal/cli` (the complete suite later covered these again).
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connectors, 0 changes.
- [x] `go run ./cmd/connectorgen certification-matrix --check`, `certification-candidates --connector github --check`, and `certification-sweep --connector github --check`.
- [x] `make connector-boundary` — detached and polled; clean, 552 connectors, 0 findings.
- [x] `make lint`, `make docs-check`, `make smoke-no-build`, `make tidy-check`, `make agent-contract-check`, and `make release-workflow-check`.
- [x] `make connector-runtime-preflight`, `make connector-canon-check`, and `make github-parity-artifacts-check`.
- [x] `go test -timeout 20m ./...` — exit 0.
- [x] `scripts/verify-gsd-workflow origin/main`.

The repository's `make verify` target is the composition of the checked
commands above. Its long test and boundary members were run with tracked
20-minute sessions and the remaining targets were run independently, as the
repository agent guidance requires under per-command time limits.

## Manual code review

Reviewed commit `a2a3e4a51` against `origin/main`: the exception is a literal
path (not a new glob), the guard validates both compressed archive and streamed
installed-binary bytes, `--print-subjects` stays machine-readable via `--quiet`,
and no connector definition, source lock, compression, or checkout fallback is
in the diff. No local finding.

Automated review route: `claude_auto` after the non-draft main-targeted PR is
opened by this trusted repository worker. No manual Claude or Copilot request
is appropriate before the automatic trigger is observed.

## Delivery

- [x] Commit messages include `Refs #4299`.
- [x] PR body includes `Refs #4299`.
- [x] Rebased against `origin/main` immediately before the branch push; no force-push or merge.
- [x] Opened [PR #4309](https://github.com/polymetrics-ai/cli/pull/4309), Conventional Commit title, targeting `main`, with measurement, inventory, TDD/GSD/skill, exception, and exclusion evidence.
- [x] Verified the GitHub API-reported base: `gh-axi pr list --state open --head fm/cli-defs-embed-slim-r1 --base main` returned exactly PR #4309. Head at PR open: `d5721f66fe6b67dd64624c076379042f255dbc69`.
