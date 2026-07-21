# Issue #474 Verification

Overall: **passed**, including the independent exact-head correction loop. The parent-declared
phase-equivalent child gate excludes Go and root `make verify` for this TypeScript-only correction.

Ready stacked PR: https://github.com/polymetrics-ai/cli/pull/483

| Gate | Result | Evidence |
|---|---|---|
| Correction RED | pass | reviewed head: 36 tests, 21 pass, 15 expected fail; 64-item child killed at one-second deadline |
| Audit gap RED | pass | full case-fold and failed-status coherence: 2 tests, 0 pass, 2 expected fail |
| Focused policy tests | pass | 36 tests, 36 pass, 0 fail; hostile component typed-rejected in bounded time |
| Full Shepherd tests | pass | 173 tests, 173 pass, 0 fail after refactor |
| Strict TypeScript / Pi 0.80.6 | pass | `tsc` 5.9.3 `--noEmit --strict` over all 12 production Shepherd modules, resolving installed Pi 0.80.6 declarations |
| Pi extension discovery | pass with tooling deviation | `pi --list-extensions` is unsupported (exit 1); supported offline RPC `get_commands` passed and returned `pm-shepherd` from the explicit project extension |
| Diff/ownership | pass | `git diff --check`; changed paths restricted to the three owned modules, matching tests, and issue #474 phase directory |
| Historical pre-correction Go gates | pass | initial implementation recorded supplemental vet/test/build; the correction loop did not rerun Go |
| Historical pre-correction `make verify` | `cancelled_by_parent_policy` | parent intentionally terminated it; the correction loop did not invoke it |

## Exact phase-equivalent commands

```bash
node --test .pi/extensions/shepherd/autonomy-policy.test.ts \
  .pi/extensions/shepherd/dependency-graph.test.ts \
  .pi/extensions/shepherd/reconciler.test.ts

node --test --test-reporter=tap .pi/extensions/shepherd/*.test.ts

tmpdir=$(mktemp -d /tmp/shepherd-474-final-ts.XXXXXX)
mkdir -p "$tmpdir/shepherd" "$tmpdir/node_modules/@earendil-works"
find .pi/extensions/shepherd -maxdepth 1 -name '*.ts' ! -name '*.test.ts' \
  -exec cp {} "$tmpdir/shepherd/" \;
ln -s /Users/karthiksivadas/.nvm/versions/node/v24.13.1/lib/node_modules/@earendil-works/pi-coding-agent \
  "$tmpdir/node_modules/@earendil-works/pi-coding-agent"
node /usr/local/lib/node_modules/@opengsd/gsd-pi/node_modules/typescript/lib/tsc.js \
  --noEmit --strict --target ES2022 --module NodeNext --moduleResolution NodeNext \
  --allowImportingTsExtensions --skipLibCheck --types node \
  --typeRoots /Users/karthiksivadas/.nvm/versions/node/v24.13.1/lib/node_modules/@earendil-works/pi-coding-agent/node_modules/@types \
  "$tmpdir"/shepherd/*.ts

printf '%s\n' '{"id":"issue-474-extension-smoke","type":"get_commands"}' | \
  PI_OFFLINE=1 pi --mode rpc --no-session --no-extensions \
  --extension .pi/extensions/shepherd/index.ts --no-skills --no-prompt-templates --approve

git diff --check
```

Runtime-backed services are not applicable: this is a pure TypeScript policy slice with no network,
credential, database, Temporal, Redis-compatible, or Podman behavior. CLI help/docs/website parity
is not applicable because no CLI surface is changed.

Automated review route, if exposed by this policy boundary: `codex_independent` using
`openai-codex/gpt-5.6-sol` with `xhigh` reasoning on an exact head/range. This issue does not wire or
invoke any review adapter. Claude and Copilot must not be requested for this sub-PR.

No automated review route was exposed or wired in this pure slice. No Claude, Copilot, human, or
mislabelled review request was made. No merge action was taken. Verified implementation head before
the final evidence-only commit: `ef2fd1e280128ccb2a0e46b749f9638472fad865`.
