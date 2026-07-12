# Verification: Issue #132 HubSpot CLI Feature Parity Parent

Date: 2026-07-10

## Adapter and setup checks

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
```

Result:

- `scripts/gsd doctor`: passed; repo root and Pi adapter resources found; 69 commands registered.
- `scripts/gsd verify-pi`: passed.
- `scripts/gsd list --json`: passed; JSON parsed with 69 commands.

## GSD programming-loop availability

```bash
scripts/gsd prompt programming-loop init --phase issue-132-hubspot-cli-parity --dry-run
```

Result: blocked with `scripts/gsd: unknown GSD command: programming-loop`.

Fallback prompts captured:

```bash
scripts/gsd prompt plan-phase issue-132-hubspot-cli-parity --skip-research
scripts/gsd prompt execute-phase issue-132-hubspot-cli-parity --dry-run
```

Result: both generated official GSD adapter prompts successfully.

## Parent seed verification

Pending commands after parent planning files are written:

```bash
python3 -m json.tool .planning/phases/issue-132-hubspot-cli-parity/RUN-STATE.json >/dev/null
python3 -m json.tool .planning/phases/issue-132-hubspot-cli-parity/ORCHESTRATION-STATE.json >/dev/null
```

## Targeted implementation verification (#134 first)

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'CLISurface|HubSpot'
go test ./internal/connectors/engine -run 'CLISurface|HubSpot'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Full parent handoff gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity checklist

Applies to #135 and any user-visible command/help changes. For #134 metadata-only changes, record whether runtime dispatch/help is not yet applicable.

- [ ] `pm help <topic>` checked where command help is exposed.
- [ ] Bare namespace help checked where command namespace is exposed.
- [ ] `pm <command> --help` checked where command exists.
- [ ] `docs/cli/**` updated or not-applicable noted.
- [ ] `website/**` updated or not-applicable noted.
- [ ] Generated help/manual artifacts updated or not-applicable noted.
- [ ] Tests cover the applicable help/docs behavior.

## Review route

No review command should be posted until implementation and local verification pass. For a non-draft parent PR targeting `main`, rely on CodeRabbit automatic review. For stacked/non-default sub-PRs, use the parent PR fallback rules and record coverage.
