# Verification Checklist — Ashby parity wave05-r1

## Results

- [x] Official Ashby inventory generated from the public ReadMe/OpenAPI page without credentials.
  - Source counts: REST operations=185, OpenAPI webhook events=27, total=212.
  - Implemented counts: streams/changefeed=71, bounded direct reads/search/file metadata=9, reverse-ETL write actions=98.
  - Blocked ledger rows=34 (27 webhooks, 4 inbound assessment-partner endpoints, 1 presigned external upload-handle workflow, 1 conditional side-effect read, 1 typed-multipart submission).
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  - `connectorgen validate: 550 connector(s) checked, 0 findings`.
  - Note: `connectorgen validate internal/connectors/defs/ashby` treats `ashby/` as a defs root and tries to validate `schemas/` and `fixtures/` as connector directories; the usable command for this repo-local tool is the full defs root.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/ashby' -count=1`
  - `ok   polymetrics.ai/internal/connectors/conformance`.
- [x] Ashby native/hook tests
  - `go test ./internal/connectors/native/ashby ./internal/connectors/hooks/ashby -count=1` passed.
- [x] CLI help/docs/website parity
  - Read `pm help docs` before docs generation.
  - `go run ./cmd/pm docs generate --dir docs/cli` passed; Ashby connector docs restored as the only connector manual changes.
  - `npm run gen:website-data` passed under `website/`.
  - `./pm docs validate --connectors-dir docs/connectors` passed.
  - `./pm help connectors`, `./pm ashby`, `./pm ashby candidate --help`, `./pm ashby candidate search --help`, `./pm connectors inspect ashby --json`, and docs/website grep smoke passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout 10m`
  - `ok   polymetrics.ai/internal/cli 271.948s`.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` passed.
- [x] `go build ./cmd/pm` passed.
- [x] `make connector-boundary` passed with clean outcome and existing documented exceptions only.
- [x] `git diff --check` passed.
- [x] `make verify` passed in this worktree.
  - Verified cwd: `/Users/karthiksivadas/.treehouse/cli-83d592/50/cli`.
  - Final line: `MAKE_VERIFY_EXIT:0`.
  - After final docs-count and `--message-body` flag-safety regenerations, `go run ./cmd/connectorgen validate internal/connectors/defs`, `go test ./internal/cli -run 'TestGolden' -count=1`, `go build ./cmd/pm`, `git diff --check`, `./pm docs validate --connectors-dir docs/connectors`, and Ashby command-help smoke checks still passed.
- [ ] One gh-axi final update across parent/subissues with counts and evidence.
- [ ] Clean commit only; no push, no PR, no no-mistakes final pipeline.

## Review fix round 2

- [x] `referralForm.info` is blocked on `ashby-referral-form-info-side-effect-foundation` and has no executable stream or CLI command.
- [x] `applicationForm.submit` is blocked on `ashby-application-form-typed-multipart-foundation` and has no JSON write or CLI command.
- [x] `user.interviewerSettings` and `report.generate` return bounded `json_redacted` direct-read results and are absent from reverse-ETL actions.
- [x] Sync-token-capable stream commands use connector-owned full-refresh-only help and name `ashby-sync-token-checkpoint-foundation`.
- [x] `go test ./internal/connectors/native/ashby -count=1` passed.

## Review fix round 3

- [x] No branch delta remains under `internal/coordination/issueguard/**` or `cmd/prissueguard/main_test.go` relative to the approved base.
- [x] Complete branch path audit contains only Ashby-owned and required Ashby-generated/planning, CLI-golden, documentation-index, website-catalog, and native-registration paths.
- [x] PR 3542 issue reference was read-only audited: its body contains only a canonical issue URL, so the outer PR phase must add `Refs #3207`; this review phase did not mutate the PR.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` passed.

## CI repair handoff

- [x] Preserved the approved issueguard contract; no workflow or guard source was weakened.
- [x] Added `PR-BODY.md` with the required incremental relationship `Refs #3207` and current Ashby scope, verification, and safety evidence.
- [x] `go run ./cmd/prissueguard --title 'chore(ashby): stage unvalidated parity checkpoint' --body-file .planning/phases/issue-3207-ashby-parity-wave05-r1/PR-BODY.md` returned `issueguard: ok (1 linked issue)`.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` and `git diff --check` passed.
- [ ] Outer PR phase applies `PR-BODY.md` to PR #3542 and reruns `require-linked-issue`.

## CI repair handoff round 2

- [x] Read-only inspection confirmed PR #3542 still uses the old checkpoint body with a bare `https://github.com/polymetrics-ai/cli/issues/3207` URL and no accepted issue keyword.
- [x] Made `PR-BODY.md` an explicit verbatim replacement input for the outer PR phase, retaining `Refs #3207` and aligning the evidence headings with the issue-first delivery contract.
- [ ] Outer PR phase replaces PR #3542's live body with `PR-BODY.md`; source-only pushes cannot repair pull-request metadata.
