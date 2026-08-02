# VERIFICATION — issue 3583 PM/no-mistakes connector lane

## Required issue checks

```bash
rg -n "connector implementation|foundation|target connector|ownership guard|no-mistakes" .agents .pi .github docs
scripts/gsd doctor
```

## Scoped docs/template checks

```bash
rg -n "exactly one target connector|ownership guard evidence|foundation issue/PR|foundation PR|no-mistakes.*foundation split|foundation split.*no-mistakes" .agents .pi .github
rg -n "target connector scope|changed-path compliance|foundation PR path" .agents .pi .github
rg -n "shared runtime/tooling|unrelated connector" .agents .pi .github
```

## Hygiene checks

```bash
git diff --check
```

## Go gates applicability

This issue edits documentation/templates/prompts only and does not edit `cmd/`, `internal/`, or generated runtime surfaces. Go gates are not required for this slice unless executable validation is added. If production Go files unexpectedly change, run:

```bash
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
make verify
```

## Results

| Command | Result | Notes |
| --- | --- | --- |
| `scripts/gsd doctor` | pass | Pre-plan |
| `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` | pass | Prompt generated |
| `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` | fallback | `scripts/gsd: unknown GSD command: programming-loop` |
| red docs validation | pass (expected failure captured) | all four required grep checks missing before production edits; command exited 1 |
| issue verification grep | pending | run after edits |
| `git diff --check` | pending | run after edits |
| no-mistakes scoped validation | pending | run after commit if feasible; stop on daemon error or ask-user finding |
