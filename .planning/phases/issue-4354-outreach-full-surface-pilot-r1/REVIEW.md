# Code review — #4354 (inline manual fallback)

## Scope review

- Candidate comparison against `060bb7864e3419e09ab10e000bb14ac1ea3724ec` identified exactly seven Outreach-owned artifact paths. The implementation imports only those artifacts, adds `outreach-six-lane-evidence.json`, a narrow Outreach binary test, and required GSD evidence.
- No Batch 6–7 connector other than Outreach is present in the diff. No shared source reader, validator, generator, schema, certification matrix, or source lock parser is changed.
- The binary test's credential is a synthetic environment value and its only HTTP server is `httptest.NewServer`; it has no provider hostname or real secret input.

## Behavioral review

- `TestPMBinaryOutreachRepresentativeCommandsReachCredentialBoundary` checks ETL, typed write, and destructive delete paths through the real binary and requires `missing --credential` for each.
- `TestPMBinaryOutreachFixtureBindsDeclaredMethodAndPath` rejects caller `--method`, `--path`, and `--source-url` flags before any request, then observes exactly `GET /api/v2/prospects` on the local fixture.
- The connector remains `COMMUNITY BUILD, UNCERTIFIED`; source mapping, runtime preflight, live/hash certification, and provider success are recorded as distinct states.
- Full CLI regression first found two generated parity artifacts missing from the candidate import. The normal skill generator added `docs/skills/pm-outreach/SKILL.md`; the supported targeted golden update added Outreach to the root manual fixture. Their focused currentness tests pass.

## Final local-gate record

- `gofmt -w internal/cli/outreach_full_surface_pilot_test.go` and `git diff --check`: pass.
- `go build -o /tmp/pm-outreach-pilot ./cmd/pm`: pass.
- `go vet ./internal/cli ./internal/connectors/commandrunner ./cmd/connectorgen`: pass.
- `go test -timeout 20m -v ./internal/cli -run 'TestPMBinaryOutreach(RepresentativeCommandsReachCredentialBoundary|FixtureBindsDeclaredMethodAndPath)$' -count=1`: pass (44.122s).
- `go test -timeout 20m -v ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`: pass (14.906s).
- `go test -timeout 20m ./internal/cli`: complete; the final cache confirmation returned `ok   polymetrics.ai/internal/cli (cached)`.
- `go run ./cmd/agentcontractgen check`: pass.
- Independent detached worktree (`cli-outreach-pilot-audit-r1` at `1d64e22ce`): `go build -o /tmp/pm-outreach-pilot-audit ./cmd/pm`, an initialized empty project, representative ETL/write/delete commands all stopped at `missing --credential`, and the focused binary test passed in 43.036s.
- Expected blocked checks are unchanged: `connectorgen validate …outreach --json` rejects `source_kind`; `operation-evidence . --check` reports global generated-artifact drift; certification matrix says Outreach is not allowlisted; certification sweep says its generated file is missing.

## Findings / disposition

1. **Blocking — source admission:** strict source import rejects the candidate lock's `source_kind`, and the candidate carries neither canonical source descriptor nor retained source artifacts. Firstmate assigned shared reader/citation work to PR #4350. This PR deliberately retains all rows and does not attempt a competing repair.
2. **Blocking — generated evidence:** `connectorgen operation-evidence . --check` reports drift. The generic projection is not regenerated or claimed shipped.
3. **Blocking — website parity:** `website/data/connectors.generated.json` has an Outreach connector record with `cliSurface: null`; there is no Outreach-specific hand-authored `docs/cli` or website page. The required global generator output cannot truthfully be refreshed until source admission succeeds, so this is not marked not-applicable.
4. **Non-blocking runtime proof:** focused commandrunner and new binary tests pass. These do not erase the three blocking gates above.
5. **Follow-up sequencing:** Firstmate message `002.msg` requires this clean pilot to be committed and draft-only. Only after PR #4350's exact repaired head receives CI and independent-audit acceptance may a successor locally integrate it and rerun source admission plus ETL/write/delete credential-boundary proofs.

## Verdict

The change is appropriately scoped and the local command boundary is proven, but this draft is **not eligible for independent audit approval or the captain-authorized pilot merge** until the listed source/evidence/website dependencies are closed and all required checks are rerun green.
