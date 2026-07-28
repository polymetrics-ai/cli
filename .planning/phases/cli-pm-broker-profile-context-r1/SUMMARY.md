# Summary — CLI PM Broker profile/context foundation

Status: implementation complete; local verification passed; no-mistakes/PR verification pending.

This phase targets CLI #566 on the `integration/pm-broker-production-program` parent branch. It adds a metadata-only PM Broker foundation for Organization, Workspace, Environment, BrokerProfile, named context state, context resolution precedence, runtime-mode policy validation, safe broker config keys, CLI help/docs surfaces, and a 426 `incompatible_contract_version` seam.

Safety boundaries held:

- No live broker/provider operations were enabled.
- No credentials, legacy vault entries, raw secret values, production resources, new dependencies, public stable auth registry/plugin SDK claims, or generic authenticated HTTP/SQL/shell/raw JSON escape hatches were added.
- Runtime modes are constrained to `remote`, `local`, and policy-bound `hybrid`; production writes and scheduled production jobs cannot use local fallback.
- PM Broker user state is safe metadata only, rejects unknown JSON fields, rejects non-canonical context/runtime/policy values, and is written owner-readable.

Manual-GSD fallback remains recorded because `scripts/gsd prompt programming-loop ...` is not exposed by the current repo adapter, though `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase ...` worked. The PM Broker context is documented in code, generated CLI docs, website docs, and this phase artifact. `fm-ensure-agents-md.sh .` reported a pre-existing project-memory convention conflict because both `AGENTS.md` and `CLAUDE.md` are tracked real files; per firstmate direction, no agent-memory files were edited in this lane.

Local verification passed with `git diff --check`, targeted PM Broker/config/CLI tests, full `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, connector docs validation, CLI help/docs parity checks, and `make verify`.
