# polymetrics RLM agent (PI mono + Python ML)

The containerized agent that `pm rlm run --mode agent` (and `pm extract` for
complicated requests) runs, orchestrated durably via Temporal by `pm worker serve`.

It runs the previous polymetrics RLM (`rlm_ruby`) pipeline ported to Python:

```
classify → generate Python (pandas/polars/scikit-learn) → execute
        → validate (query-validates-query via validate.py, NO LLM) → reflect → retry (≤ PM_RLM_MAXITER)
```

## Runtime contract

| Path | Meaning |
|------|---------|
| `/work/in/input.ndjson` | warehouse-enveloped input `{"_polymetrics_raw_id":..,"record":{..}}` (read-only mount) |
| `/work/in/request.json` | `{request, spec, in_table, out_table}` |
| `/work/out/output.ndjson` | flat scored rows (`_polymetrics_raw_id` + `_rlm_score` in [0,1]), written atomically on success |
| `/work/out/manifest.json` | `{expected_count, records_read}` — the host asserts the row count to reject truncation |

Exit codes: `0` ok · `3` reflection-exhausted · `4` LLM unreachable. The host
maps these to non-retryable Temporal errors.

## LLM provider (env)

Provider-agnostic via pi-ai. Default OpenRouter.

- `PM_LLM_PROVIDER` (default `openrouter`), `PM_LLM_MODEL`
- API key: `OPENROUTER_API_KEY` / `PM_LLM_API_KEY` (injected via `--env-file`, never baked in the image)
- Subscription OAuth (opt-in, flagged): mount `~/.pi/agent/auth.json` for Anthropic Pro/Max or Codex/ChatGPT
- `PM_RLM_MAXITER` (default 4)

## Build & run (HUMAN-GATED)

Building/publishing the image and adding Podman are human gates (see the Phase 6
plan §Human gates). Once approved:

```bash
pm agent image build          # podman build -f build/agent/Containerfile -t <image> build/agent
# or: podman build -t ghcr.io/polymetrics/rlm-agent:latest build/agent
```

The host runs the container default-deny (`--network=none`, `--read-only`,
`--cap-drop=ALL`, `--security-opt=no-new-privileges`, mem/pid limits, `--rm`,
deterministic `--name`); the LLM endpoint is reached only via a host egress
allowlist proxy. The agent never sees credentials beyond the injected
`PM_LLM_*` env / mounted auth file, and the container is the trust boundary.

## Verifying packages before build

`@earendil-works/pi-ai@0.80.2` and `@earendil-works/pi-agent-core@0.80.2` are the
current packages (the `@mariozechner/*` names are deprecated). Re-confirm the
`Agent` / `AgentTool` / `getModel` API against the installed versions if you bump.
