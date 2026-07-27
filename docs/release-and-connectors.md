# Preparing the `pm` v0.1.0 binary release and connector issue work

This is the operator/authoring recipe for two recurring tasks in this repo:

1. Preparing a **`pm` binary release** through release-please + GoReleaser.
2. Adding a connector as a parent issue plus standard sub-issues.

Everything below is derived from the actual repo config: `release-please-config.json`,
`.release-please-manifest.json`, `.goreleaser.yaml`, `.github/workflows/release.yml`,
`scripts/verify-release-assets.sh`, and `.github/ISSUE_TEMPLATE/agent_task.yml`. If those files
change, update this doc in the same PR.

## Product boundary: binary release and website release are independent

The `pm` binary and the website are independent products. A `pm` binary release must **not** dispatch,
trigger, deploy, or otherwise publish the website. Website checks and deploys stay owned by
`.github/workflows/website.yml`; binary release preparation stays owned by `.github/workflows/release.yml`.

This document authorizes no tag, no GitHub Release, no website workflow dispatch, no website deploy,
and no merge. It prepares the binary release path only.

## 1. How `pm` binary releases work

Binary releases are automated by two tools wired through `.github/workflows/release.yml`.

### release-please computes the release PR

- Config: `release-please-config.json` — `release-type: go`, `package-name: pm`,
  `changelog-path: CHANGELOG.md`, `include-component-in-tag: false`, and release PR titles such as
  `chore: release ${version}`.
- Manifest: `.release-please-manifest.json` — the tracked last released version for the root package.
- Bootstrap: `bootstrap-sha: e28260ffb333951c0882f5e6535655c78fcc4610` — release-please only
  considers commits after this SHA when computing the next version.

On every push to `main`, the `release-please` job opens or updates the release PR. Merging that
release PR is the human-gated action that creates the git tag and GitHub Release.

### GoReleaser builds and attaches the binaries

- Config: `.goreleaser.yaml` — `version: 2`, `project_name: pm`, `release.mode: keep-existing`.
- Build: `builds[0]` compiles `./cmd/pm` with `CGO_ENABLED=0` across six targets:
  - macOS amd64: `pm_<version>_darwin_amd64.tar.gz`
  - macOS arm64: `pm_<version>_darwin_arm64.tar.gz`
  - Linux amd64: `pm_<version>_linux_amd64.tar.gz`
  - Linux arm64: `pm_<version>_linux_arm64.tar.gz`
  - Windows amd64: `pm_<version>_windows_amd64.zip`
  - Windows arm64: `pm_<version>_windows_arm64.zip`
- Checksum: `checksums.txt` covers exactly those six archives.
- Version injection: ldflags stamp the tag into `pm version` through
  `polymetrics.ai/internal/cli.version`, with commit and build date metadata.

`scripts/verify-release-assets.sh` is the release-asset guardrail. It checks the exact archive set,
archive contents (`LICENSE`, `NOTICE`, `README.md`, `pm` binary), and checksum coverage. Keep that
script aligned with `.goreleaser.yaml`.

### Release workflow jobs

| Job | Trigger | Purpose |
| --- | --- | --- |
| `release-please` | push to `main` / `workflow_dispatch` | Opens or updates the release PR; on release PR merge, creates the tag and GitHub Release. |
| `package-check` | `pull_request` | Runs `goreleaser release --snapshot --clean` and `scripts/verify-release-assets.sh dist`, so every PR proves the binary asset shape before merge. |
| `release-assets` | `release: published` or `release-please.outputs.release_created == 'true'` | Checks out the release tag, rebuilds verified assets, and uploads the six archives plus `checksums.txt` to the published GitHub Release. |

There is intentionally no job that dispatches or deploys the website from a `pm` binary release.

## 2. Making the next `pm` release `v0.1.0`

The next binary release must be `v0.1.0`, not `v1.0.0`. Use Release Please's supported one-shot
override: a commit body containing `Release-As: 0.1.0`.

Example commit shape:

```text
ci(release): prepare pm v0.1.0 binary release

Release-As: 0.1.0
```

Release Please documents that when a commit on `main` has `Release-As: x.x.x` in the commit body, it
opens or updates a release PR for that specified version. This is a one-time override; do **not** add
a persistent `release-as` field to `release-please-config.json`, because that would keep forcing later
releases back to `0.1.0`.

Merge-method caution: if a maintainer squash-merges this preparation PR, the squash commit body must
preserve `Release-As: 0.1.0`; otherwise Release Please will not see the one-shot override. A merge
commit or rebase merge preserves the branch commit message directly.

The actual `v0.1.0` release must not be cut until:

1. the mandatory Bahmni corrective PR is merged,
2. the exact intended release commit on `main` is green, and
3. the captain/maintainer approves the release PR merge.

No one should create the `v0.1.0` tag or publish release assets manually.

## 3. How to cut `v0.1.0` after approval

1. Merge the approved preparation PR that contains `Release-As: 0.1.0` in a commit body.
2. Wait for the `release-please` workflow on `main` to open or update the release PR to
   `chore: release 0.1.0`.
3. Review that release PR's generated `CHANGELOG.md` and `.release-please-manifest.json` update.
4. Human gate: captain/maintainer merges the release PR only after the intended release commit is
   green and the mandatory Bahmni corrective PR is included.
5. Confirm the published GitHub Release contains exactly six archives plus `checksums.txt`.
6. Download one archive and verify `pm version` reports `0.1.0`.

This preparation PR performs steps 1's setup only. It does not cut the release.

## 4. Connector definitions and binary patch releases

Connector definitions are embedded in the `pm` binary under `internal/connectors/defs/<name>/`.
There is currently no independent connector package, registry artifact, or connector-specific
version stream. A released `pm` binary contains exactly the connector definitions merged into the
release commit.

Consequences:

- The first `v0.1.0` binary includes only connectors already merged into the exact release commit.
- A compatible connector fix after `v0.1.0` ships in the next `pm` patch release, for example
  `v0.1.1`.
- A new connector or user-facing connector feature normally drives the next pre-1.0 minor release,
  for example `v0.2.0`, unless a maintainer deliberately applies a one-shot version override.
- Do not claim an in-flight connector is included until its PR is merged into the release commit.
  In particular, WhatsApp must not be described as part of `v0.1.0` while its PR is unmerged.

## 5. The connectors-as-data model

Connectors are config-driven data bundles, not bespoke Go per connector. A connector lives at
`internal/connectors/defs/<name>/` and, mirroring `internal/connectors/defs/gong/`, contains:

| File | Purpose |
| --- | --- |
| `metadata.json` | Connector identity, auth spec, and config fields. |
| `spec.json` | Connector configuration schema. |
| `api_surface.json` | Full official provider surface; every endpoint is covered or explicitly blocked. |
| `cli_surface.json` | Validated CLI command surface generated from the API surface. |
| `streams.json` | Stream-backed ETL read operations. |
| `operations.json` | Direct-read and typed operation definitions. |
| `writes.json` | Named reverse-ETL write actions; no generic raw writes. |
| `docs.md` | Rendered connector help/docs. |
| `schemas/`, `fixtures/` | Record schemas and fixture-backed request/response tests. |

Every connector must be usable locally via config and inspectable with
`pm connectors inspect <name> --json`, which does not read credentials. Authoring rules and the
engine dialect are in `docs/migration/conventions.md`.

Self-verify a connector bundle before calling it done:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/<name>
go build ./internal/connectors/...
go vet ./internal/connectors/...
go test ./internal/connectors/conformance -run 'TestConformance/<name>'
```

## 6. Standard connector-add issue structure

Adding a connector is delivered as one parent issue plus a fixed set of sub-issues, all labeled
`connector-cli`. This mirrors the gong roadmap: parent #133 with sub-issues #141-147 and follow-ons
#252-254.

Parent issue sections:

- Objective
- Official API surface research baseline
- App operation candidate breakdown
- Full-surface safety target
- Sub-issues table
- Acceptance criteria
- Human gates

Sub-issue shape:

- `Parent: #<parent>` header
- Objective
- Branching
- Full-surface safety rule
- Acceptance criteria
- `Refs #<parent>` footer

| Lane | Gong reference | Intent |
| --- | --- | --- |
| CLI surface metadata | #141 | Produce validated `cli_surface.json` from the provider surface. |
| Help renderer | #142 | Render connector help/docs/website snippets from metadata. |
| Stream runner | #143 | Execute stream-backed ETL reads. |
| Operation ledger | #144 | Reconcile full `api_surface.json` coverage. |
| Direct read | #145 | Bounded JSON GET/query direct reads. |
| Advanced query / binary engine | #146 | Model query bodies and binary upload/download safely. |
| Sensitive / admin policy | #147 | Write risk tiers, redaction, approval, and typed confirmation. |
| Typed POST read-query execution | #252 | Schema-gated POST reads with no raw body escape hatch. |
| Top-level JSON array bodies | #253 | Schema-gated array request bodies. |
| Bounded multipart upload | #254 | Typed, bounded multipart upload support. |

Each new connector reuses this exact shape. Because connector definitions ship embedded in the
binary, connector issue completion changes the next `pm` release according to the Conventional
Commit type that lands on `main`.
