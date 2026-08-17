#!/usr/bin/env bash
# Replays only already-measured non-passing direct reads through their fixed
# declaration-owned PM argv so the provider's safe error message can classify
# missing fixtures separately from provider refusals. It never promotes a
# replay to a certification pass and never handles a credential value.
set -u
set -o pipefail

if [ "$#" -ne 5 ]; then
  echo "usage: $0 <pm-binary> <candidate-json> <safe-batch-summary> <project-root> <out-jsonl>" >&2
  exit 2
fi

pm_binary=$1
candidate_json=$2
batch_summary=$3
project_root=$4
out_jsonl=$5
result_path="$project_root/direct-read-classification-result.json"

if [ ! -x "$pm_binary" ] || [ ! -f "$candidate_json" ] || [ ! -f "$batch_summary" ]; then
  echo "missing classification input" >&2
  exit 2
fi

case "$project_root" in
  */.throughput-projects/*) ;;
  *) echo "classification project must be under a disposable throughput project root" >&2; exit 2 ;;
esac

: > "$out_jsonl"
while IFS= read -r command_name; do
  argv=()
  while IFS= read -r argument; do
    argv+=("$argument")
  done < <(jq -r --arg command "$command_name" '
    .direct_read_candidates[] |
    select(.command == $command) |
    .args[] |
    if .connector == true then "github"
    elif .source_credential == true then "throughput-classifier"
    elif has("literal") then .literal
    elif has("config_key") and has("default") then .default
    else error("candidate argument cannot be replayed safely")
    end
  ' "$candidate_json")
  if [ "${#argv[@]}" -eq 0 ]; then
    echo "candidate command missing from selected manifest: $command_name" >&2
    exit 1
  fi
  if "$pm_binary" --root "$project_root" "${argv[@]}" > "$result_path"; then
    jq -cn --arg command "$command_name" '{command:$command, replay_exit:0, replay_result:"produced_value_not_evaluated"}' >> "$out_jsonl"
  else
    replay_exit=$?
    jq --arg command "$command_name" --argjson replay_exit "$replay_exit" '
      {
        command:$command,
        replay_exit:$replay_exit,
        error_code:(.error.code // "unknown"),
        error_message:(.error.message // "unavailable")
      }
    ' "$result_path" >> "$out_jsonl"
  fi
done < <(jq -r '.failures[].command' "$batch_summary")
