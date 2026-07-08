# Verification — Phase 1 Inventory and Surface Reconciliation

## Required Commands

```bash
node -e "JSON.parse(require('fs').readFileSync('.planning/config.json','utf8')); console.log('config ok')"
test -f .planning/PROJECT.md
test -f .planning/REQUIREMENTS.md
test -f .planning/ROADMAP.md
test -f .planning/STATE.md
test -d .planning/codebase
test -f .planning/codebase/STACK.md
test -f .planning/codebase/INTEGRATIONS.md
test -f .planning/codebase/ARCHITECTURE.md
test -f .planning/codebase/STRUCTURE.md
test -f .planning/codebase/CONVENTIONS.md
test -f .planning/codebase/TESTING.md
test -f .planning/codebase/CONCERNS.md
rg -n "connector parity|reverse ETL|binary|direct-read|GraphQL|XML|SOAP|CDC|queue|human-gated|certification|conformance" .planning

git diff --check
git diff --name-only -- cmd internal
```

## Expected Results

- `config ok` is printed.
- All `test` commands exit 0.
- `rg` finds multi-surface connector parity, safety, certification, and conformance language.
- `git diff --check` exits 0.
- `git diff --name-only -- cmd internal` prints no output.

## Archive Verification

Previous active planning archive:

```text
../planning-archives/polymetrics-cli-issue-122-pre-rebootstrap-20260708173641.tar.gz
sha256 e0959e4c8eba6e8610255a0cd9a98b39267902ba19600515abfdab726bfd57f5
```

## Not Run

- `go test ./...` — not required for planning-only issue and no Go source changed.
- Credentialed connector checks — explicitly out of scope.
- Reverse ETL execution — explicitly out of scope.
- Runtime-backed integration tests — optional and not needed for issue #122.
