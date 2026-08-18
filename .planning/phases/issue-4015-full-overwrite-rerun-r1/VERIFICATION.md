# Verification Checklist

**Status:** locally verified; PR base/API and external review gates remain

## Acceptance checks

- [x] Red unit test proves full-refresh modes incorrectly receive prior checkpoints before the fix.
- [x] Red live binary test reproduces run-two `0/0` and an invalid empty target state with exact counts.
- [x] Green unit test proves `full_append` and `full_overwrite` omit prior checkpoints.
- [x] Green unit test proves all four requested incremental modes retain prior checkpoints.
- [x] Green live binary test independently proves target replacement after source deletion/update/insertion.
- [x] Green live incremental test proves unchanged `incremental_upsert` still skips with `0/0`.
- [x] Targeted packages pass with `-timeout 20m`.
- [x] `go vet ./...` and `go build ./cmd/pm` pass.
- [x] Applicable generated, docs, contract, boundary, and release-workflow gates pass.
- [x] Code review has no unresolved critical or warning findings.
- [ ] PR base is API-read back as `integration/4015-mvp-flat-r1`.

Exact commands, results, intentionally excluded monolithic commands, and the known unrelated tagged CDC-restart failure are recorded in `traces/local-verification.md`.

## CLI/docs/website parity

Not applicable: this task changes no command, flag, output schema, connector surface, help topic, manual content, or website behavior.

## Live runtime safety

- Use only the existing opt-in harness and explicit direct local Unix endpoint.
- Do not start, stop, or restart Colima/Docker/Podman.
- Do not print or persist credential values.
- Verify harness cleanup through its existing unconditional cleanup assertions.
