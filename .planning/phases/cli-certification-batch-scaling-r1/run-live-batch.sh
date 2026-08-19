#!/usr/bin/env bash
# Runs a fixed-size, direct-read-only measurement from a disposable #4214
# candidate source. The credential is intentionally supplied only as the name
# of an already-exported environment variable; this script never reads, prints,
# writes, or places a credential on argv.
set -u
set -o pipefail

if [ "$#" -ne 5 ]; then
  echo "usage: $0 <candidate-source> <project-root> <count> <batch-id> <safe-summary-path>" >&2
  exit 2
fi

candidate_source=$1
project_root=$2
count=$3
batch_id=$4
summary_path=$5

case "$count" in
  ''|*[!0-9]*) echo "count must be a positive integer" >&2; exit 2 ;;
esac
if [ "$count" -lt 1 ] || [ "$count" -gt 120 ]; then
  echo "count must be in 1..120" >&2
  exit 2
fi
if [ -z "${PM_CERTIFICATION_THROUGHPUT_TOKEN+x}" ] || [ -z "$PM_CERTIFICATION_THROUGHPUT_TOKEN" ]; then
  echo "PM_CERTIFICATION_THROUGHPUT_TOKEN is unset" >&2
  exit 2
fi

case "$project_root" in
  "$candidate_source"/.throughput-projects/*) ;;
  *) echo "project root must be inside candidate-source/.throughput-projects" >&2; exit 2 ;;
esac

certification_json="$candidate_source/internal/connectors/defs/github/certification.json"
sweep_json="$candidate_source/internal/connectors/defs/github/certification-sweep.json"
report_path="$project_root/report.json"
checkpoint_path="$project_root/.polymetrics/certifications/progress/github-direct-read.json"
pm_binary="$candidate_source/pm-throughput"

if [ ! -f "$certification_json" ] || [ ! -f "$sweep_json" ]; then
  echo "missing candidate source or generated sweep" >&2
  exit 2
fi

# Generated candidates are selected first; the 100-operation run adds only the
# minimum three named overrides needed because #4214 supplies 97 generated
# direct reads. This rewrites only the disposable source copy.
tmp_json="$certification_json.throughput"
jq --argjson count "$count" '
  .direct_read_candidates as $all |
  ($all | map(select(.generated == true))) as $generated |
  ($all | map(select(.generated != true))) as $manual |
  (if $count <= ($generated | length)
   then $generated[:$count]
   else $generated + $manual[:($count - ($generated | length))]
   end) as $selected |
  if ($selected | length) != $count then error("insufficient direct-read candidates")
  else .direct_read_candidates = $selected | del(.direct_read_generation)
  end
' "$certification_json" > "$tmp_json" && mv "$tmp_json" "$certification_json"

build_started_ns=$(/opt/homebrew/bin/gdate +%s%N)
if ! (cd "$candidate_source" && go build -o pm-throughput ./cmd/pm); then
  echo "build of the pinned candidate source failed" >&2
  exit 1
fi
build_completed_ns=$(/opt/homebrew/bin/gdate +%s%N)

manifest_summary=$(jq -c --argjson count "$count" '
  .direct_read_candidates as $candidates |
  if ($candidates | length) != $count then error("candidate count drift") else . end |
  if any($candidates[];
    (.command | type) != "string" or
    (.output_assertions | length) == 0 or
    any(.output_assertions[]; (.json_pointer | startswith("/response") | not)) or
    ((.generated == true) and ((.output_assertions | length) != 1 or .output_assertions[0].json_pointer != "/response" or .output_assertions[0].value_type != "object_or_array"))
  )
  then error("candidate is not an assertion-bearing direct read")
  else {
    candidate_count: ($candidates | length),
    generated_count: ($candidates | map(select(.generated == true)) | length),
    manual_override_count: ($candidates | map(select(.generated != true)) | length),
    commands: ($candidates | map(.command))
  }
  end
' "$certification_json")

rm -rf "$project_root"
mkdir -p "$project_root"
started_ns=$(/opt/homebrew/bin/gdate +%s%N)
"$pm_binary" init --root "$project_root" --json > "$project_root/init.json"
if "$pm_binary" --root "$project_root" connectors certify github --direct-read-only --resume \
  --config owner=Polymetrics-Cert \
  --config repo=pm-cert-3993-20260810-wz0fru \
  --config rate_limit_account=polymetrics-ai-certification \
  --from-env token=PM_CERTIFICATION_THROUGHPUT_TOKEN --json > "$report_path"; then
  command_exit=0
else
  command_exit=$?
fi
before_teardown_ns=$(/opt/homebrew/bin/gdate +%s%N)
checkpoint_completed=0
if [ -f "$checkpoint_path" ]; then
  checkpoint_completed=$(jq '.completed | length' "$checkpoint_path")
fi

jq \
  --arg batch_id "$batch_id" \
  --arg source_sha "7306b9ec3e079b51ac9c70a674605a3a27f6e09b" \
  --argjson command_exit "$command_exit" \
  --argjson build_started_ns "$build_started_ns" \
  --argjson build_completed_ns "$build_completed_ns" \
  --argjson started_ns "$started_ns" \
  --argjson before_teardown_ns "$before_teardown_ns" \
  --argjson checkpoint_completed "$checkpoint_completed" \
  --argjson manifest "$manifest_summary" \
  --slurpfile source "$certification_json" \
  --slurpfile sweep "$sweep_json" '
    ($source[0].direct_read_candidates | map({key: .stage_name, value: .command}) | from_entries) as $command_for_stage |
    ($sweep[0].product_defects | map(.command)) as $product_defect_commands |
    .report as $report |
    ([$report.stages[] | select((.name | startswith("generated_direct_read_")) or (.name | startswith("direct_read_sweep_"))) | . + {command: $command_for_stage[.name]}]) as $reads |
    def missing_fixture:
      (.error // "" | ascii_downcase) as $error |
      ($error | test("no analysis found|no seat found|not found|does not exist|not enabled|could not be found"));
    def product_defect:
      .command as $command |
      ($command != null and ($product_defect_commands | index($command)) != null);
    ($reads | map(select(.passed == true and (.resumed != true))) | length) as $produced_value_passes |
    ($reads | map(select(.passed != true and product_defect)) | length) as $product_defects |
    ($reads | map(select(.passed != true and (product_defect | not) and missing_fixture)) | length) as $missing_fixtures |
    ($reads | map(select(.passed != true and (product_defect | not) and (missing_fixture | not))) | length) as $provider_refusals |
    {
      schema_version: 1,
      kind: "github_direct_read_throughput_batch",
      batch_id: $batch_id,
      candidate_source_pr: 4214,
      candidate_source_sha: $source_sha,
      command_exit: $command_exit,
      binary_build_elapsed_ms: (($build_completed_ns - $build_started_ns) / 1000000),
      started_epoch_ns: $started_ns,
      before_teardown_epoch_ns: $before_teardown_ns,
      lifecycle_elapsed_ms: (($before_teardown_ns - $started_ns) / 1000000),
      manifest: $manifest,
      checkpoint_completed: $checkpoint_completed,
      report_passed: $report.passed,
      direct_stage_count: ($reads | length),
      produced_value_passes: $produced_value_passes,
      product_defects: $product_defects,
      missing_fixtures: $missing_fixtures,
      provider_refusals: $provider_refusals,
      resumed_stages: ($reads | map(select(.resumed == true)) | length),
      failures: ($reads | map(select(.passed != true) | {stage: .name, command, error: .error, exit_code: .cli.exit_code})),
      rate_limit_events: (($report.rate_limit_events // []) | map({type, stage, method, attempt, reset_at, reason}))
    }
  ' "$report_path" > "$summary_path"

teardown_started_ns=$(/opt/homebrew/bin/gdate +%s%N)
rm -rf "$project_root"
teardown_completed_ns=$(/opt/homebrew/bin/gdate +%s%N)
jq --argjson teardown_started_ns "$teardown_started_ns" --argjson teardown_completed_ns "$teardown_completed_ns" '
  . + {
    teardown_elapsed_ms: (($teardown_completed_ns - $teardown_started_ns) / 1000000),
    teardown_verified: true,
    lifecycle_plus_teardown_elapsed_ms: (.lifecycle_elapsed_ms + (($teardown_completed_ns - $teardown_started_ns) / 1000000))
  }
' "$summary_path" > "$summary_path.tmp" && mv "$summary_path.tmp" "$summary_path"
