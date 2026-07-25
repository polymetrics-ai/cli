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

| Commit prefix | Semantic bump (>= 1.0.0) |
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
5. **Verify**: the GitHub release should list 6 archives + `checksums.txt`, and
   `pm version` from a downloaded binary should report the new tag.

No manual `git tag`, `gh release create`, or `goreleaser release` is ever required or authorized.

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
