# Phase 426 Verification

Session `issue-426-pi-openai-codex-gpt-5.6-sol-high-20260718T104457Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `54bfcbab5a997c81676b286fe78de00a109f5fba`.

## TDD and behavior checklist

- [x] Six phase artifacts created before production edits.
- [ ] Exact focused RED captured before production edits.
- [ ] Native `skills` namespace and `generate` action; legacy wrapper removed.
- [ ] Native `--dir` preserves spaced, assigned, repeated, bare, comma/path forms.
- [ ] Unknown flags and extra action args retain legacy compatibility.
- [ ] Bare/text/JSON/flag/short/positional help parity.
- [ ] Missing dir validation and invalid action usage categories preserved.
- [ ] Global/config `--root`/`--json` forms, including assigned booleans, preserved.
- [ ] Existing skill generation files, metadata-only security, and output envelopes preserved.

## Gates

- [ ] Focused skills/router tests.
- [ ] Focused skills/router/golden tests.
- [ ] Full `internal/cli` tests.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check`.

## CLI help/manual/website parity

- [ ] `pm help skills`.
- [ ] bare `pm skills`.
- [ ] `pm skills --help` and `-h`.
- [ ] positional `pm skills help`.
- [ ] JSON manuals for flag and positional routes.
- [ ] invalid action remains usage and does not mask error with help.
- [ ] `docs/cli/skills.md` temp-generation diff reviewed (expected unchanged).
- [ ] website generator/diff reviewed (expected unchanged).
- [ ] generated/golden help diff reviewed (expected unchanged).
- [ ] completion/discovery unchanged except tested native no-file seam; Phase 15 deferred.

## Safety/scope/delivery

- [ ] No secrets/credentials read or stored; generated content remains metadata-only.
- [ ] Filesystem tests use `t.TempDir`; no path-policy weakening.
- [ ] No services, credentialed checks, destructive/admin actions, or deploys.
- [ ] No dependency or `go.mod`/`go.sum` delta.
- [ ] No connector-def or unrelated namespace production delta.
- [ ] Coherent checkpoints committed and pushed to `origin/refactor/426-skills-native-cobra`.
- [ ] No PR or external review requested.

Result: pending implementation and verification.
