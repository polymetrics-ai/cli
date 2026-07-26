# Releasing `pm` and adding a connector

This is the operator/authoring recipe for two recurring tasks in this repo:

1. **Cutting a `pm` binary release** (how versioning and asset publishing work end to end).
2. **Adding a connector** as a set of GitHub issues that mirror the established gong structure.

Everything below is derived from the actual repo config: `release-please-config.json`,
`.release-please-manifest.json`, `.goreleaser.yaml`, `.github/workflows/release.yml`,
`scripts/verify-release-assets.sh`, and `.github/ISSUE_TEMPLATE/agent_task.yml`. If any of those
files change, update this doc in the same PR (CLI/release parity rule in `AGENTS.md`).

---

## 1. How releases work here

Releases are fully automated by two tools wired together in `.github/workflows/release.yml`:

### release-please drives the version from Conventional Commits

- Config: `release-please-config.json` — `release-type: go`, `package-name: pm`,
  `changelog-path: CHANGELOG.md`, `include-component-in-tag: false`, and
  `pull-request-title-pattern: "chore: release ${version}"`.
- Manifest: `.release-please-manifest.json` — the tracked "last released" version, seeded at
  `0.1.0`.
- Bootstrap: `bootstrap-sha: e28260ffb333951c0882f5e6535655c78fcc4610` — release-please only
  considers commits *after* this SHA when computing the next version.

release-please reads the Conventional Commit prefixes since the last release and picks the bump:

| Commit prefix | Default bump in this repo |
| --- | --- |
| `feat:` | **minor** (e.g. `0.3.0` → `0.4.0`) |
| `fix:` | **patch** (e.g. `0.3.1` → `0.3.2`) |
| `feat!:` / `fix!:` / `BREAKING CHANGE:` footer | **major** (e.g. `0.3.0` → `1.0.0`) |
| `docs:`, `test:`, `chore:`, `style:`, `ci:`, `refactor:` | no version bump on their own |

On every push to `main`, the `release-please` job opens or updates a single **`chore: release
<version>`** PR that stages the computed version bump, the generated `CHANGELOG.md`, and the
manifest update. That PR is the release candidate; nothing is tagged until it is merged.

### goreleaser builds and attaches the binaries

- Config: `.goreleaser.yaml` — `version: 2`, `project_name: pm`, `release.mode: keep-existing`
  (it attaches to the release release-please already created rather than recreating it).
- Build: `builds[0]` compiles `./cmd/pm` with `CGO_ENABLED=0` across **6 targets** —
  `{linux, darwin, windows} × {amd64, arm64}`.
- Version injection: ldflags stamp the tag into the binary via
  `-X polymetrics.ai/internal/cli.version={{ .Version }}` (plus `commit` and `buildDate`), so
  `pm version` reports the released tag.
- Archives: `tar.gz` for linux/darwin, `zip` for windows, each bundling `LICENSE`, `NOTICE`,
  `README.md` alongside the `pm` binary. A `checksums.txt` manifest covers all archives.

### The workflow jobs (`.github/workflows/release.yml`)

| Job | Trigger | What it does |
| --- | --- | --- |
| `release-please` | push to `main` / `workflow_dispatch` | Opens/updates the `chore: release` PR; on merge of that PR it creates the tag + GitHub release and exports `release_created` / `tag_name`. |
| `package-check` | every `pull_request` | Runs `goreleaser release --snapshot --clean` then `scripts/verify-release-assets.sh dist` — so asset shape is verified on **every PR**, well before a release. |
| `release-assets` | `release: published` OR `release-please.outputs.release_created == 'true'` | Checks out the tag, runs `goreleaser release --clean --skip=publish`, re-verifies assets, and uploads the 6 verified archives + `checksums.txt` to the published release. |
| `website-release` | same run, after `release-assets` **succeeds** (`needs: [release-please, release-assets]`) | Dispatches `website.yml` on `main` (`gh workflow run`) so the website rebuilds and deploys **with** the code release, carrying the release's docs. The dispatch passes the release tag and skips itself when a website run for that tag already exists. Still gated downstream by `WEBSITE_DEPLOY_ENABLED`. |

`scripts/verify-release-assets.sh` is the guardrail: it asserts exactly one archive per
`{os, arch}` target, verifies each archive's contents (`LICENSE`, `NOTICE`, `README.md`, binary),
and checks that `checksums.txt` covers exactly those assets and that every checksum validates. Keep
the target list in that script aligned with `.goreleaser.yaml`.

---

## 2. How to cut a release

You do **not** create tags or run goreleaser by hand. The steps are:

1. **Merge feature/fix work to `main`** using Conventional Commit titles. Each push updates the
   pending `chore: release <version>` PR.
2. **Review the `chore: release <version>` PR** that release-please maintains (currently
   [PR #16](https://github.com/polymetrics-ai/cli/pull/16); see §4). Confirm the version bump and
   the generated `CHANGELOG.md` look correct.
3. **Merge that PR** (human gate — captain/maintainer only). Merging it:
   - bumps `.release-please-manifest.json` to the new version,
   - commits the `CHANGELOG.md`,
   - creates the git **tag** `v<version>` and the **GitHub release**.
4. **`release-assets` runs automatically** off `release_created == true`: it rebuilds the 6
   platform archives + `checksums.txt`, re-verifies them, and attaches them to the GitHub release.
5. **`website-release` runs automatically** in the same workflow run, but only **after
   `release-assets` succeeds** (see “Coordinated release” below): it dispatches `website.yml` on
   `main`, which rebuilds the site from the current tip of `main` and — when
   `WEBSITE_DEPLOY_ENABLED == 'true'` — deploys it. That tip already contains the release's docs
   (they merged before the release was cut), but it is **not pinned to the exact release commit**:
   `workflow_dispatch --ref main` resolves to whatever `main` points at when the dispatched run
   starts, so it may also carry docs from PRs merged after the release. Pinning the dispatch to the
   release tag is not possible without breaking `website.yml`'s `github.ref_name == 'main'` deploy
   gate. Code release and website publish thus happen together off the one merge, with the site
   serving a superset of the release's docs.
6. **Verify**: the GitHub release should list 6 archives + `checksums.txt`, `pm version` from a
   downloaded binary should report the new tag, and the website should be serving the updated docs.

No manual `git tag`, `gh release create`, or `goreleaser release` is ever required or authorized.

### Coordinated release: code + website together

**Before this coupling** the two publishes were decoupled: `.github/workflows/website.yml` deploys only on
`push` to `main` **path-filtered to `website/**`** (plus `workflow_dispatch`), and its `deploy` job is
gated by the `WEBSITE_DEPLOY_ENABLED` repo variable. A release-please merge commit changes
`CHANGELOG.md` and `.release-please-manifest.json` — **not** `website/**` — so it never triggered a
website deploy, and the website was never tied to the release tag.

**The coupling** is the `website-release` job in `.github/workflows/release.yml`. It runs
in the same workflow run as `release-please`, declares `needs: [release-please, release-assets]`, and
is gated on `needs.release-assets.result == 'success'` **plus** the release trigger
(`release_created == 'true'` or a `release: published` event). The assets gate is what makes the
“together” guarantee real: the site is never published for a release whose binaries failed to build
or verify. On the manual `release: published` path `release-please` is skipped, which is why the
success check reads `release-assets`' job result rather than a `release-please` output. A tag-scoped
`concurrency` group (`website-release-<tag>`, `cancel-in-progress: false`) mirrors `release-assets`
and serializes the idempotence check when both the push run and a `release: published` run fire for
one release — which happens when `RELEASE_PLEASE_TOKEN` (a PAT) is configured, since a PAT-created
release *does* emit `release: published`. The job searches `website.yml` workflow-dispatch runs on
`main` for the release-tagged run name first, dispatches only if none exists, then waits until the
new run is visible before the concurrency slot is released. The website workflow receives the tag as
an optional `release_tag` dispatch input:

```yaml
gh workflow run website.yml --repo "$GITHUB_REPOSITORY" --ref main -f release_tag="$TAG_NAME"
```

Why this shape (not a `release:`-triggered website run): a release created with `GITHUB_TOKEN` does
**not** emit a `release: published` event that starts fresh workflow runs (GitHub's recursion guard).
The repo already works around this for assets by gating `release-assets` on the `release_created`
job output in the *same* run; `website-release` follows the same pattern. `workflow_dispatch` is the
documented exception that `GITHUB_TOKEN` **is** allowed to trigger, and dispatching `--ref main`
keeps `website.yml`'s existing `github.ref_name == 'main'` deploy conditions satisfied while picking
up the docs on `main`'s tip (see the caveat in step 5 above — the ref is `main`, not the tag).

**Safety / reversibility:**

- The job runs **only** on a real release (`release_created`/`release: published`) whose
  `release-assets` job succeeded, never on PRs or ordinary pushes, and dispatches the website at
  most once for a release tag.
- Actual **deployment** is still gated by `WEBSITE_DEPLOY_ENABLED`; if it is not `'true'`, the
  dispatched run still runs checks, builds the site, and **pushes the `:main`-tagged website image to
  GHCR** — `website.yml` gates its GHCR login and `PUSH_IMAGE` only on
  `github.ref_name == 'main' && event_name in (push, workflow_dispatch)`, with no
  `WEBSITE_DEPLOY_ENABLED` check — but the `deploy` job is skipped. So merging this coupling **does
  not** deploy anything by itself, and no website deploy happens until a release is actually cut with
  `WEBSITE_DEPLOY_ENABLED == 'true'`.
- To stop release-triggered website runs entirely, delete the `website-release` job. To keep those
  runs from deploying publicly, set `WEBSITE_DEPLOY_ENABLED` to a value other than `'true'`.
- Manual fallback (if the dispatch is ever removed or fails): after merging the `chore: release` PR,
  a maintainer runs `gh workflow run website.yml --ref main -f release_tag=v<version>` to publish
  the site for that release.

This doc authorizes no website deploy and no release; it only wires and documents the coupling.

---

## 3. Version ↔ connector relationship

Connectors are shipped as data under `internal/connectors/defs/<name>/` (see §5). Each new
connector lands via `feat:` (often `feat(<name>): ...`) commits, so **adding a connector bumps the
MINOR version** in the next release (patch releases are reserved for `fix:` commits). A wave of
several new connectors still bumps one minor version for the batch, not one per connector — the
bump is computed from the highest-priority commit type since the last release.

---

## 4. Investigation: is the first release PR open? (Yes — PR #16)

**Finding: release-please is healthy and the release PR already exists.** The premise that "no
`chore: release` PR has opened" does not hold as of this writing.

Evidence gathered:

- The **Release workflow runs and succeeds on every push to `main`** (latest run at time of
  writing: `30160297968`, conclusion `success`, `release-please` job green).
- There are **41 commits since `bootstrap-sha`** (16 `feat`, 8 `fix`, 1 `feat!`, plus
  docs/chore/etc.) — i.e. plenty of releasable commits.
- The `release-please` job log shows it computed a candidate and reports:
  `✔ Successfully opened pull request: 16.` → **[PR #16 `chore: release 1.0.0`](https://github.com/polymetrics-ai/cli/pull/16)** is open.

So **no release-*enabling* change is required** — the automation works and a candidate PR is open.
The only open item is the **version number**, addressed next.

### Version discrepancy: manifest `0.1.0` vs computed `1.0.0`

- `.release-please-manifest.json` seeds `0.1.0` and is treated here as the **authoritative target
  first version** (per this task's instruction, and because the repo does not clearly mandate
  otherwise).
- PR #16 nevertheless computes **`1.0.0`**. Reason: the history since bootstrap contains a breaking
  change — `feat!: complete connector architecture v2 migration (#29)`. Under release-please's
  **default `versioning-strategy`, a breaking change is a MAJOR bump even below 1.0.0**, so
  `0.1.0` → `1.0.0`.
- `docs/architecture/cli-architecture-v2-release-split.md` explicitly records that "release version
  remains a separate decision while release PR #16 (`1.0.0`) is open. No tag or GitHub prerelease
  exists or is authorized." So the `0.1.0`-vs-`1.0.0` choice is a **deferred human gate**, not a
  bug.

**Resolution taken in this doc:** treat `0.1.0` as authoritative, flag `1.0.0` on PR #16 as the
discrepancy, and document (below) the minimal change that would realign the first release to
`0.1.0` — **without applying it**, because (a) a release PR already exists so no enabling change is
required, and (b) the final version is an explicit human gate. A maintainer applies the change
deliberately if `0.1.0` is confirmed as the intended first version.

### Minimal change to make the first release `0.1.0` (documented, NOT applied)

Pick one; **Option A** is the minimal, most explicit:

- **Option A — force the exact version (one line).** Add `"release-as": "0.1.0"` to the `"."`
  package in `release-please-config.json`:

  ```jsonc
  "packages": {
    ".": {
      "package-name": "pm",
      "changelog-path": "CHANGELOG.md",
      "release-as": "0.1.0"
    }
  }
  ```

  On the next push to `main`, release-please rewrites PR #16 to `chore: release 0.1.0`.
  `release-as` is **sticky** — remove it in the follow-up commit after `0.1.0` is cut, otherwise
  every subsequent release also tries to release `0.1.0`.

- **Option B — adopt pre-1.0 semantics (correct long-term).** Set the manifest baseline to
  `{ ".": "0.0.0" }` and add `"bump-minor-pre-major": true` + `"bump-patch-for-minor-pre-major":
  true` to the `"."` package. Below `1.0.0` this maps breaking → minor and `feat` → patch, so the
  accumulated commits yield `0.1.0` as the first release and future breaking changes stay in the
  `0.x` line until a deliberate `1.0.0`. Heavier than Option A and it edits the (authoritative)
  manifest, but it fixes versioning for all future pre-1.0 releases, not just this one.

- **Option C — one-time footer (idiomatic).** Land a commit on `main` whose message body contains
  `Release-As: 0.1.0`. One-shot, no config residue, but requires a dedicated commit.

### Exact steps to cut `0.1.0` (once a human confirms `0.1.0`)

1. Apply the minimal change (Option A recommended) via a normal PR to `main`; merge it.
2. On the next `main` push, confirm release-please updated the release PR title to
   **`chore: release 0.1.0`** and that its `CHANGELOG.md` reads `## [0.1.0]`.
3. **Merge the `chore: release 0.1.0` PR** (human gate) → creates tag `v0.1.0` + GitHub release.
4. If Option A was used, land a tiny follow-up PR removing `"release-as": "0.1.0"` so future
   releases compute normally.
5. Confirm `release-assets` attached 6 archives + `checksums.txt`, and a downloaded binary reports
   `pm version` → `0.1.0`.

> Do not perform steps 3–5 without an explicit human release decision. This doc authorizes no tag
> and no publish.

---

## 5. The connectors-as-data model

Connectors are **config-driven data bundles**, not bespoke Go per connector. A connector lives at
`internal/connectors/defs/<name>/` and (mirroring `internal/connectors/defs/gong/`) contains:

| File | Purpose |
| --- | --- |
| `metadata.json` | Connector identity, auth spec, config fields. |
| `spec.json` | Connector configuration schema (the fields a user supplies via config). |
| `api_surface.json` | The full official provider surface; every endpoint is `covered_by` a stream / direct_read / write / bounded-binary, or explicitly blocked. |
| `cli_surface.json` | The validated CLI command surface generated from the API surface. |
| `streams.json` | Stream-backed (ETL) read operations. |
| `operations.json` | Direct-read / typed operation definitions. |
| `writes.json` | Named reverse-ETL write actions (no generic raw writes). |
| `docs.md` | Rendered connector help/docs. |
| `schemas/`, `fixtures/` | Record schemas and fixture-backed request/response tests. |

Every connector must be **usable locally via config** — you point it at a target with connector
config fields and inspect it with `pm connectors inspect <name> --json` (which does not read
credentials). Authoring rules and the engine dialect are in `docs/migration/conventions.md`.

Self-verify a connector bundle before calling it done:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/<name>
go build ./internal/connectors/... && go vet ./internal/connectors/...
go test ./internal/connectors/conformance -run 'TestConformance/<name>'
```

---

## 6. The standard connector-add issue structure

Adding a connector is delivered as **one parent issue + a fixed set of sub-issues**, all labeled
`connector-cli` (the default label in `.github/ISSUE_TEMPLATE/agent_task.yml`). This mirrors the
gong roadmap: parent **#133 "Gong CLI feature parity parent roadmap"** with sub-issues
**#141–#147** and follow-ons **#252–#254**.

**Parent issue** — sections: Objective · Official API surface research baseline · App operation
candidate breakdown · Full-surface safety target · Sub-issues table · Acceptance criteria · Human
gates. It links every sub-issue.

**Sub-issue set** (each: `Parent: #<parent>` header · Objective · Branching · Full-surface safety
rule · Acceptance criteria · `Refs #<parent>` footer; labeled `connector-cli`):

| Lane | gong reference | Intent |
| --- | --- | --- |
| CLI surface metadata | #141 | Produce validated `cli_surface.json` from the provider surface. |
| Help renderer | #142 | Render connector help/docs/website snippets from metadata. |
| Stream runner | #143 | Execute stream-backed (ETL) reads. |
| Operation ledger | #144 | Reconcile the full `api_surface.json` coverage. |
| Direct read | #145 | Bounded JSON GET/query direct reads. |
| Advanced query / binary engine | #146 | Model query bodies and binary upload/download safely. |
| Sensitive / admin policy | #147 | Write risk tiers, redaction, approval, typed confirmation. |
| Typed POST read-query execution | #252 | Schema-gated POST reads (no raw body escape hatch). |
| Top-level JSON array bodies | #253 | Schema-gated array request bodies. |
| Bounded multipart upload | #254 | Typed, bounded multipart upload support. |

Each new connector reuses this exact shape. In-flight examples authored to this template:
**WhatsApp** (WhatsApp Business API) and **Bahmni-docker** (config-driven inspection of a local
Bahmni deployment). See the issue tracker for their parent + sub-issue numbers.

Because each connector lands as `feat:` commits, completing one advances the next release by a
minor version (see §3).
