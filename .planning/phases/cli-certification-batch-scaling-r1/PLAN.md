# Plan — certification batch scaling

## Scope

Measure live execution throughput only. Do not change connector runtime behavior, candidate generation, importer behavior, credential-scope validation, fixture data, or generated connector surfaces.

## TDD / measurement slices

1. **Red — reject incomplete measurement configuration.** Start the runner without the non-secret `rate_limit_account` required by the connector's declared authenticated-user policy. It must refuse before a provider request instead of silently running without accountable rate-limit coordination.
2. **Green — candidate manifest integrity.** Add the declared non-secret rate-limit account, then generate the 10 and 100 temporary manifests from PR #4214's committed `certification.json`; verify respectively 10 and 100 direct-read candidates with exactly one `/response` object-or-array assertion and no mutation intent.
3. **Live 10.** Time fresh setup → PM init → credential validation/direct reads/checkpoint persistence → report persistence → teardown. Capture safe stage and rate metadata.
4. **Live 100.** Repeat the complete lifecycle for the deterministic hundred-operation manifest. Every terminal result is classified without collapsing non-passes.
5. **Live repeated 100s.** Run at least three further fresh hundred-operation batches serially. If a rate-limit event occurs, stop and resume from that run's checkpoint after the authoritative reset; record both wait and resumed count.
6. **Projection / staging / verification.** Derive the remaining-read estimate from the measured repeated-batch curve, stage only sanitized non-accepted report input, run repository gates, review, commit, push, and open the direct PR.

## Expected red/green evidence

- **Red:** The runner without `rate_limit_account` is refused by the connector's declared authenticated-user rate-limit policy before it can send a provider request.
- **Green:** With the required non-secret config supplied, all timed manifests are direct-read-only, every candidate has the generated `/response: object_or_array` assertion, and only an assertion-bearing successful stage is counted as a produced-value pass.
- **Green:** Each timed manifest has the declared count, all commands are direct reads, and every pass is assertion-bearing.

## Verification plan

- `go run ./cmd/pm connectors certify --help` and `go run ./cmd/pm help connectors` before live use.
- Candidate source validation: `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m ./internal/connectors/certify`, and `go run ./cmd/connectorgen certification-candidates --connector github --check` in the pinned disposable source.
- Live operational check: terminal rows, bucket sums, assertion types, reusable checkpoint fingerprints, safe rate events, and teardown absence.
- Repository gates: `gofmt -l cmd internal`; `go vet ./...`; `go test -timeout 20m ./cmd/connectorgen`; `go test -timeout 20m ./internal/connectors/certify`; `go test -timeout 20m ./internal/cli`; `go build ./cmd/pm`; `make` verification sub-gates; `scripts/verify-gsd-workflow`; and generator byte-stability where applicable.
