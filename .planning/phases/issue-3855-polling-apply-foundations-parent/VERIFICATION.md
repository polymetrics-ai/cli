# Verification — #3855 parent topology scaffold

## Required evidence

| Check | Evidence / command | Status |
| --- | --- | --- |
| Supplied isolated worktree | `pwd -P` equals `git rev-parse --show-toplevel`; worktree began detached and clean | passed |
| Initial source ref | local `origin/feat/3862-any-to-any-transport` and `git ls-remote` both resolved to `30b2fb4aeb121641b6158903fe1d3b54668599a6` | passed — historical seed point |
| Combined branch ref | local `origin/docs/4015-connector-release-certification` and `git ls-remote` both resolved to `5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12` | passed |
| Requested branch | `git switch --create feat/3855-polling-apply-foundations refs/remotes/origin/feat/3862-any-to-any-transport` | passed locally |
| GSD adapter | `scripts/gsd doctor`; required `sources` resolutions; generated lifecycle prompts | passed |
| Canonical delivery projection | `go run ./cmd/agentcontractgen check` | passed |
| Parent/child GitHub hierarchy | `gh-axi issue view` / `gh-axi issue subissue list` | live check found five open children: #3856–#3860; #3860 follows #3856–#3859 |
| #3880, #4016, #4019 live inspection | `gh-axi pr view` | passed during parent topology audit; #3880 remains partial reuse only |
| Scope recovery | guarded `no-mistakes axi sync --recover`, whole revert, `git diff --exit-code b2d0f630... HEAD`, and ancestry proof | passed — terminal `81f37c2b...` retained in ancestry; `409933563...` restores the accepted tree |
| Protected inherited blob | `git diff --exit-code` and `git rev-parse <base|HEAD>:docs/architecture/github-postgres-warehouse-certification.md` | passed — both blobs are `889e4eddffabb76aa8be46a934c3e9abe0610f4c` (owner #4015) |
| Markdown and diff hygiene | `git diff --check` and temporary-base changed-path assertion | passed — exactly the nine phase files are present |
| Repository checks | `go run ./cmd/agentcontractgen check`; `make tidy-check`; `make docs-check`; `make lint` | passed after whole revert |
| Prior no-mistakes | `01KZPY9EBYX84WZM11EN1F6C83` with `--skip=push,pr,ci`, no `--yes` | passed; terminal scope expansion was whole-reverted |
| Draft PR shape | `gh-axi pr create` and REST inspection | passed — #4041 is one open draft with exact temporary base/head and nine files |
| Fresh no-mistakes | guarded argv with `--skip=push,pr,ci`; document gate not skipped | pending for this evidence commit |
| Accepted transport head | targeted `git fetch` and `git ls-remote` both resolved `origin/feat/3862-any-to-any-transport` to `c67f40a5ff67a131950f3123e70527027dca8493` (#4059 merged into #4019) | passed |
| Rebase necessity | `git merge-base --is-ancestor c67f40a5... b61d0fa7...` exited 1; current transport head was absent from #3855 ancestry | passed — refresh required |
| Preserved parent range | guarded `no-mistakes axi sync --recover`; `git rebase --onto origin/feat/3862-any-to-any-transport 30b2fb4a...`; `git range-diff 30b2fb4a...b61d0fa7... c67f40a5...HEAD` | passed — all eight commits patch-equivalent, no conflicts |
| Refreshed ancestry and scope | `git merge-base --is-ancestor c67f40a5... HEAD`; `git diff --check`; changed-path assertion | passed — accepted transport head is an ancestor and exactly the nine phase files remain |
| Causal pre-bridge start | fresh `no-mistakes axi sync --check` returned `custody_returned` / `run_pipeline`; one `axi run --skip=push,pr,ci` | RED observed — internal validation-gate ref `b61d0fa7...` rejected clean `e541170e...` as non-fast-forward before new-run creation or external effect |
| Audited preserved-history bridge | exact `git merge --no-ff -s ours` command after fixed-SHA rechecks | passed — `72a4fc32...` tree equals `3b6e3e2e...`; parents are checkpoint then `b61d0fa7...`; `7ea7350b...`, `b61d0fa7...`, `c67f40a5...` are ancestors; remote comparison `0 13`; protected blob unchanged |
| Post-bridge local no-mistakes | exact Stage-5 argv with `--skip=push,pr,ci`, no `--yes` | pending — every synchronous gate must be owned; no external stage may run |
| Current planning-only local gates | `scripts/gsd` execute/verify/review prompts resolved inline; `go run ./cmd/agentcontractgen check`; `make tidy-check`; `make docs-check`; `make lint` | passed at bridge head — no Go/runtime/product gate is applicable |

## GSD verification result

The adapter and command provenance are valid. The inline fallback is required because the canonical
single-worker delivery contract forbids spawning a GSD role. This phase has no executable product
acceptance test: verification is the exact branch, artifact, PR-shape, and scope-fence evidence
listed above. A passing documentation check never certifies product behavior.

## Explicit non-applicable checks

- Go formatting, package tests, build, runtime services, credentials, provider calls, and database
  containers: no Go or runtime path changes.
- CLI help/manual/website parity: no command, flag, output, connector surface, or end-user
  documentation changes.
- Automated review: this PR must remain draft; review coverage is recorded as pending rather than
  requested while draft.

## Pending audited recovery checks

The bounded gap plan and audited tree-preserving non-force `-s ours` bridge now passed every fixed
SHA, parent, tree, ancestry, protected-blob, scope, and local planning-only gate assertion. Run the
contract-owned no-mistakes local pipeline without `--yes` and with
`--skip=push,pr,ci`. No external push, PR update, or CI action is part of this stage. Any later
existing-branch update must first prove `7ea7350b...` is an ancestor and use a normal non-force
push, followed by exact-head CI and live draft/base/head inspection.
