# Phase 427 Verification

Session `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `ab847da28ebf78e5732ac1edcde8e39f92dc2656`.

## TDD and behavior

- [x] Six phase artifacts created before production edits.
- [x] Exact focused RED captured before production edits (`11.332s`; native count/ownership failures only).
- [ ] Native `docs` namespace plus `generate` and `validate`; legacy wrapper removed.
- [ ] Native `--dir`/`--connectors-dir` preserve applicable spaced, assigned, repeated-last-wins, bare `true`, comma/path forms.
- [ ] Unknown flags and extra action arguments retain legacy compatibility.
- [ ] Bare/text/JSON/flag/short/positional help parity.
- [ ] Missing/empty output path and invalid action/error categories preserved.
- [ ] Global/config root and assigned boolean forms preserved.
- [ ] Generated CLI bytes, connector docs/catalog/icons, validation behavior, filesystem containment, and output text preserved.
- [ ] Docs-only `parseFlags` call removed; shared parser remains for unrelated legacy/dynamic namespaces.

## Gates

- [ ] Focused docs/router tests.
- [ ] Focused docs/router/golden tests.
- [ ] Golden transcript test.
- [ ] Full `internal/cli`.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check`.

## CLI help/manual/website parity

- [ ] `pm help docs`, bare `pm docs`, `pm docs --help`, `pm docs -h`, and `pm docs help` byte-identical, exit 0.
- [ ] Bare, flag, assigned-boolean, and positional JSON help return canonical `CommandManual/docs`.
- [ ] Built operation checks cover generate/validate, both output-directory flags and applicable forms, unknown flags, configured JSON, assigned true/false JSON, and spaced/assigned/late root/global forms.
- [ ] Invalid actions remain usage exit 2 and real failures are not masked by help.
- [ ] `docs/cli/docs.md`: update or record no-update byte-parity evidence.
- [ ] Connector docs generation and validation pass from a safe temp root.
- [ ] `website/**`: update or record no-update website-generator diff evidence.
- [ ] Generated/golden help: update or record no-update golden/diff evidence.
- [ ] Completion/discovery: action names unchanged; no-file completion seam tested; Phase 15 implementation deferred.
- [ ] Phase 14 viewer and Phase 19 focused help/man churn remain untouched.

## Safety/scope/delivery

- [ ] No secrets/credentials requested, read, printed, summarized, or stored.
- [ ] Filesystem checks use temporary roots and validate generated paths stay local.
- [ ] No services, credentialed connector checks, destructive/admin actions, or production deploys.
- [ ] No dependency or `go.mod`/`go.sum` delta.
- [ ] No connector-def, unrelated namespace, Phase 14 viewer, or Phase 19 help/man production delta.
- [ ] Required `make verify` local sample smoke, if run, follows reverse ETL plan → preview → approval → run.
- [ ] Planning, RED, implementation, and verification checkpoints committed and pushed to `origin/refactor/427-docs-native-cobra`.
- [ ] No PR or external review requested.

Result: in progress.
