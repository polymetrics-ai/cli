# Issue 400 Verification Checklist — Cobra Router Shell

**Issue:** #400
**Branch:** `refactor/400-cobra-router-shell`
**Base:** `feat/cli-architecture-v2`

## Required gates

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/ -run Golden -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go test ./internal/cli/ -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`

## Focused TDD gates

- [ ] Red validation before implementation: `go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'` fails.
- [ ] Red focused test before implementation: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` fails.
- [ ] Green focused test after implementation: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` passes.
- [ ] Golden suite byte-identical: no golden fixture update unless explicitly recorded.
- [ ] Certify in-process re-entrancy gate stays green.

## Dependency gate

- [ ] Direct dependency added exactly: `github.com/spf13/cobra v1.10.2`.
- [ ] Expected transitives only: `github.com/spf13/pflag`, `github.com/inconshreveable/mousetrap`.
- [ ] No additional direct dependency introduced.
- [ ] `make tidy-check` remains green through `make verify`.

## CLI help / docs / website parity

Applies: yes, because routing/help behavior is CLI-visible even though target is byte-identical.

- [ ] Runtime help: build local binary and run `/tmp/pm-400 help connectors`.
- [ ] Bare namespace: `/tmp/pm-400 connectors` exits 0 and prints existing contextual manual.
- [ ] Command help: `/tmp/pm-400 docs --help` preserves current golden behavior.
- [ ] Hidden command help: `/tmp/pm-400 worker --help --json` preserves current golden behavior.
- [ ] `docs/cli/**`: not updated if `go test ./internal/cli/ -run Golden` and docs-generate-diff remain green; exact reason to record.
- [ ] `website/**`: not updated if no command/flag/help text changes; exact reason to record.
- [ ] Generated help/manual artifacts: golden suite remains byte-identical; docs-generate-diff remains green.
- [ ] Completion/discovery metadata: not applicable in Phase 2; shell completion is Phase 15.

## Review route

- [ ] Open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #400` and `Refs #397`.
- [ ] Do not post `@claude review` because repository Claude workflow is already `disabled_manually`.
- [ ] Do not request Copilot because quota is exhausted for this blocker window.
- [ ] Record review status as human/parent-PR fallback pending; no approval claims.

## Results

Pending.
