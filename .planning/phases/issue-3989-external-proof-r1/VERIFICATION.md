# #3989 verification checklist

- [x] Every acceptance row has an observable state-change assertion: evidence writer, observer, ephemeral-session, relay, and fresh-child tests each assert a positive write/request/fingerprint or an explicit zero-write refusal.
- [x] External binary is freshly built; evidence records its SHA, exact safe argv, and successful `flow_plan`/`flow_preview`/`flow_run`/`flow_status` references (`TestExternalProofFreshChildCapturesCompleteHTTPSProviderTranscript`).
- [x] No raw credential appears in captured parent streams, project tree, vault/key, or artifact in the fresh TLS child test; parent relay refuses both streams before writing on a canary match. `--value-stdin` is rewritten to a child-only environment reference with no value in argv.
- [x] Error/refusal paths write zero accepted-evidence artifacts (`TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites` and `TestExternalProofFreshChildRefusesNoHTTPSWithoutArtifact`).
- [x] HTTPS transcript covers exact request/response observation and explicit byte bounds; transport tests cover bounded error bodies, complete redirect source/final exchanges, and a zero-write refusal beyond the redirect/retry cap while preserving the child-visible body.
- [x] `go test -timeout 20m ./internal/connectors/certify/...` passes after the final TLS/relay changes.
- [x] Required scoped local gates and inline standard code review complete: `go test -timeout 20m ./internal/{app,cli}`, `go test -timeout 20m ./cmd/connectorgen`, scoped `go vet`, `go build ./cmd/pm`, and the individual `make verify` component gates pass. No #3989 review finding remains.
- [x] CLI help/manual/website parity: `pm connectors`, `pm help connectors`, `pm connectors certify --help`, golden transcript, and `make docs-check` pass after final documentation changes.
- [x] Opt-in PostgreSQL container lane was run with the supplied Docker endpoint. `TestPostgresManagedTargetDriverLiveControlAssertions` still fails with the pre-existing #4158 route-mismatch refusal where a durable acknowledgement is expected; it is out of #3989 scope and no PostgreSQL source was changed.
