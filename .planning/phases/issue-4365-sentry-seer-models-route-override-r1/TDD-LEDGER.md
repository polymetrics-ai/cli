# Issue #4365 TDD ledger

## Red — planned before production edits

| Slice | Test | Expected initial failure | Status |
| --- | --- | --- | --- |
| Happy | `TestSentrySeerModelsSourceBoundRoute` | `go test -timeout 20m ./internal/connectors/engine -run '^TestSentrySeerModels' -count=1` failed: Sentry has no command surface. | red captured |
| Bad | `TestSentrySeerModelsRouteRejectsIdentityDriftBeforeProviderIO` | The same focused run failed in every named mutation because Sentry has no command surface to bind to an operation. | red captured |
| Edge | `TestSentrySeerModelsRoutePreservesPathAcrossBaseSlashForms` | The same focused run failed with `operation "sentry.seer_models_list" not found in bundle "sentry"`. | red captured |
| CLI boundary | `TestSentrySeerModelsCommandStopsBeforeProviderIOWithoutCredential` | Pending the generated Sentry command; the standalone built-binary proof will be captured after the green declaration and build. | red planned |

## Green/refactor record

### Green

- Minimal declaration: added only Sentry `operations.json` and `cli_surface.json`, one
  `sentry_api_v0` route in `streams.json`, the exact `api_surface.json` direct-read
  coverage row, and the necessary literal default for the already-declared `base_url`.
  `surface-sync` generated the command-owned endpoint metadata and the one embedded
  Sentry endpoint-ledger row. No shared route/HTTP/parser/hook code changed.
- `go test -timeout 20m ./cmd/connectorgen -run
  '^TestSentrySeerModelsSourceProjectionKeepsExactRouteBinding$' -count=1` — green.
  It reads the preserved Sentry source lock and proves its exact source ID,
  `GET /api/0/seer/models/`, provider operation ID, citation URL, Sentry route,
  operation, CLI path, and `api_surface` row agree.
- `go test -timeout 20m ./internal/connectors/engine -run '^TestSentrySeerModels'
  -count=1` — green. Happy proof finds exactly one stable command and endpoint-ledger
  row, one declared `sentry_api_v0` base/version identity, and no authored
  provider route/base/method/path flags; bad source ID/route/base/method/path
  mutations fail before provider I/O; both trailing and non-trailing declared
  bases send exactly `/api/0/seer/models/`.
- `go test -timeout 20m ./internal/cli -run
  '^TestSentrySeerModels(CommandStopsBeforeProviderIOWithoutCredential|HelpAndBareNamespaces)$'
  -count=1` — green. The empty-credential command returns exactly
  `error: missing --credential` and its process-local HTTP transport spy records zero
  requests; help, bare connector/group, and invalid-action behavior are also covered.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1` regenerated only the nine root-help
  transcript variants that project the connector namespace. The matching focused
  `TestGoldenTranscripts` rerun is green; the tracked help transcript now names
  Sentry exactly once.

### Executable proof

After `go build -o pm ./cmd/pm`, a newly initialized project with no credential ran:

```text
./pm sentry seer list-models --root <fresh-project>
error: missing --credential
exit=1
```

The same path's test spy saw zero provider I/O. The command has no request inputs, so
the bare invocation is its valid typed input. `pm help sentry`, `pm sentry`,
`pm sentry seer`, and `pm sentry seer list-models --help` all render the one declared
command without a caller-controlled path, method, URL, or raw HTTP option.

### Refactor/generated surfaces

- `go run ./cmd/connectorgen validate internal/connectors/defs/sentry` and full-defs
  validation are green.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs` followed by
  `--check` is green; it generated the Sentry endpoint ledger.
- declaration admission and operation evidence are green and unchanged for their
  existing source-lock cohort; Sentry's preserved lock is intentionally test-fixture
  evidence rather than a new broad source-import cohort.
- `pm docs generate` / `pm docs validate` and the website connector-data generator
  regenerated the Sentry manual, connector skill, agent skill, catalog, and website
  CLI surface. The website typecheck, lint (warnings only), and production build are
  green.
