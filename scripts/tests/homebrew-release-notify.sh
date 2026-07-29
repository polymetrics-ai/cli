#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
HELPER="$ROOT_DIR/scripts/notify-homebrew-formula-update.sh"

fail() {
  printf 'homebrew-release-notify test failed: %s\n' "$1" >&2
  exit 1
}

expect_failure() {
  local expected="$1"
  shift
  local output status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e
  if [[ $status -eq 0 ]]; then
    printf '%s\n' "$output" >&2
    fail "expected command to fail with: $expected"
  fi
  if [[ "$output" != *"$expected"* ]]; then
    printf '%s\n' "$output" >&2
    fail "expected failure containing '$expected'"
  fi
}

run_static_assertions() {
  python3 - "$ROOT_DIR" <<'PY'
from pathlib import Path
import json
import sys

root = Path(sys.argv[1])
release = (root / ".github/workflows/release.yml").read_text()
website = (root / ".github/workflows/website.yml").read_text()
release_please = json.loads((root / "release-please-config.json").read_text())
helper = root / "scripts/notify-homebrew-formula-update.sh"

def require(condition, message):
    if not condition:
        raise SystemExit(message)

require(helper.exists(), "notification helper is missing")
require("permissions:\n  contents: read" in release, "release workflow must keep top-level contents: read")
for ignored in ['"website/**"', '".github/workflows/website.yml"', '".gitlab-ci.yml"']:
    require(ignored in release, f"release workflow must ignore website-only path {ignored}")
pm_package = release_please["packages"]["."]
require(pm_package["package-name"] == "pm", "release-please root package must remain PM")
expected_excludes = {"website", ".github/workflows/website.yml", ".gitlab-ci.yml"}
require(set(pm_package.get("exclude-paths", [])) >= expected_excludes, "release-please must exclude website-only paths from PM version calculation")
require("  notify-homebrew-tap:" in release, "release workflow must define notify-homebrew-tap job")
notify = release.split("  notify-homebrew-tap:", 1)[1]
release_assets = release.split("  release-assets:", 1)[1].split("  notify-homebrew-tap:", 1)[0]
release_please_job = release.split("  release-please:", 1)[1].split("  package-check:", 1)[0]
require("pm_cli_release_created" in release_please_job, "release-please must expose a PM CLI component release guard")
require("needs.release-please.outputs.pm_cli_release_created == 'true'" in release_assets, "release-assets must only follow PM CLI release-please releases")
require("release_component: ${{ steps.release-metadata.outputs.release_component }}" in release_assets, "release-assets must expose the PM release component")
require("echo \"release_component=pm-cli\"" in release_assets, "release metadata must classify canonical PM CLI releases")
require("source_run_id: ${{ steps.homebrew-source-run.outputs.source_run_id }}" in release_assets, "release-assets must not hard-code current run id for existing assets")
require("source_run_id: ${{ github.run_id }}" not in release_assets, "existing-asset notifications must not claim the current run signed old assets")
require("SKIP_RELEASE_ASSETS" in release_assets and "source_run_id=" in release_assets, "existing verified assets must omit mismatched source run id")
require("needs: release-assets" in notify, "notify-homebrew-tap must depend on release-assets")
require("homebrew_notification_ready" in notify, "notify-homebrew-tap must require verified release-assets output")
require("needs.release-assets.outputs.release_component == 'pm-cli'" in notify, "notify job must require PM CLI release component")
require("concurrency:" in notify and "homebrew-tap-notify-${{ needs.release-assets.outputs.tag_name }}" in notify, "notify job must serialize duplicate tag notifications")
require("cancel-in-progress: false" in notify, "duplicate notification concurrency must not cancel in-flight runs")
require("permissions:\n      contents: read" in notify, "notify job must keep GITHUB_TOKEN read-only")
require("ref: ${{ github.workflow_sha }}" in notify, "notify job must check out the trusted workflow-source commit before exposing App secrets")
require("PM_HOMEBREW_PR_APP_ID: ${{ secrets.PM_HOMEBREW_PR_APP_ID }}" in notify, "notify job must use approved App id secret name")
require("PM_HOMEBREW_PR_PRIVATE_KEY: ${{ secrets.PM_HOMEBREW_PR_PRIVATE_KEY }}" in notify, "notify job must use approved App private key secret name")
require("GH_TOKEN:" not in notify and "GITHUB_TOKEN:" not in notify, "notify job must not expose ordinary GitHub token")
for required in [
    "--release-assets-verified",
    "--source-repo polymetrics-ai/cli",
    "--tap-repo polymetrics-ai/homebrew-tap",
    "--workflow pm-formula-update.yml",
    "--dispatch-schema pm-homebrew-formula/v1",
    "--target-commitish-policy ignore",
    "--dry-run true",
    "--dispatch",
]:
    require(required in notify, f"notify job missing exact helper argument: {required}")
require("--dry-run false" not in notify, "notify job must not enable live Homebrew mutation")
require("workflow_run:" not in release, "release workflow must not be triggered by website workflow_run events")
for forbidden in ["homebrew", "pm-formula-update", "PM_HOMEBREW"]:
    require(forbidden.lower() not in website.lower(), f"website workflow must not reference {forbidden}")
require("release-assets:" in release and "homebrew_notification_ready" in release, "release-assets must expose a verified notification output")

def website_only(paths):
    return all(path.startswith("website/") or path in expected_excludes for path in paths)

scenarios = {
    "website only": (["website/app/page.tsx", ".github/workflows/website.yml"], False, False),
    "go only": (["cmd/pm/main.go", "internal/app/app.go"], True, True),
    "mixed go and website": (["cmd/pm/main.go", "website/app/page.tsx"], True, True),
    "website gitlab only": ([".gitlab-ci.yml"], False, False),
}
for name, (paths, release_workflow_runs, pm_release_candidate) in scenarios.items():
    ignored = website_only(paths)
    require((not ignored) == release_workflow_runs, f"{name} release workflow path classification is wrong")
    require((not ignored) == pm_release_candidate, f"{name} release-please path classification is wrong")
PY
}

start_fake_github_api() {
  local log_file="$1"
  local port_file="$2"
  python3 - "$log_file" "$port_file" <<'PY' >/dev/null 2>&1 &
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import json
import sys

log_path = Path(sys.argv[1])
port_path = Path(sys.argv[2])

class Handler(BaseHTTPRequestHandler):
    server_version = "FakeGitHub/1"

    def log_message(self, fmt, *args):
        return

    def read_json(self):
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length) if length else b""
        if not raw:
            return None
        return json.loads(raw.decode("utf-8"))

    def record(self, body=None):
        with log_path.open("a") as fh:
            fh.write(json.dumps({
                "method": self.command,
                "path": self.path,
                "authorization": self.headers.get("Authorization", ""),
                "body": body,
            }, sort_keys=True) + "\n")

    def send_json(self, status, payload):
        data = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def send_no_content(self):
        self.send_response(204)
        self.end_headers()

    def require_auth(self, expected=None):
        auth = self.headers.get("Authorization", "")
        if expected is None:
            ok = auth.startswith("Bearer ") and auth.count(".") == 2
        else:
            ok = auth == f"Bearer {expected}"
        if not ok:
            self.send_json(401, {"message": "bad auth"})
            return False
        return True

    def do_GET(self):
        if self.path == "/app":
            if not self.require_auth():
                return
            self.record()
            self.send_json(200, {"id": 12345, "slug": "pm-homebrew-pr-bot"})
            return
        if self.path == "/app/installations":
            if not self.require_auth():
                return
            self.record()
            self.send_json(200, [{"id": 98765, "account": {"login": "polymetrics-ai"}}])
            return
        if self.path == "/repos/polymetrics-ai/homebrew-tap/actions/workflows/pm-formula-update.yml":
            if not self.require_auth("app-token-for-homebrew-dispatch"):
                return
            self.record()
            self.send_json(200, {
                "id": 322895357,
                "name": "PM formula update",
                "path": ".github/workflows/pm-formula-update.yml",
                "state": "active",
            })
            return
        self.record()
        self.send_json(404, {"message": "not found"})

    def do_POST(self):
        body = self.read_json()
        if self.path == "/app/installations/98765/access_tokens":
            if not self.require_auth():
                return
            self.record(body)
            if body != {"repositories": ["homebrew-tap"], "permissions": {"actions": "write"}}:
                self.send_json(400, {"message": "unexpected token request", "body": body})
                return
            self.send_json(201, {
                "token": "app-token-for-homebrew-dispatch",
                "permissions": {"actions": "write", "metadata": "read"},
            })
            return
        if self.path == "/repos/polymetrics-ai/homebrew-tap/actions/workflows/pm-formula-update.yml/dispatches":
            if not self.require_auth("app-token-for-homebrew-dispatch"):
                return
            self.record(body)
            self.send_no_content()
            return
        self.record(body)
        self.send_json(404, {"message": "not found"})

server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
port_path.write_text(str(server.server_address[1]))
server.serve_forever()
PY
  echo $!
}

run_helper_tests() {
  [[ -x "$HELPER" ]] || fail "notification helper is not executable"
  command -v openssl >/dev/null || fail "openssl is required for helper tests"
  command -v curl >/dev/null || fail "curl is required for helper tests"
  command -v python3 >/dev/null || fail "python3 is required for helper tests"

  local tmp_dir key_file private_key log_file port_file server_pid api_base output
  tmp_dir="$(mktemp -d)"
  key_file="$tmp_dir/test-app-key.pem"
  log_file="$tmp_dir/fake-github.jsonl"
  port_file="$tmp_dir/fake-github.port"
  trap 'if [[ -n "${server_pid:-}" ]]; then kill "$server_pid" 2>/dev/null || true; fi; rm -rf "$tmp_dir"' RETURN

  openssl genrsa 2048 > "$key_file" 2>/dev/null
  private_key="$(cat "$key_file")"
  server_pid="$(start_fake_github_api "$log_file" "$port_file")"
  for _ in {1..50}; do
    [[ -s "$port_file" ]] && break
    sleep 0.1
  done
  [[ -s "$port_file" ]] || fail "fake GitHub API did not start"
  api_base="http://127.0.0.1:$(cat "$port_file")"

  local common_args=(
    --api-base "$api_base"
    --release-assets-verified true
    --tag v1.2.3
    --release-id 456789
    --source-run-id 987654
    --source-repo polymetrics-ai/cli
    --tap-repo polymetrics-ai/homebrew-tap
    --workflow pm-formula-update.yml
    --dispatch-schema pm-homebrew-formula/v1
    --target-commitish-policy ignore
    --dry-run true
  )
  local common_args_without_source_run=(
    --api-base "$api_base"
    --release-assets-verified true
    --tag v1.2.3
    --release-id 456789
    --source-repo polymetrics-ai/cli
    --tap-repo polymetrics-ai/homebrew-tap
    --workflow pm-formula-update.yml
    --dispatch-schema pm-homebrew-formula/v1
    --target-commitish-policy ignore
    --dry-run true
  )

  output="$(env -u GH_TOKEN -u GITHUB_TOKEN \
    PM_HOMEBREW_PR_APP_ID=12345 \
    PM_HOMEBREW_PR_PRIVATE_KEY="$private_key" \
    "$HELPER" "${common_args[@]}" --dispatch 2>&1)"
  [[ "$output" == *"requested Homebrew formula update dry-run for v1.2.3"* ]] || fail "authorized dispatch summary missing"
  [[ "$output" != *"app-token-for-homebrew-dispatch"* ]] || fail "helper printed the installation token"
  [[ "$output" != *"BEGIN RSA PRIVATE KEY"* && "$output" != *"BEGIN PRIVATE KEY"* ]] || fail "helper printed the private key"

  output="$(env -u GH_TOKEN -u GITHUB_TOKEN \
    PM_HOMEBREW_PR_APP_ID=12345 \
    PM_HOMEBREW_PR_PRIVATE_KEY="$private_key" \
    "$HELPER" "${common_args[@]}" --dispatch 2>&1)"
  [[ "$output" == *"requested Homebrew formula update dry-run for v1.2.3"* ]] || fail "duplicate dispatch summary missing"

  output="$(env -u GH_TOKEN -u GITHUB_TOKEN \
    PM_HOMEBREW_PR_APP_ID=12345 \
    PM_HOMEBREW_PR_PRIVATE_KEY="$private_key" \
    "$HELPER" "${common_args_without_source_run[@]}" --dispatch 2>&1)"
  [[ "$output" == *"requested Homebrew formula update dry-run for v1.2.3"* ]] || fail "existing-assets dispatch summary missing"

  python3 - "$log_file" <<'PY'
import json
import sys

events = [json.loads(line) for line in open(sys.argv[1]) if line.strip()]
dispatches = [event for event in events if event["path"].endswith("/dispatches")]
if len(dispatches) != 3:
    raise SystemExit(f"expected three dry-run dispatches, got {len(dispatches)}")
first = dispatches[0]["body"]
second = dispatches[1]["body"]
if first != second:
    raise SystemExit("duplicate dispatch payloads differed")
expected_inputs = {
    "dispatch_schema": "pm-homebrew-formula/v1",
    "source_repo": "polymetrics-ai/cli",
    "tag": "v1.2.3",
    "release_id": "456789",
    "source_run_id": "987654",
    "target_commitish_policy": "ignore",
    "dry_run": "true",
}
if first != {"ref": "main", "inputs": expected_inputs}:
    raise SystemExit(f"unexpected dispatch payload: {first!r}")
expected_existing_inputs = dict(expected_inputs)
del expected_existing_inputs["source_run_id"]
third = dispatches[2]["body"]
if third != {"ref": "main", "inputs": expected_existing_inputs}:
    raise SystemExit(f"unexpected existing-assets dispatch payload: {third!r}")
for event in dispatches:
    if event["authorization"] != "Bearer app-token-for-homebrew-dispatch":
        raise SystemExit("dispatch did not use explicit App token")
for event in events:
    if "ambient" in event["authorization"].lower():
        raise SystemExit("ambient credential was used")
PY

  expect_failure "PM_HOMEBREW_PR_APP_ID is required" \
    env -u GH_TOKEN -u GITHUB_TOKEN -u PM_HOMEBREW_PR_APP_ID -u PM_HOMEBREW_PR_PRIVATE_KEY \
    "$HELPER" "${common_args[@]}" --dispatch

  expect_failure "ambient GITHUB_TOKEN/GH_TOKEN is forbidden" \
    env -u GH_TOKEN -u PM_HOMEBREW_PR_APP_ID -u PM_HOMEBREW_PR_PRIVATE_KEY GITHUB_TOKEN=ambient-token \
    "$HELPER" "${common_args[@]}" --dispatch

  expect_failure "upstream PM release assets are not verified" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args[@]}" --release-assets-verified false --no-dispatch

  expect_failure "invalid PM release tag" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args[@]}" --tag 1.2.3 --no-dispatch

  expect_failure "unsupported Homebrew tap repository" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args[@]}" --tap-repo polymetrics-ai/other-tap --no-dispatch

  expect_failure "unsupported Homebrew workflow" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args[@]}" --workflow other.yml --no-dispatch

  expect_failure "live Homebrew formula mutation is not enabled" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args[@]}" --dry-run false --no-dispatch

  expect_failure "source_run_id must be a numeric Release workflow run id" \
    env -u GH_TOKEN -u GITHUB_TOKEN "$HELPER" "${common_args_without_source_run[@]}" --source-run-id current-run --no-dispatch
}

run_static_assertions
run_helper_tests
printf 'homebrew release notification assertions passed\n'
