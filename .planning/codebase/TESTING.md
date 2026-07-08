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

## Connector Verification Boundaries

- Planning-only work must not run credentialed live connector checks.
- Fixture/replay gates may run without secrets when relevant.
- Missing live credentials should be recorded as `uncertified`, not failure.
- Reverse ETL execution is out of scope unless plan, preview, approval, execute are explicitly completed.

---
*Testing notes refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
