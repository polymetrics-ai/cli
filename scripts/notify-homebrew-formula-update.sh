#!/usr/bin/env bash
set -euo pipefail

APP_SLUG="pm-homebrew-pr-bot"
APP_OWNER="polymetrics-ai"
APP_REPOSITORY="homebrew-tap"
APP_REPOSITORY_FULL="polymetrics-ai/homebrew-tap"
SOURCE_REPOSITORY_FULL="polymetrics-ai/cli"
WORKFLOW_FILE="pm-formula-update.yml"
WORKFLOW_PATH=".github/workflows/${WORKFLOW_FILE}"
DISPATCH_SCHEMA="pm-homebrew-formula/v1"
TARGET_COMMITISH_POLICY="ignore"
DISPATCH_REF="main"
USER_AGENT="polymetrics-cli-homebrew-release-notify"
API_BASE="${GITHUB_API_URL:-https://api.github.com}"

release_assets_verified=""
tag=""
release_id=""
source_run_id=""
source_repo="$SOURCE_REPOSITORY_FULL"
tap_repo="$APP_REPOSITORY_FULL"
workflow="$WORKFLOW_FILE"
dispatch_schema="$DISPATCH_SCHEMA"
target_commitish_policy="$TARGET_COMMITISH_POLICY"
dry_run="true"
dispatch="false"

usage() {
  cat <<'USAGE'
Usage: scripts/notify-homebrew-formula-update.sh --release-assets-verified true --tag vX.Y.Z --release-id N [--source-run-id N] --dispatch

Validates and sends the initial dry-run-only PM release notification to the
polymetrics-ai/homebrew-tap PM formula update workflow. The script never uses
GITHUB_TOKEN, GH_TOKEN, PATs, or ambient gh credentials; it mints and uses only
the approved pm-homebrew-pr-bot GitHub App installation token.
USAGE
}

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

require_value() {
  local flag="$1"
  local value="${2:-}"
  [[ -n "$value" ]] || fail "$flag requires a value"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --release-assets-verified)
      require_value "$1" "${2:-}"
      release_assets_verified="$2"
      shift 2
      ;;
    --tag)
      require_value "$1" "${2:-}"
      tag="$2"
      shift 2
      ;;
    --release-id)
      require_value "$1" "${2:-}"
      release_id="$2"
      shift 2
      ;;
    --source-run-id)
      require_value "$1" "${2:-}"
      source_run_id="$2"
      shift 2
      ;;
    --source-repo)
      require_value "$1" "${2:-}"
      source_repo="$2"
      shift 2
      ;;
    --tap-repo)
      require_value "$1" "${2:-}"
      tap_repo="$2"
      shift 2
      ;;
    --workflow)
      require_value "$1" "${2:-}"
      workflow="$2"
      shift 2
      ;;
    --dispatch-schema)
      require_value "$1" "${2:-}"
      dispatch_schema="$2"
      shift 2
      ;;
    --target-commitish-policy)
      require_value "$1" "${2:-}"
      target_commitish_policy="$2"
      shift 2
      ;;
    --dry-run)
      require_value "$1" "${2:-}"
      dry_run="$2"
      shift 2
      ;;
    --api-base)
      require_value "$1" "${2:-}"
      API_BASE="$2"
      shift 2
      ;;
    --dispatch)
      dispatch="true"
      shift
      ;;
    --no-dispatch)
      dispatch="false"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

API_BASE="${API_BASE%/}"

[[ -z "${GITHUB_TOKEN:-}" && -z "${GH_TOKEN:-}" ]] || \
  fail "ambient GITHUB_TOKEN/GH_TOKEN is forbidden; use the approved GitHub App secrets only"
[[ "$release_assets_verified" == "true" ]] || \
  fail "upstream PM release assets are not verified; refusing Homebrew notification"
[[ "$source_repo" == "$SOURCE_REPOSITORY_FULL" ]] || \
  fail "unsupported source repository: $source_repo"
[[ "$tap_repo" == "$APP_REPOSITORY_FULL" ]] || \
  fail "unsupported Homebrew tap repository: $tap_repo"
[[ "$workflow" == "$WORKFLOW_FILE" ]] || \
  fail "unsupported Homebrew workflow: $workflow"
[[ "$dispatch_schema" == "$DISPATCH_SCHEMA" ]] || \
  fail "unsupported Homebrew dispatch schema: $dispatch_schema"
[[ "$target_commitish_policy" == "$TARGET_COMMITISH_POLICY" ]] || \
  fail "unsupported target_commitish_policy: $target_commitish_policy"
[[ "$dry_run" == "true" ]] || \
  fail "live Homebrew formula mutation is not enabled; dry_run must remain true"
[[ "$tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || \
  fail "invalid PM release tag: $tag"
[[ "$release_id" =~ ^[0-9]+$ ]] || fail "release_id must be a numeric GitHub Release id"
if [[ -n "$source_run_id" && ! "$source_run_id" =~ ^[0-9]+$ ]]; then
  fail "source_run_id must be a numeric Release workflow run id"
fi

if [[ "$dispatch" != "true" ]]; then
  printf 'validated Homebrew formula update dry-run for %s (not dispatched)\n' "$tag"
  exit 0
fi

for tool in curl openssl python3; do
  command -v "$tool" >/dev/null || fail "$tool is required to notify the Homebrew tap"
done

[[ -n "${PM_HOMEBREW_PR_APP_ID:-}" ]] || fail "PM_HOMEBREW_PR_APP_ID is required"
[[ -n "${PM_HOMEBREW_PR_PRIVATE_KEY:-}" ]] || fail "PM_HOMEBREW_PR_PRIVATE_KEY is required"
[[ "${PM_HOMEBREW_PR_APP_ID}" =~ ^[0-9]+$ ]] || fail "PM_HOMEBREW_PR_APP_ID must be numeric"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

base64url() {
  openssl base64 -A | tr '+/' '-_' | tr -d '='
}

api_request() {
  local method="$1"
  local path="$2"
  local bearer="$3"
  local body_file="${4:-}"
  local expected_status="${5:-200}"
  local response_file status curl_status
  response_file="$(mktemp "$tmp_dir/api-response.XXXXXX")"

  local curl_args=(
    --silent
    --show-error
    --output "$response_file"
    --write-out '%{http_code}'
    --request "$method"
    --header 'Accept: application/vnd.github+json'
    --header 'X-GitHub-Api-Version: 2022-11-28'
    --header "User-Agent: ${USER_AGENT}"
    --header "Authorization: Bearer ${bearer}"
  )
  if [[ -n "$body_file" ]]; then
    curl_args+=(--header 'Content-Type: application/json' --data "@$body_file")
  fi

  set +e
  status="$(curl "${curl_args[@]}" "${API_BASE}${path}")"
  curl_status=$?
  set -e
  if [[ $curl_status -ne 0 ]]; then
    fail "GitHub API request failed before HTTP response: ${method} ${path}"
  fi

  if [[ "$status" != "$expected_status" ]]; then
    printf 'GitHub API request failed: %s %s returned HTTP %s (expected %s)\n' \
      "$method" "$path" "$status" "$expected_status" >&2
    python3 - "$response_file" <<'PY' >&2 || true
import json
import sys
from pathlib import Path
body = Path(sys.argv[1]).read_text(errors="replace")
try:
    message = json.loads(body).get("message")
except Exception:
    message = body.strip()
if message:
    print(f"GitHub API error: {message}")
PY
    exit 1
  fi

  printf '%s\n' "$response_file"
}

key_file="$tmp_dir/app-private-key.pem"
umask 077
printf '%s\n' "$PM_HOMEBREW_PR_PRIVATE_KEY" > "$key_file"
now="$(date +%s)"
header="$(printf '{"alg":"RS256","typ":"JWT"}' | base64url)"
payload="$(python3 - "$PM_HOMEBREW_PR_APP_ID" "$now" <<'PY' | base64url
import json
import sys
app_id = sys.argv[1]
now = int(sys.argv[2])
print(json.dumps({"iat": now - 60, "exp": now + 540, "iss": app_id}, separators=(",", ":")))
PY
)"
signing_input="${header}.${payload}"
signature="$(printf '%s' "$signing_input" | openssl dgst -sha256 -sign "$key_file" -binary | base64url)"
jwt="${signing_input}.${signature}"

app_response="$(api_request GET /app "$jwt" '' 200)"
python3 - "$app_response" "$PM_HOMEBREW_PR_APP_ID" "$APP_SLUG" <<'PY'
import json
import sys
app = json.load(open(sys.argv[1]))
expected_id = sys.argv[2]
expected_slug = sys.argv[3]
if str(app.get("id")) != expected_id or app.get("slug") != expected_slug:
    raise SystemExit("GitHub App identity is not the approved pm-homebrew-pr-bot")
PY

installations_response="$(api_request GET /app/installations "$jwt" '' 200)"
installation_id="$(python3 - "$installations_response" "$APP_OWNER" <<'PY'
import json
import sys
installations = json.load(open(sys.argv[1]))
owner = sys.argv[2]
for installation in installations:
    if installation.get("account", {}).get("login") == owner:
        print(installation["id"])
        break
else:
    raise SystemExit("pm-homebrew-pr-bot installation for polymetrics-ai was not found")
PY
)"

token_request="$tmp_dir/token-request.json"
python3 - "$token_request" "$APP_REPOSITORY" <<'PY'
import json
import sys
with open(sys.argv[1], "w") as fh:
    json.dump({"repositories": [sys.argv[2]], "permissions": {"actions": "write"}}, fh, separators=(",", ":"))
PY

token_response="$(api_request POST "/app/installations/${installation_id}/access_tokens" "$jwt" "$token_request" 201)"
app_token="$(python3 - "$token_response" <<'PY'
import json
import sys
response = json.load(open(sys.argv[1]))
permissions = response.get("permissions", {})
unexpected = set(permissions) - {"actions", "metadata"}
if permissions.get("actions") != "write" or permissions.get("metadata") not in (None, "read") or unexpected:
    raise SystemExit("pm-homebrew-pr-bot token permissions did not match the approved notification scope")
print(response["token"])
PY
)"

workflow_response="$(api_request GET "/repos/${APP_REPOSITORY_FULL}/actions/workflows/${WORKFLOW_FILE}" "$app_token" '' 200)"
python3 - "$workflow_response" "$WORKFLOW_PATH" <<'PY'
import json
import sys
workflow = json.load(open(sys.argv[1]))
expected_path = sys.argv[2]
if workflow.get("path") != expected_path or workflow.get("state") != "active":
    raise SystemExit("Homebrew workflow is not the approved active pm-formula-update.yml contract")
PY

dispatch_payload="$tmp_dir/workflow-dispatch.json"
python3 - "$dispatch_payload" "$tag" "$release_id" "$source_run_id" "$DISPATCH_REF" <<'PY'
import json
import sys
inputs = {
    "dispatch_schema": "pm-homebrew-formula/v1",
    "source_repo": "polymetrics-ai/cli",
    "tag": sys.argv[2],
    "release_id": sys.argv[3],
    "target_commitish_policy": "ignore",
    "dry_run": "true",
}
if sys.argv[4]:
    inputs["source_run_id"] = sys.argv[4]
payload = {
    "ref": sys.argv[5],
    "inputs": inputs,
}
with open(sys.argv[1], "w") as fh:
    json.dump(payload, fh, sort_keys=True, separators=(",", ":"))
PY

api_request POST "/repos/${APP_REPOSITORY_FULL}/actions/workflows/${WORKFLOW_FILE}/dispatches" "$app_token" "$dispatch_payload" 204 >/dev/null
printf 'requested Homebrew formula update dry-run for %s via %s/%s\n' \
  "$tag" "$APP_REPOSITORY_FULL" "$WORKFLOW_PATH"
