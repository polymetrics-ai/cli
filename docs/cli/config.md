```
NAME
  pm help config - configuration reference

SYNOPSIS
  pm help config
  pm <command> --root <path> [--json] [--plain] [--no-input] [--progress ndjson]

DESCRIPTION
  pm resolves typed invocation configuration once per CLI run. The loader uses a
  fresh Viper instance for each invocation and never uses the package-level Viper
  singleton, AutomaticEnv, or file watching. Root and json affect invocation
  bootstrap. Runtime, RLM, schedule, and telemetry keys are consumed by migrated
  non-secret call sites. Worker/RLM agent Temporal execution remains opt-in:
  runtime.temporal_addr must be explicitly set by env or config file for those
  paths, while runtime doctor keeps local Compose defaults. OpenTelemetry
  tracing and metrics are default-off and exit-code neutral. Malformed
  .polymetrics/config.yaml files fail as validation errors.

PRECEDENCE
  1. Bound global config flags: --root and --json.
  2. Explicit POLYMETRICS_* environment variables.
  3. Documented PM_* legacy aliases when the primary POLYMETRICS_* variable is
     not set.
  4. .polymetrics/config.yaml under the invocation project root.
  5. Built-in defaults.

UI AND PROGRESS FLAGS
  --plain
    Force the plain output path. Use this in scripts, pipes, and tests that
    need deterministic non-interactive behavior.

  --no-input
    Disable prompts and TTY UI. Interactive-only paths must fail by naming the
    flag or file to provide instead of prompting.

  --progress ndjson
    Stream sanitized progress events to stderr as newline-delimited JSON.
    Stdout remains reserved for the command's final output or single JSON
    envelope. Supported value: ndjson. On failures, stderr may also include the
    final error diagnostic after progress events.

  Flow and ETL run dashboards render only when stdin and stdout are TTYs.
  --json, --plain, --no-input, non-empty PM_NO_TUI, non-empty CI, TERM=dumb,
  non-TTY stdin, and non-TTY stdout force the plain path.

TELEMETRY
  telemetry.exporter defaults to none (off is accepted as a disabled alias). No
  SDK is constructed and no .polymetrics/telemetry directory is created while
  disabled. Set PM_TELEMETRY=file or POLYMETRICS_TELEMETRY=file to write
  stdouttrace JSONL spans and stdoutmetric JSONL metrics under
  telemetry.directory. Network OTLP tracing/metrics and any custom collector
  endpoint must be selected from trusted env/flag sources; config-file OTLP
  exporter or endpoint values are ignored. Set PM_TELEMETRY=otlp and configure
  OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT,
  OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, or PM_TELEMETRY_ENDPOINT for OTLP
  HTTP/protobuf. OTEL_SDK_DISABLED=true always disables tracing and metrics.

  Telemetry failures are warnings on stderr, not command failures. Stdout keeps
  the command's normal human output or single JSON envelope. Span attributes are
  allowlisted; connector HTTP spans record method, scheme, host, path, status,
  and retry/attempt metadata only. Metrics use batched counters such as
  pm.records.* and pm.batches.flushed, flushed at batch boundaries instead of
  per record. Query strings, request/response bodies, headers, raw argv, and
  credential values are never recorded. Set telemetry.capture=minimal to strip
  span attributes while keeping span names.

CONFIG FILE
  The config file path is <project-root>/.polymetrics/config.yaml. Missing files
  are allowed. The root key in a config file does not relocate config-file
  discovery for the same invocation; use --root, POLYMETRICS_ROOT, or PM_ROOT to
  select a different project root before the file is read. If a file is
  malformed, --json, POLYMETRICS_JSON=true, or PM_JSON=true selects the same
  single JSON Error envelope on stdout used by other validation errors.

  Example:

    version: 1
    project: polymetrics-local
    warehouse:
      connector: warehouse
      path: .polymetrics/warehouse
    runtime:
      postgres_url: postgres://localhost:15433/polymetrics?sslmode=disable
      dragonfly_addr: localhost:6379
      temporal_addr: localhost:7233
    rlm:
      image: ghcr.io/polymetrics/rlm-agent:latest
      podman_bin: podman
      fake_runner: false
      embedded_worker: false
      llm:
        provider: openrouter
        base_url: https://openrouter.ai/api/v1
        model: ""
    schedule:
      crontab_file: ""
    telemetry:
      exporter: none
      endpoint: ""
      directory: .polymetrics/telemetry
      capture: default

KEYS
  root
    Default: invocation root (.). Primary env: POLYMETRICS_ROOT. Alias: PM_ROOT.
    Flag: --root.

  json
    Default: false. Primary env: POLYMETRICS_JSON. Alias: PM_JSON. Flag: --json.

  version
    Default: 1. Primary env: POLYMETRICS_VERSION. Alias: PM_VERSION.

  project
    Default: polymetrics-local. Primary env: POLYMETRICS_PROJECT. Alias:
    PM_PROJECT.

  warehouse.connector
    Default: warehouse. Primary env: POLYMETRICS_WAREHOUSE_CONNECTOR. Alias:
    PM_WAREHOUSE_CONNECTOR.

  warehouse.path
    Default: .polymetrics/warehouse. Primary env: POLYMETRICS_WAREHOUSE_PATH.
    Alias: PM_WAREHOUSE_PATH.

  runtime.postgres_url
    Default: local Compose PostgreSQL DSN. Primary env: POLYMETRICS_POSTGRES_URL.
    Alias: PM_POSTGRES_URL. Command output redacts PostgreSQL userinfo.

  runtime.dragonfly_addr
    Default: localhost:6379. Primary env: POLYMETRICS_DRAGONFLY_ADDR. Alias:
    PM_DRAGONFLY_ADDR.

  runtime.temporal_addr
    Default: localhost:7233. Primary env: POLYMETRICS_TEMPORAL_ADDR. Alias:
    PM_TEMPORAL_ADDR.

  rlm.image
    Default: ghcr.io/polymetrics/rlm-agent:latest. Primary env:
    POLYMETRICS_RLM_IMAGE. Alias: PM_RLM_IMAGE.

  rlm.podman_bin
    Default: podman. Primary env: POLYMETRICS_PODMAN_BIN. Alias:
    PM_PODMAN_BIN.

  rlm.fake_runner
    Default: false. Primary env: POLYMETRICS_RLM_FAKE_RUNNER. Alias:
    PM_RLM_FAKE_RUNNER.

  rlm.embedded_worker
    Default: false. Primary env: POLYMETRICS_RLM_EMBEDDED_WORKER. Alias:
    PM_RLM_EMBEDDED_WORKER.

  rlm.llm.provider
    Default: openrouter. Primary env: POLYMETRICS_LLM_PROVIDER. Alias:
    PM_LLM_PROVIDER.

  rlm.llm.base_url
    Default: https://openrouter.ai/api/v1. Primary env: POLYMETRICS_LLM_BASE_URL.
    Alias: PM_LLM_BASE_URL.

  rlm.llm.model
    Default: empty. Primary env: POLYMETRICS_LLM_MODEL. Alias: PM_LLM_MODEL.

  schedule.crontab_file
    Default: empty. Primary env: POLYMETRICS_CRONTAB_FILE. Alias:
    PM_CRONTAB_FILE. Intended for local scheduler redirection and tests.

  telemetry.exporter
    Default: none. Primary env: POLYMETRICS_TELEMETRY. Alias: PM_TELEMETRY.
    Supported values: none, off, file, otlp. off is a disabled alias.

  telemetry.endpoint
    Default: empty. Primary env: POLYMETRICS_TELEMETRY_ENDPOINT. Aliases:
    PM_TELEMETRY_ENDPOINT, POLYMETRICS_OTEL_EXPORTER_OTLP_ENDPOINT,
    OTEL_EXPORTER_OTLP_ENDPOINT, and OTEL_EXPORTER_OTLP_TRACES_ENDPOINT. OTLP
    endpoints must be http/https URLs without userinfo, query strings, or
    fragments. Custom OTLP endpoints must come from a trusted env/flag source;
    config-file endpoint values alone are ignored, and OTLP otherwise uses the
    local default collector endpoints. For a separate metrics endpoint, set the
    env-only OTEL_EXPORTER_OTLP_METRICS_ENDPOINT; it is validated and sanitized
    the same way and is never read from config.yaml.

  telemetry.directory
    Default: .polymetrics/telemetry. Primary env: POLYMETRICS_TELEMETRY_DIR.
    Alias: PM_TELEMETRY_DIR. Paths must be relative, stay under --root, and must
    not traverse symlinked telemetry directories or files.

  telemetry.capture
    Default: default. Primary env: POLYMETRICS_TELEMETRY_CAPTURE. Alias:
    PM_TELEMETRY_CAPTURE. Supported values: default, minimal.

SECURITY
  Configuration is an allowlist. pm does not ingest arbitrary POLYMETRICS_* or
  PM_* variables. User-named credential env vars supplied to --from-env and
  connector certification credsfile entries are credential data, not app config.
  Do not store secret values in config.yaml or examples. LLM API keys such as
  PM_LLM_API_KEY and provider-specific keys remain environment-only secret
  inputs and are not documented with values. OTLP endpoint URLs with userinfo,
  query strings, or fragments are rejected; ambient OTLP header/TLS/compression
  env vars are warned and neutralized before exporter construction. Emitted
  spans and metrics drop userinfo, query strings, headers, bodies, raw argv, and
  credential values.

EXIT STATUS
  0 success
  3 validation error, including malformed config or invalid UI/progress flag

```
