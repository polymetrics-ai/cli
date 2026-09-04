# Polymetrics — Setup & Usage Guide

Everything you need to install Polymetrics, understand its model, and run real
extract → query → write-back workflows. For the elevator pitch see the
[README](../README.md); this is the hands-on reference.

🌐 Homepage: **[polymetrics.ai](https://polymetrics.ai)**

- [Install & build](#install--build)
- [Core concepts](#core-concepts)
- [Workflow 1 — ETL (extract)](#workflow-1--etl-extract)
- [Workflow 2 — Query & analyze (DuckDB)](#workflow-2--query--analyze-duckdb)
- [Workflow 3 — Reverse-ETL (write back / take action)](#workflow-3--reverse-etl-write-back--take-action)
- [Connectors](#connectors)
- [Driving Polymetrics from an AI agent](#driving-polymetrics-from-an-ai-agent)
- [Optional runtime services](#optional-runtime-services)
- [Troubleshooting](#troubleshooting)
- [Contributing a connector](#contributing-a-connector)

---

## Install & build

**Build prerequisites:** Go **1.25.11+**. (The `Makefile` sets `GOTOOLCHAIN=auto`,
so `make` targets auto-fetch the right toolchain even on an older local Go.)

### Install with `go install`

```bash
go install polymetrics.ai/cmd/pm@latest      # installs the `pm` binary into $(go env GOBIN)
```

The module path is `polymetrics.ai`; the binary is always named `pm` (it builds from
`cmd/pm`). This resolves once `polymetrics.ai` serves its Go module meta tag — until then,
build from source below.

### Install a release binary

Release assets are published from `polymetrics-ai/cli` for Linux and macOS on
amd64 and arm64. Windows is not published; see Troubleshooting.

This path requires the GitHub CLI (`gh`) and standard archive tools (`tar` on
macOS/Linux), but does not require Go.

```bash
os_name="$(uname -s)"
arch_name="$(uname -m)"

case "$os_name" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "unsupported OS: $os_name" >&2; exit 1 ;;
esac

case "$arch_name" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $arch_name" >&2; exit 1 ;;
esac

asset_pattern="pm_*_${os}_${arch}.tar.gz"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
gh release download --repo polymetrics-ai/cli --pattern "$asset_pattern" --dir "$tmpdir"

tar -xzf "$tmpdir"/pm_*_"${os}"_"${arch}".tar.gz -C "$tmpdir"
binary_name=pm

install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"
cp "$tmpdir/$binary_name" "$install_dir/$binary_name"
chmod +x "$install_dir/$binary_name" 2>/dev/null || true
"$install_dir/$binary_name" version
```

Each future release also publishes `checksums.txt`, keyless Sigstore bundles, GitHub artifact attestations, and standalone Linux `.deb`/`.rpm` packages; see [release verification](release-verification.md).

### Build from source

```bash
git clone https://github.com/polymetrics-ai/cli
cd cli
make build          # produces ./pm  (pure-Go, CGO-free)
```

Put it on your PATH:

```bash
make install                       # installs to ~/.local/bin/pm
make install INSTALL_DIR=/usr/local/bin
pm help
```

### Build with DuckDB analytics (optional)

DuckDB is embedded in every build. It is the query engine and the only Parquet
implementation in the binary, so building `pm` requires cgo and a C toolchain:

```bash
CGO_ENABLED=1 go build -o pm ./cmd/pm
```

There is no build tag and no CGO-free variant. Two builds that wrote different
table formats would mean a query that works on one install and fails on another,
so `pm query` supports read-only `SELECT` and `WITH` in full everywhere.

### Verify your build

```bash
make verify          # gofmt + go vet + go test ./... + build + end-to-end smoke
```

---

## Core concepts

| Concept | What it is |
|---|---|
| **Project** | A `.polymetrics/` directory created by `pm init`. Holds state, the encrypted vault, and the local warehouse. Target a directory with `--root <dir>` (defaults to the current directory). |
| **Credential** | A named, encrypted set of secrets + config for one connector (`pm credentials add …`). Secrets are stored AES-GCM encrypted and never printed. |
| **Connector** | A system Polymetrics can read from and/or write to (`github`, `stripe`, `postgres`, …). One package per system. |
| **Connection** | A configured source → destination pairing with stream/primary-key/cursor/table mapping. |
| **Catalog** | The set of streams a source exposes (`pm catalog refresh`). |
| **Warehouse** | Local Parquet tables rebuilt from append-only JSONL write-ahead logs and queried by DuckDB. |
| **Sync mode** | A connection's materialization setting. See [ETL sync modes](cli/etl.md) for accepted names, capability requirements, and typed-compatibility admission. |

### Adding credentials without leaking secrets

Never paste secrets on the command line. Use one of:

```bash
# from an environment variable
export GITHUB_TOKEN=ghp_…
pm credentials add gh --connector github --config owner=OWNER --config repo=REPO --from-env token=GITHUB_TOKEN

# from stdin (e.g. a private key file)
pm credentials add gh-app --connector github \
  --config owner=OWNER --config repo=REPO --config auth_type=github_app \
  --config app_id=12345 --config installation_id=67890 \
  --value-stdin private_key < app-private-key.pem
```

Connector runtime commands use declared execution JSON and ordinary credential or approval boundaries. Deterministic test fixtures are test-harness inputs, not a generic user-facing fixture-valued connector configuration.

---

## Workflow 1 — ETL (extract)

Pull data from a source into the local warehouse. Example with a public GitHub repo
(no token needed for public reads):

```bash
pm init
pm credentials add github    --connector github    --config owner=octocat --config repo=Hello-World --config public_access=true
pm credentials add warehouse --connector warehouse --config path=.polymetrics/warehouse

pm connections create gh \
  --source github:github --destination warehouse:warehouse \
  --stream issues --primary-key id --cursor updated_at --table github_issues

pm catalog refresh --connection gh --json          # discover streams
pm etl run --connection gh --stream issues --json  # extract
```

**Incremental & bounded runs.** Connectors checkpoint a cursor between runs.
For large streams, bound the work:

```bash
pm etl run --connection gh --stream pull_requests --batch-size 100 --json
# HTTP connectors default to one page for safe local runs; exhaust a stream with:
pm credentials add github --connector github --config owner=OWNER --config repo=REPO --config public_access=true --config max_pages=0
```

The same pattern works for `stripe`, `postgres`, `slack`, and other connectors with
enabled read streams. Planned ledger-only connectors such as `hubspot` remain visible
in the catalog but list no executable streams until their future implementation lanes
ship. Use `pm connectors inspect <name>` to see a connector's streams, cursors, and
required config.

---

## Workflow 2 — Query & analyze (DuckDB)

Once data is in the warehouse, query it. With the **DuckDB build**, `pm query run`
runs real analytical SQL over your warehouse tables:

```bash
# simple table read (works in any build)
pm query run --table github_issues --limit 10 --json

# analytical SQL (DuckDB build) — joins, aggregations, window functions
pm query run --sql "
  SELECT user_login, COUNT(*) AS issues
  FROM github_issues
  GROUP BY user_login
  ORDER BY issues DESC
  LIMIT 10" --json
```

Queries are **read-only**: only `SELECT`/`WITH` statements are accepted;
`INSERT`/`UPDATE`/`DELETE`/DDL and statement chaining are rejected. This makes the
query surface safe to expose to an agent.

> Without the `duckdb` build tag, `pm query` supports simple table reads only.
> See [Build with DuckDB analytics](#build-with-duckdb-analytics-optional).

---

## Workflow 3 — Reverse-ETL (write back / take action)

Reverse-ETL turns warehouse rows into **writes against a destination** — upserting
records, or taking actions like opening a GitHub pull request. Every write goes
through `plan → preview → approve → execute`.

```bash
export GITHUB_TOKEN=ghp_…
pm credentials add github-write --connector github \
  --config owner=OWNER --config repo=REPO --from-env token=GITHUB_TOKEN

# 1. PLAN — describe the write; returns a plan id + one-time approval token + a sample
pm reverse plan prs_to_github \
  --source-table github_pr_candidates \
  --destination github:github-write \
  --action create_pull_request \
  --map title:title --map body:body --map head:head --map base:base \
  --json

# 2. PREVIEW — see exactly what would be written (no mutation)
pm reverse preview <plan-id> --json

# 3. EXECUTE — enter the approval token as one stdin line; nothing changes until then
pm reverse run <plan-id> --approval-token-stdin --json
```

Approval tokens are **single-use and time-bounded**. A `run` without a valid token
is rejected — this is the guardrail that lets an agent propose writes without being
able to perform them unsupervised.

**Write actions** are connector-specific and allow-listed. GitHub, for example,
supports `create_issue`, `update_issue`, `comment_issue`, `close_issue`,
`create_pull_request`, `update_pull_request`, `close_pull_request`,
`request_reviewers`, and `merge_pull_request`. Planned ledger-only connectors such
as HubSpot intentionally report no executable write actions until connector-owned
actions and evidence ship. Inspect a connector to see its actions:

```bash
pm connectors inspect github --json   # → manifest.write_actions
```

---

## Connectors

```bash
pm connectors list            # connectors compiled into your binary
pm connectors list --all      # the full catalog (implemented + planned)
pm connectors inspect stripe  # auth modes, streams, sync modes, write actions
pm connectors inspect stripe --json
```

The generated connector catalog is the source of truth for current connector counts and
capabilities. A few examples:

- **GitHub** (`github`) — stream reads plus approval-gated fixed REST and GraphQL writes.
  Set `public_access` for unauthenticated public reads; private or higher-rate-limit
  reads use a classic/fine-grained PAT, OAuth token, Actions `GITHUB_TOKEN`, an
  installation token, or a GitHub App (auto-signs a JWT → installation token). Use
  `pm connectors inspect github --json` and the generated GitHub connector manual for
  current execution and lane availability.
- **Stripe** (`stripe`) — Bearer (secret key) auth, cursor pagination, core CRM/billing
  streams, plus approval-gated customer create/update/delete writes. Run
  `pm connectors inspect stripe --json` for the current action list and destructive confirmation notes.
- **Postgres** (`postgres`) — connects via `pgx`; discovers schemas/columns; snapshot +
  cursor-incremental reads. Logical-replication CDC is on the roadmap.
- Most SaaS connectors are read-first; they add approval-gated writes where the upstream
  API has safe, well-defined mutations.

Use `pm connectors inspect <name> --json` to discover the connector's declared executable surfaces; do not rely on a credential-free connector simulation.

---

## Driving Polymetrics from an AI agent

Polymetrics is designed to be operated by an LLM agent. The contract:

1. **JSON everywhere.** Append `--json` to any command. The envelope carries
   `api_version`, a `kind`, and typed fields. Data → stdout; logs/progress → stderr.
2. **Branchable exit codes** (no output parsing needed to decide next steps):

   | Code | Category | Meaning |
   |---:|---|---|
   | 0 | — | success |
   | 2 | usage | bad command/flags (don't retry) |
   | 3 | validation | bad input (fix args) |
   | 4 | auth | missing/invalid credentials |
   | 5 | connector | connector/API error |
   | 6 | runtime | local dependency unavailable |
   | 7 | policy | blocked (e.g. approval required) |
   | 1 | internal | unexpected error |

3. **Gated writes.** Reverse-ETL is always `plan → preview → approve → execute`, so an
   agent can propose changes and surface a diff without mutating anything until approved.
4. **Generated agent docs.** Produce machine-readable command docs and per-connector
   skill files:

   ```bash
   pm docs generate --dir docs/cli
   pm skills generate --dir docs/skills --json
   ```

Any command group invoked with `--json` and no subcommand returns its manual as a
JSON envelope — agents can discover capabilities at runtime.

---

## Optional runtime services

The core CLI is fully local and needs nothing else. An optional Compose stack
(PostgreSQL for a run-ledger, DragonflyDB for coordination, Temporal for durable
workflows) is available for advanced setups. Runtime preference is Podman first,
Docker fallback.

```bash
scripts/setup-runtime-macos.sh      # or setup-runtime-linux.sh
make runtime-up                     # start services
make runtime-doctor                 # health check
make runtime-down
```

| Service | Endpoint |
|---|---|
| PostgreSQL | `localhost:15433` |
| DragonflyDB | `localhost:6379` |
| Temporal gRPC | `localhost:7233` |
| Temporal UI | `http://localhost:8080` |

Details: [docs/runtime/SETUP.md](runtime/SETUP.md).

---

## Troubleshooting

- **`go.mod requires go >= 1.25` / toolchain errors** — use `make` targets (they set
  `GOTOOLCHAIN=auto`), or run `GOTOOLCHAIN=auto go build ./cmd/pm`, or install Go 1.25.11+.
- **DuckDB build fails to link** — `pm` needs CGO and a C compiler (`CGO_ENABLED=1`,
  plus gcc or clang). There is no CGO-free build: DuckDB reads and writes every
  warehouse table.
- **No Windows release** — `windows/arm64` has no prebuilt DuckDB library and cannot be
  built; `windows/amd64` was dropped for having no user asking for it. `pm` still builds
  from source on Windows amd64 with a C toolchain, and WSL works today. Both targets
  return on a customer ask.
- **An HTTP connector only returns one page** — connectors default to one page for safe
  local runs. Set `max_pages=0` (aliases: `all`, `unlimited`) on the credential, or
  `--source-config max_pages=0` on the connection.
- **`connector "x" not found`** — run `pm connectors list` to see what's compiled in;
  `pm connectors list --all` shows planned (not-yet-implemented) catalog entries.
- **Secrets** — Polymetrics never prints secret values. If you need to rotate one,
  re-add the credential with `--from-env` or `--value-stdin`.

---

## Contributing a connector

Adding a connector is the highest-leverage contribution. The pattern:

1. **Author one source lock** — add
   `internal/connectors/defs/<name>/source.lock.json` using schema version 4,
   shared request/response/record schemas, canonical operations, and all seven
   lane declarations.
2. **Publish a closed execution generation** — use `connectorgen lock-render`;
   it stages and revalidates metadata, spec, streams/schemas, writes,
   operations, CLI surface, optional typed execution files, and publication
   metadata before atomically selecting one complete generation. Runtime still
   consumes execution JSON only.
3. **Use hooks only for real gaps** — add `internal/connectors/hooks/<name>/` when auth,
   pagination, fan-out, or write behavior cannot be represented declaratively.
4. **Promote full protocols to native** — use `internal/connectors/native/<name>/` only when
   a connector is inherently dynamic or too large for a hook. Native connectors are wired by
   `native/nativeset`, not package `init()` side effects.
5. **Test first** — use deterministic render checks, execution-only discovery,
   fake-server encoder tests, DuckDB saved-flow tests, and focused hook/native tests.
6. **Add the icon release artifact** — connector icons are registry-backed and validated.
   Run `PM_ICON_REGISTRY_SOURCE=<registry-json-url> make icons-generate` to seed icons from an upstream registry. If the seeded icon is stale,
   compare it against the vendor website or official documentation, replace the SVG under
   `docs/connectors/icons/`, and update the icon entry in `internal/connectors/icon_data.json`
   with `review_status` set to `official_verified` or `manual_override`. Assets under
   `docs/connectors/icons/simple-icons/` are the exception: they are fetched and checksum-pinned,
   so refresh them with `node scripts/fetch-simple-icons.mjs --update-lockfile` from `website/`
   instead of hand-editing. Registry keys are bare connector identifiers —
   `source-*`/`destination-*` keys are rejected — and every connector needs its own entry; see
   `docs/migration/icon-registry-single-source.md`.
7. **Validate generated output** — run `lock-render`, then confirm
   `lock-render --check` reads the exact selected generation without writing.
   `go run ./cmd/connectorgen validate internal/connectors/defs` must report
   zero findings. Run `go run ./cmd/connectorgen gen` only if hook/native
   package sets changed, then `make verify`.

```bash
PM_ICON_REGISTRY_SOURCE=<registry-json-url> make icons-generate
go run ./cmd/connectorgen lock-render <connector>
go run ./cmd/connectorgen lock-render <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen gen            # only when hook/native package sets change
make verify                              # must stay green
```

Optional local hook setup:

```bash
git config core.hooksPath .githooks
```

The hook and CI both route through the same docs/icon validation path, so a connector cannot
ship without icon metadata, a local SVG asset, generated docs, and safe SVG content.

Open a PR — and thank you. 🙌
