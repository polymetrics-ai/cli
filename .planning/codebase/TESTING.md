# Testing and Verification

**Generated via:** official GSD Core Pi adapter command path.

## Standard Go Gates

Use when Go source changes:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Runtime-backed optional checks:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```

Runtime-backed checks cover optional Podman/Docker Compose services: PostgreSQL on `localhost:15433`, DragonflyDB/Redis-compatible coordination on `localhost:6379`, Temporal on `localhost:7233`, and Temporal UI on `http://localhost:8080`. They are not required for dependency-free unit tests or planning-only work.

RLM/Pi-agent runtime checks, when explicitly in scope, may include:

```bash
pm runtime doctor --json
pm agent image ensure --json
pm worker status --json
pm rlm run --spec spec.json --in customers --out scored_customers --mode agent --request "score leads" --json
```

Do not run RLM agent mode, worker services, or credentialed runtime checks unless the issue explicitly requests runtime-backed verification.

## CLI Help / Docs / Website Parity Gates

Use for CLI command, flag, help, output, or connector-surface feature work:

```bash
pm help <topic>
pm <namespace>
pm <command> --help
rg -n "<command>|<flag>|<topic>" docs/cli website
```

Expected behavior:

- namespace commands such as `pm connectors` render contextual help/subcommand summary and exit 0
  when no action is selected; future human-first bare `pm query`/`pm reverse` workspaces are the
  documented dual-TTY exception and must fall back to deterministic help on every bypass path;
- invalid actions still return usage errors;
- `docs/cli/**`, `website/**`, generated help/manual artifacts, and golden tests are updated or explicitly marked not applicable.

## Planning-Only GSD/Pi Gates

Use for issue #122 planning/agent refreshes:

```bash
scripts/gsd doctor
scripts/gsd version
scripts/gsd list
scripts/gsd verify-pi
node -e "JSON.parse(require('fs').readFileSync('.planning/config.json','utf8')); console.log('config ok')"
node -e "JSON.parse(require('fs').readFileSync('.gsd/commands.json','utf8')); JSON.parse(require('fs').readFileSync('.gsd/upstream.lock.json','utf8')); JSON.parse(require('fs').readFileSync('.pi/settings.json','utf8')); console.log('json ok')"
git diff --check
git diff --name-only -- cmd internal
git diff --name-only -- .planning/phases
```

Expected for this refresh:

- `cmd/internal` diff is empty.
- `.planning/phases` diff is empty.
- GSD adapter reports official source and Pi resources.

## Agent Spec Verification

Use when `.agents/**` changes:

```bash
python3 - <<'PY'
import pathlib, yaml
for path in pathlib.Path('.agents').rglob('*.yaml'):
    yaml.safe_load(path.read_text())
print('yaml ok')
PY
```

If PyYAML is unavailable, use a runtime-neutral parse/check available in the current environment and record fallback.

## Website Verification Boundaries

When website docs or UI change, choose relevant checks from:

```bash
cd website
npm run gen:website-data
npm run typecheck
npm run test:unit
npm run test:e2e
npm run build
```

Do not add frontend dependencies without human approval.

## Connector Verification Boundaries

- Planning-only work must not run credentialed live connector checks.
- Fixture/replay gates may run without secrets when relevant.
- Missing live credentials should be recorded as `uncertified`, not failure.
- Reverse ETL execution is out of scope unless plan, preview, approval, execute are explicitly completed.

---
*Testing notes refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
