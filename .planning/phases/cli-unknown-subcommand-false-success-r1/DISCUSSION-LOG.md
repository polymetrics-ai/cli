# Discussion log — cli-unknown-subcommand-false-success-r1

## GSD discussion execution

`scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed. The resolved
`discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` prompts
were generated with `scripts/gsd prompt` and are being followed inline.

Pi's interactive runtime is unavailable to this worker, while the repository's canonical
single-worker contract and this task prohibit role spawning. This is the documented manual-GSD
fallback from `.agents/agentic-delivery/references/gsd-pi-adapter.md`.

## Fixed decisions

| Area | Decision | Source |
| --- | --- | --- |
| Scope | Only CLI resolution/error handling and its tests change. | Task brief |
| Invalid help path | A registered connector with an unresolved path, even when `--help` is present, returns the existing usage error shape. | Task brief |
| Valid help | Connector root, a declared group, and a declared deep command retain successful help output. | Task brief |
| Error behavior | Connector-level unknown names retain `unknown command \"<name>\"` at exit 2; JSON retains `usage_error`. | Task brief |
| Artifacts | No bundle definitions or generated command surfaces are changed. Golden transcripts are regenerated only if their input matrix changes. | Task brief |

No product ambiguity remains. The work uses local fixtures only and does not resolve credentials or make a network request.
