# Gong eight-surface live-certification runbook

This is a pre-execution runbook. It is deliberately paused: do not make a provider request until
the versioned-artifact-query foundation in PR #4335 is merged into `main` and the external-proof
command-privacy foundation #4337 is available. It never authorizes an agentic Gong endpoint,
a browser-session fallback, a generic HTTP command, or a provider mutation without a declaration-
owned self-cleaning pairing.

## Eight-surface contract

| Surface | Gong mapping | Planned clean-run evidence | Current limit |
| --- | --- | --- | --- |
| ETL | 12 declared streams | bounded full-harness append/overwrite/incremental classifications and record counts | five streams previously remained uncertified; no result is promoted from an HTTP classification alone |
| Direct read | 30 exact bounded commands | only the declared `gong_ordinary_rest_users_extensive` candidate runs live; the full built-binary preflight proves the broader command surface credential-free | paid `entities get-brief` and `entities ask` are never selected |
| Direct write | 27 fixed declaration-owned operations | preflight and unassessed inventory classification | no live mutation until a self-cleaning pairing exists |
| Reverse ETL | the same 27 named actions | plan → preview → approval → execute is checked without sending a provider write | no plan may be approved or applied for a record not created by this run |
| Binary download | zero official binary-response operations | exact-source `not_applicable` from `SOURCE-AUDIT.md` | none |
| Binary upload | three multipart operations: call media, CRM entities, target assignments | generic multipart conformance and unassessed live-write classification | each is applicable but has no disposable create/readback/cleanup pairing yet |
| Flow | application `flow_roundtrip`, with Gong as the declared source | plan order, side-effect-free preview, run, and status stages | not a claim that Gong exposes a separate flow-execution API |
| Schedule | application `schedule_roundtrip`, over the bounded Gong-backed flow | isolated-crontab create/list/install/fire/inspect/remove, byte-identical cleanup, no residue | not a claim that Gong exposes a scheduler API |

All three additions therefore have sensible mappings. Binary upload is provider-backed but not
yet safely live-testable; flow and schedule are application-workflow cells, not missing provider
operations. Those limits are certification outcomes, never grounds to relabel an implemented Gong
command as partial.

## Preconditions and evidence boundary

1. Confirm PR #4335 is merged and merge current `origin/main` without rebasing or discarding the
   preserved branch. Its provider-neutral `parseBatchArtifactURL` change must make the scoped
   source-import check pass for the immutable official versioned document.
2. The full harness must use its only declaration-owned ordinary-REST candidate and `--limit 1`.
   It must not add an agentic candidate, invoke either paid endpoint, or pass `--write`,
   `--write-only`, or `--full-parity`.
3. The captain-provisioned keychain entries are read only at point of use. Check byte lengths in
   shell predicates and do not echo them. The account-scoped base URL is supplied from the same
   approved non-printing runtime reference; it is never written to this repository, a status
   line, PR body, or a proof.
4. A proof-producing command cannot currently put `--config base_url=...` in its outer argv:
   `external_proof.go` serializes the process command verbatim. Wait for the provider-neutral
   sanitizer/config-loader foundation #4337 that replaces account-scoped argument values with safe
   fingerprints before the external-proof command below is run. Until then a local legacy report
   may be used only for diagnosis and cannot be published as certification evidence.
5. Use a new private root and retain it only until the safe aggregate is recorded. Keep the raw
   report and external proof outside git; record only counts, status classes, and SHA-256
   fingerprints in the phase ledger after the run.

## Exact post-foundation sequence

The variables below are placeholders, never committed values. `GONG_CERT_BASE_URL` is injected
out of band; its value must not occur in a command that a proof serializer preserves verbatim.
The `external-proof command sanitizer available` checkpoint is a hard gate, not a comment.

```sh
set -eu
git fetch origin
git merge --no-edit origin/main
git merge-base --is-ancestor origin/main HEAD

task_pm_bin="$PWD/.tmp/gong-cert-pm"
mkdir -p "$PWD/.tmp"
go build -o "$task_pm_bin" ./cmd/pm

go run ./cmd/connectorgen source-import gong --check
go run ./cmd/connectorgen validate internal/connectors/defs/gong --json
go run ./cmd/connectorgen certification-candidates --connector gong --check
go run ./cmd/connectorgen certification-sweep --connector gong --check
go run ./cmd/connectorgen certification-subject --pm "$task_pm_bin"
go run ./cmd/connectorgen certification-subject --pm "$task_pm_bin" --check

task_cert_root="$(mktemp -d "$PWD/.tmp/gong-cert-root.XXXXXX")"
chmod 700 "$task_cert_root"
task_access_key="$(security find-generic-password -s gong-api-access-key -w 2>/dev/null)"
task_access_secret="$(security find-generic-password -s gong-api-secret -w 2>/dev/null)"
[ "${#task_access_key}" -eq 32 ]
[ "${#task_access_secret}" -eq 152 ]
[ -n "${GONG_CERT_BASE_URL:-}" ]
export GONG_CERT_ACCESS_KEY="$task_access_key"
export GONG_CERT_ACCESS_SECRET="$task_access_secret"
unset task_access_key task_access_secret

# Hard gate: run only after the shared external-proof command sanitizer/config loader is present.
# It must fingerprint or omit the account-scoped base-url argument in the resulting proof.
set +e
"$task_pm_bin" --root "$task_cert_root" connectors certify gong --full --external-proof \
  --from-env access_key=GONG_CERT_ACCESS_KEY \
  --from-env access_key_secret=GONG_CERT_ACCESS_SECRET \
  --config "base_url=$GONG_CERT_BASE_URL" \
  --config certification_cohort=ordinary_rest \
  --limit 1 --json >"$task_cert_root/gong-report.json"
task_cert_exit=$?
set -e
unset GONG_CERT_ACCESS_KEY GONG_CERT_ACCESS_SECRET
```

`connectorgen surface-sync --check` remains a repository-wide gate, not a per-Gong command: it
requires the complete defs root and therefore must run later with the full repository gate rather
than against the invalid inner `gong` directory. Do not substitute that inner directory merely to
make the command look scoped.

Do not hide a nonzero `task_cert_exit`: it is expected until every applicable cell is actually
complete. Before any retry, classify the report by stage and use `--resume` only for an interrupted
matching direct-read checkpoint. A retry cannot turn an unassessed write, paid endpoint, or failed
ETL stream into a pass.

## Expected safe evidence extraction

After a completed run, retain the raw report/proof only under `task_cert_root`, then produce a
content-free aggregate. The aggregate may retain the fixed stage names, result classes, bounded
counts, leak count, rate-event count, binary hash, and report/proof SHA-256 fingerprints. It must
not retain `argv_redacted`, command arguments, provider URLs, provider values, request/response
bodies, identifiers, credentials, or file paths.

```sh
task_proof="$(find "$task_cert_root/.polymetrics/certifications/external-proof/gong" -type f -name 'external-*.json' -print -quit)"
[ -n "$task_proof" ]
shasum -a 256 "$task_pm_bin" "$task_cert_root/gong-report.json" "$task_proof" | awk '{print $1}' >"$task_cert_root/fingerprints.txt"
jq '{
  passed,
  stage_status_counts: ([.stages[]?.status] | sort | group_by(.) | map({status: .[0], count: length})),
  leak_count: (.leaks | length),
  rate_event_count: (.rate_limit_events | length),
  direct_read: (.capabilities.direct_read | {result, stages_checked, resumed_stages}),
  binary: (.capabilities.binary.result // "absent"),
  flow: (.capabilities.flow.result // "absent"),
  schedule: (.capabilities.schedule | {result, backend, residue}),
  write_result_counts: ([.capabilities.write_actions[]?.result] | sort | group_by(.) | map({result: .[0], count: length}))
}' "$task_cert_root/gong-report.json" >"$task_cert_root/gong-safe-summary.json"
```

Expected observations are:

- `credentials_test` authenticates; `gong_ordinary_rest_users_extensive` is the sole live
  direct-read candidate and returns a bounded ordinary REST success classification.
- `flow_plan`, `flow_preview`, `flow_run`, and `flow_status` prove the flow cell; the preview has
  no side effect. `schedule_create`, `schedule_list`, `schedule_install`, `schedule_fire`,
  `schedule_inspect`, and `schedule_remove` prove schedule with zero residue.
- Binary download remains `not_applicable`; the three binary uploads and all 27 writes remain
  `unassessed` until a declaration-owned disposable pairing independently reads back and cleans up
  a run-created record.
- All paid agentic cells remain `uncertified`; no HTTP exchange references their paths.
- The external proof must contain only fingerprints and bounded HTTP classifications. It may claim
  `observed_operations`, never `full_parity`, while writes or any other applicable cell remain
  unassessed, failed, blocked, or uncertified.
