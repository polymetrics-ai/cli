# Verification checklist — github documented-operation parity

Phase `github-parity-sweep-r1`. GSD lifecycle and fallback recorded in
`.planning/traces/gsd-top50-sweep-continue-r1.md`.

## Derivation

- [x] Artifact re-fetched from GitHub's own `rest-api-description` and checked by **byte count**:
      **12,920,264 bytes**, identical to the sweep's recorded derivation. Identical bytes are what
      make it the same artifact rather than a lookalike, and let the derivation be reproduced instead
      of trusted.
- [x] `openapi: 3.0.3`, `info.version: 1.1.4` confirmed from the artifact, not from the ledger.
- [x] 808 paths → **1220** method entries, all unique.
      `GET 636 · POST 193 · PUT 134 · DELETE 187 · PATCH 70`; 37 `deprecated: true`.
- [x] Extraction compared **set-to-set** against the committed `DERIVED-OPERATIONS.json`: 0 operations
      only in the spec, 0 only in the derivation.
- [x] Webhook events counted where they actually live (`x-webhooks`, **270**) and excluded from the
      operation surface by policy; the 28 webhook **management** operations are counted.
- [x] No count adopted from the provider ledger.

## Red before green

- [x] `cmd/connectorgen/github_documented_surface_test.go` was written and **run** against the
      unmodified bundle in slice 1, and the verbatim failure is recorded in `RUN-STATE.json` and
      `TDD-LEDGER.md`. Slice 1 was committed at that red state.
- [x] Finding F5 check: github has **two** surface tests, and the **whole** `cmd/connectorgen` package
      was run before every push — never a targeted `-run`.
- [x] `covered_by.writes` got its own red first: both new tests failed to **compile**
      (`unknown field Writes in struct literal of type engine.SurfaceCoverage`) before the field existed.
- [x] No test weakened, skipped, narrowed or deleted.
- [x] One assertion was **replaced**: the slice-1 spot-pin
      `GET /enterprises/{enterprise}/copilot/billing/seats` names an endpoint the artifact does not
      document (Copilot billing is org-scoped). It was substituted with a real enterprise-scope GET,
      with the reason written into the test. Pin count unchanged.
- [x] `TestGitHubAPISurfaceOperationLedgerMetrics` had its **counts** updated as the truth changed;
      every structural assertion is byte-for-byte unchanged.

## Gates

- [x] `go run ./cmd/connectorgen validate` — 551 connectors, **0 findings**.
- [x] `go run ./cmd/connectorgen surface-sync --check` — clean after the ledger sync.
- [x] `go test ./cmd/connectorgen/` — **ok**, whole package.
- [x] `go test ./internal/connectors/commandrunner/` — **ok**, includes
      `TestEveryImplementedCommandPassesRuntimePreflight`.
- [x] `go test ./internal/connectors/engine/ ./internal/connectors/conformance/` — **ok**, run because
      `covered_by.writes` touched shared code.
- [x] `gofmt` clean; `go build ./...` clean.
- [x] `GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow` — exit 0.

## Reachability — every command run, not assumed

- [x] All **1079** implemented+partial commands invoked through the built binary.
- [x] Routing asserted on the rendered `NAME` line, **not** on the exit code. `pm github <nonsense>
      --help` exits **0** because a namespace miss renders group help by design; the first probe was
      wrong for exactly this reason and was rewritten and re-run.
- [x] `pm github issue close` and `pm github pr reopen` still route after the synthetic rows were
      folded away.

## Surface shape

- [x] **1224** rows = **1220 REST + 4 GraphQL**, reconciling exactly with the artifact.
- [x] **1126** covered · **98** blocked · **0** excluded · **0** blank · **0** duplicates.
- [x] **0** rows containing a space, `?` or `*` — no query-string variants, no wildcards, no
      behaviour-in-the-path.
- [x] **0 of 98** blocked rows lack a `Named dependency:` marker outside the duplicate/deprecated
      models.
- [x] `operation_ledger_version: 1` retained.

## Blast radius

- [x] Shared `operation_endpoint_ledger.json` diffed **object-by-object** against `HEAD`: exactly one
      connector key changed (`github`, 162 → 164). No connector added, none removed.
- [x] `connectorgen validate` re-run across all 551 connectors after every change.
- [x] `covered_by.writes` verified not to change behaviour for any existing bundle: engine,
      conformance and commandrunner suites all green, and `surface-sync --check` reports no drift
      anywhere else.

## Repo safety overlay

- [x] No secret requested, printed, stored or summarized. Four endpoints are **blocked precisely
      because** their documented request body carries an OAuth `access_token`.
- [x] No dependency added.
- [x] No credentialed connector check run; every binary invocation was `--help`, which resolves no
      credentials and calls no provider.
- [x] Reverse ETL remains plan → preview → approval → execute; every new write command declares risk
      and approval text.
- [x] No generic shell, generic HTTP write or generic SQL write tool exposed.

## CLI help / docs / website parity — **open, by design**

- [x] Runtime help is done: every command renders and was invoked to prove it.
- [ ] `docs/cli/**`, `website/**` and generated help/manual artifacts — **regenerated once at the end
      of the sweep**, not per connector (finding F6: a per-connector run rewrites ~1,034 files of
      pre-existing `main` drift).
- [ ] **`TestGoldenTranscripts/root_bare_manual` fails on this branch.** Verified pre-existing before
      this phase — chatwoot and gmail already added command surfaces the committed transcript
      predates. Discharged by the end-of-sweep regeneration, which must happen before the PR merges.

## Not done here, recorded rather than discovered late

- [ ] 12 `oneOf`/`anyOf`-rooted mutations blocked rather than split into per-arm actions.
      `TDD-LEDGER.md` §3b.3 names which are trivially splittable.
- [ ] GitHub's GraphQL schema beyond 4 fixed operations — named scope gap.
- [ ] notion ships the same synthetic-path defect and can now adopt `covered_by.writes`; it is merged
      (#3894) and is not converted here.
- [ ] `code-sanning`, a typo in a pre-existing shipped top-level command name.
