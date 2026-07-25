#!/usr/bin/env bash
# loop-trace — contained, private, size-capped observability for the autonomous loop.
#
# Automatic trace selection accepts only the exact path/id in the driver's private
# AUTO_LOOP_STATE_DIR/DRIVER-SESSION.json record. Complete evidence additionally requires exact Pi
# runtime/repository provenance and a clean base/head/tree/lineage binding. Session output uses a
# strict metadata allowlist; user/assistant/tool text and argument values are never persisted.
#
# Usage:
#   scripts/loop-trace.sh sessions            # list diagnostic session metadata
#   scripts/loop-trace.sh latest              # show exact driver-bound metadata
#   scripts/loop-trace.sh distill [path|id]   # exclusively write a private metadata pair
#   scripts/loop-trace.sh live                # follow exact driver-bound metadata (always unbound)
#   scripts/loop-trace.sh full [path|id]      # detailed allowlisted event metadata to stdout
#   scripts/loop-trace.sh html [path|id]      # exclusively write private allowlisted event metadata
#   scripts/loop-trace.sh turn <n>            # show digests recorded for driver turn n
set -euo pipefail
umask 077

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${AUTO_LOOP_STATE_DIR:-$REPO_ROOT/.planning/auto-loop}"
[[ "$STATE_DIR" = /* ]] || STATE_DIR="$REPO_ROOT/$STATE_DIR"
TRACE_DIR="$STATE_DIR/trace"
CMD="${1:-latest}"
ARG="${2:-}"

if (( $# > 2 )); then
  echo "usage: loop-trace.sh sessions|latest|distill [file]|full [file]|html [file]|live|turn <n>" >&2
  exit 2
fi

python3 - "$CMD" "$ARG" "$REPO_ROOT" "$STATE_DIR" "$TRACE_DIR" <<'PY'
import datetime as dt
import hashlib
import html as html_module
import json
import os
import re
import stat
import subprocess
import sys
import time

CMD, ARG, REPO_ROOT, STATE_DIR, TRACE_DIR = sys.argv[1:]
REPO_ROOT = os.path.realpath(REPO_ROOT)
STATE_DIR = os.path.abspath(STATE_DIR)
TRACE_DIR = os.path.abspath(TRACE_DIR)
USAGE = "usage: loop-trace.sh sessions|latest|distill [file]|full [file]|html [file]|live|turn <n>"
MAX_MD = 120
MAX_ITEMS = 40
MAX_SESSION_BYTES = 64 * 1024 * 1024
FUTURE_SKEW_SECONDS = 0
DRIVER_RECORD_SCHEMA = "polymetrics.ai/loop-trace-driver-session/v1"
EVENT_TYPES = ("session", "message", "model_change", "thinking_level_change", "compaction", "custom_message", "other")
MESSAGE_ROLES = ("user", "assistant", "toolResult", "system", "other")
CONTENT_TYPES = ("text", "thinking", "toolCall", "image", "other")
TOOL_NAMES = ("bash", "read", "edit", "write", "grep", "find", "ls", "subagent", "other")
os.umask(0o077)


class TraceError(Exception):
    def __init__(self, message, code="binding_failed"):
        super().__init__(message)
        self.code = safe_trace_slug(code, "diagnostic code")


def contains_control(value):
    return any(ord(char) < 32 or ord(char) == 127 for char in value)


def safe_trace_slug(value, label="trace slug"):
    """Return a bounded single path component or reject it without normalization."""
    if not isinstance(value, str) or not value or len(value) > 64:
        raise TraceError(f"invalid {label}")
    if contains_control(value) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]{0,63}", value):
        raise TraceError(f"unsafe {label}")
    return value


SENSITIVE_NAME = re.compile(
    r"(?i)(?:api[_-]?key|token|secret|password|passwd|private[_-]?key|access[_-]?key|"
    r"client[_-]?secret|authorization|credential|cookie)"
)
PRIVATE_KEY = re.compile(
    r"-----BEGIN(?: [A-Z0-9]+)? PRIVATE KEY-----.*?-----END(?: [A-Z0-9]+)? PRIVATE KEY-----",
    re.IGNORECASE | re.DOTALL,
)
ASSIGNMENT = re.compile(
    r"(?i)(\b(?:[A-Z0-9_]*(?:API_?KEY|TOKEN|SECRET|PASSWORD|PASSWD|PRIVATE_?KEY|"
    r"ACCESS_?KEY|CREDENTIAL)[A-Z0-9_]*|authorization|cookie)\b\s*(?:=|:)\s*)"
    r"(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)"
)
QUOTED_ASSIGNMENT = re.compile(
    r"(?i)((?:\"|')(?:[^\"']*(?:api[_-]?key|token|secret|password|passwd|private[_-]?key|"
    r"access[_-]?key|client[_-]?secret|authorization|credential|cookie)[^\"']*)(?:\"|')"
    r"\s*(?:=|:)\s*)(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)"
)
SECRET_FLAG = re.compile(
    r"(?i)(--(?:api[-_]?key|token|secret|password|passwd|private[-_]?key|access[-_]?key|"
    r"client[-_]?secret|authorization|credential)(?:\s+|=))(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)"
)
AUTH_SCHEME = re.compile(r"(?i)\b(bearer|basic)\s+[A-Za-z0-9._~+/=-]+")
URL_CREDENTIAL = re.compile(r"(?i)(https?://)[^/@\s:]+:[^/@\s]+@")
QUERY_SECRET = re.compile(
    r"(?i)([?&](?:api[_-]?key|access[_-]?token|token|secret|password|credential)=)[^&#\s]+"
)


def redact_text(value):
    value = PRIVATE_KEY.sub("[REDACTED PRIVATE KEY]", value)
    value = SECRET_FLAG.sub(r"\1[REDACTED]", value)
    value = QUOTED_ASSIGNMENT.sub(r"\1[REDACTED]", value)
    value = ASSIGNMENT.sub(r"\1[REDACTED]", value)
    value = AUTH_SCHEME.sub(lambda match: f"{match.group(1)} [REDACTED]", value)
    value = URL_CREDENTIAL.sub(r"\1[REDACTED]@", value)
    value = QUERY_SECRET.sub(r"\1[REDACTED]", value)
    return value


def redact_trace_value(value, key=""):
    """Recursively redact secret-like keys and values before any trace output or persistence."""
    if key and SENSITIVE_NAME.search(str(key)) and key not in {"session_id", "session_sha256"}:
        return "[REDACTED]"
    if isinstance(value, dict):
        return {str(k): redact_trace_value(v, str(k)) for k, v in value.items()}
    if isinstance(value, (list, tuple)):
        return [redact_trace_value(item) for item in value]
    if isinstance(value, str):
        return redact_text(value)
    if value is None or isinstance(value, (bool, int, float)):
        return value
    return redact_text(str(value))


def one_line(value, limit=240):
    value = str(redact_trace_value("" if value is None else value))
    value = " ".join(value.split())
    return value[:limit]


def require_sha(value, label):
    if not isinstance(value, str) or not re.fullmatch(r"[0-9a-f]{40}", value):
        raise TraceError(f"invalid {label}")
    return value


def read_regular_bytes(path, label, max_bytes=None):
    path = os.path.abspath(path)
    try:
        before = os.lstat(path)
    except OSError as exc:
        raise TraceError(f"{label} is unavailable") from exc
    if stat.S_ISLNK(before.st_mode) or not stat.S_ISREG(before.st_mode):
        raise TraceError(f"{label} must be a regular non-symlink file")
    if max_bytes is not None and before.st_size > max_bytes:
        raise TraceError(f"{label} exceeds the size limit", "session_too_large")
    flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
    try:
        fd = os.open(path, flags)
    except OSError as exc:
        raise TraceError(f"cannot open {label} safely") from exc
    try:
        after = os.fstat(fd)
        if not stat.S_ISREG(after.st_mode) or (before.st_dev, before.st_ino) != (after.st_dev, after.st_ino):
            raise TraceError(f"{label} changed during validation")
        if max_bytes is not None and after.st_size > max_bytes:
            raise TraceError(f"{label} exceeds the size limit", "session_too_large")
        chunks = []
        total = 0
        while True:
            chunk = os.read(fd, 1024 * 1024)
            if not chunk:
                break
            total += len(chunk)
            if max_bytes is not None and total > max_bytes:
                raise TraceError(f"{label} exceeds the size limit", "session_too_large")
            chunks.append(chunk)
        return b"".join(chunks), path, after
    finally:
        os.close(fd)


def read_json_file(path, label):
    data, _, _ = read_regular_bytes(path, label)
    try:
        value = json.loads(data.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise TraceError(f"invalid {label}") from exc
    if not isinstance(value, dict):
        raise TraceError(f"invalid {label}")
    return value


def load_session(path):
    data, real_path, source_stat = read_regular_bytes(path, "session source", MAX_SESSION_BYTES)
    if source_stat.st_uid != os.geteuid() or stat.S_IMODE(source_stat.st_mode) & 0o022:
        raise TraceError("session source permissions are unsafe", "unsafe_session_source")
    now = time.time()
    if source_stat.st_mtime > now + FUTURE_SKEW_SECONDS:
        raise TraceError("session source has a future modification time", "future_timestamp")
    try:
        text = data.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise TraceError("session source is not UTF-8") from exc
    events = []
    previous_epoch = None
    for line_number, line in enumerate(text.splitlines(), 1):
        if not line.strip():
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError as exc:
            raise TraceError(f"session source contains invalid JSON at line {line_number}") from exc
        if not isinstance(event, dict):
            raise TraceError(f"session source contains a non-object event at line {line_number}")
        epoch = parse_timestamp(event.get("timestamp"), required=True)
        if epoch > now + FUTURE_SKEW_SECONDS:
            raise TraceError("session source contains a future timestamp", "future_timestamp")
        if previous_epoch is not None and epoch < previous_epoch:
            raise TraceError("session timestamps are not monotonic", "invalid_timestamp_order")
        previous_epoch = epoch
        events.append(event)
    if not events or events[0].get("type") != "session":
        raise TraceError("session source must begin with its session event")
    session_events = [event for event in events if event.get("type") == "session"]
    if len(session_events) != 1:
        raise TraceError("session source must contain exactly one session event")
    session_event = session_events[0]
    if session_event.get("version") != 3:
        raise TraceError("session is not exact Pi v3 runtime output", "runtime_provenance_mismatch")
    session_id = safe_trace_slug(session_event.get("id"), "session id")
    cwd = session_event.get("cwd", "")
    if not isinstance(cwd, str) or not cwd or contains_control(cwd):
        raise TraceError("invalid session cwd")
    filename = os.path.basename(real_path)
    filename_match = re.fullmatch(
        rf"([0-9]{{4}}-[0-9]{{2}}-[0-9]{{2}})T([0-9]{{2}})-([0-9]{{2}})-([0-9]{{2}})-([0-9]{{3}})Z_{re.escape(session_id)}\.jsonl",
        filename,
    )
    if filename_match is None:
        raise TraceError("session filename does not match its exact Pi timestamp/id", "session_id_mismatch")
    filename_timestamp = (
        f"{filename_match.group(1)}T{filename_match.group(2)}:{filename_match.group(3)}:"
        f"{filename_match.group(4)}.{filename_match.group(5)}Z"
    )
    filename_epoch = parse_timestamp(filename_timestamp, required=True)
    session_epoch = parse_timestamp(session_event["timestamp"], required=True)
    if filename_epoch > now + FUTURE_SKEW_SECONDS:
        raise TraceError("session filename contains a future timestamp", "future_timestamp")
    if abs(filename_epoch - session_epoch) > 1:
        raise TraceError("session filename and event timestamps disagree", "invalid_timestamp_order")
    return {
        "path": real_path,
        "events": events,
        "session_id": session_id,
        "session_sha256": hashlib.sha256(data).hexdigest(),
        "cwd": cwd,
        "size": source_stat.st_size,
        "mtime": source_stat.st_mtime,
        "inode": (source_stat.st_dev, source_stat.st_ino),
        "started": events[0]["timestamp"],
        "ended": events[-1]["timestamp"],
    }


def exact_global_session_root():
    # Pi's exact project-session slug is '-' + absolute path with '/' mapped to '-' + '--'.
    slug = "-" + REPO_ROOT.replace("/", "-") + "--"
    return os.path.join(os.path.expanduser("~/.pi/agent/sessions"), slug)


def session_paths(include_global=False):
    roots = [(os.path.join(STATE_DIR, "sessions"), False)]
    if include_global:
        roots.append((exact_global_session_root(), True))
    paths = []
    seen = set()
    for root, require_repo_cwd in roots:
        if not os.path.lexists(root):
            continue
        try:
            root_stat = os.lstat(root)
        except OSError as exc:
            raise TraceError("session root is unavailable") from exc
        if stat.S_ISLNK(root_stat.st_mode) or not stat.S_ISDIR(root_stat.st_mode):
            raise TraceError("session root must be a non-symlink directory")
        for current, subdirs, names in os.walk(root, followlinks=False):
            safe_subdirs = []
            for name in subdirs:
                candidate = os.path.join(current, name)
                try:
                    mode = os.lstat(candidate).st_mode
                except OSError:
                    continue
                if stat.S_ISDIR(mode) and not stat.S_ISLNK(mode) and name not in {".git", "node_modules", "vendor"}:
                    safe_subdirs.append(name)
            subdirs[:] = safe_subdirs
            for name in names:
                if not name.endswith(".jsonl"):
                    continue
                candidate = os.path.abspath(os.path.join(current, name))
                try:
                    mode = os.lstat(candidate).st_mode
                except OSError:
                    continue
                if stat.S_ISLNK(mode) or not stat.S_ISREG(mode):
                    continue
                identity = os.path.realpath(candidate)
                if identity in seen:
                    continue
                if require_repo_cwd:
                    try:
                        session = load_session(candidate)
                    except TraceError:
                        continue
                    if not session["cwd"] or os.path.realpath(session["cwd"]) != REPO_ROOT:
                        continue
                seen.add(identity)
                paths.append(candidate)
    paths.sort(key=lambda path: os.lstat(path).st_mtime_ns, reverse=True)
    return paths


def normalize_runtime_provenance(value):
    if not isinstance(value, dict) or set(value) - {"name", "role", "pid", "version"}:
        raise TraceError("driver runtime provenance is malformed", "runtime_provenance_missing")
    if value.get("name") != "pi" or value.get("role") not in {"orchestrator", "driver"}:
        raise TraceError("driver runtime provenance is not the Pi orchestrator", "runtime_provenance_mismatch")
    if "pid" in value and (not isinstance(value["pid"], int) or isinstance(value["pid"], bool) or value["pid"] <= 0):
        raise TraceError("driver runtime pid is malformed", "runtime_provenance_mismatch")
    if "version" in value:
        safe_trace_slug(value["version"], "runtime version")
    return {"name": "pi", "role": "orchestrator"}


def driver_session_record():
    try:
        state_stat = os.lstat(STATE_DIR)
    except OSError as exc:
        raise TraceError("driver state root is missing", "driver_record_missing") from exc
    if stat.S_ISLNK(state_stat.st_mode) or not stat.S_ISDIR(state_stat.st_mode) or state_stat.st_uid != os.geteuid():
        raise TraceError("driver state root is unsafe", "unsafe_state_root")
    path = os.path.join(STATE_DIR, "DRIVER-SESSION.json")
    if not os.path.lexists(path):
        raise TraceError("driver session record is missing", "driver_record_missing")
    data, _, source_stat = read_regular_bytes(path, "driver session record", 64 * 1024)
    if source_stat.st_uid != os.geteuid() or stat.S_IMODE(source_stat.st_mode) & 0o022:
        raise TraceError("driver session record permissions are unsafe", "driver_record_invalid")
    if source_stat.st_mtime > time.time() + FUTURE_SKEW_SECONDS:
        raise TraceError("driver session record has a future modification time", "future_timestamp")
    try:
        value = json.loads(data.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise TraceError("driver session record is malformed", "driver_record_invalid") from exc
    allowed = {
        "schema_version", "session_path", "session_id", "session_sha256", "repository_cwd",
        "runtime", "recorded_at", "exact_base_sha", "exact_head_sha", "exact_head_tree",
        "candidate_lineage", "turn",
    }
    if not isinstance(value, dict) or set(value) - allowed:
        raise TraceError("driver session record has unsupported fields", "driver_record_invalid")
    if value.get("schema_version") != DRIVER_RECORD_SCHEMA:
        raise TraceError("driver session record schema is not exact", "driver_record_invalid")
    session_id = safe_trace_slug(value.get("session_id"), "driver session id")
    raw_path = value.get("session_path")
    if not isinstance(raw_path, str) or not raw_path or contains_control(raw_path):
        raise TraceError("driver session path is malformed", "driver_record_invalid")
    session_root = os.path.abspath(os.path.join(STATE_DIR, "sessions"))
    session_path = os.path.abspath(raw_path if os.path.isabs(raw_path) else os.path.join(STATE_DIR, raw_path))
    resolved_root = os.path.realpath(session_root)
    resolved_path = os.path.realpath(session_path)
    try:
        inside_root = os.path.commonpath([resolved_root, resolved_path]) == resolved_root
    except ValueError:
        inside_root = False
    if not inside_root or resolved_path == resolved_root:
        raise TraceError("driver session path is outside the exact session root", "driver_path_mismatch")
    try:
        root_stat = os.lstat(session_root)
    except OSError as exc:
        raise TraceError("driver session root is missing", "driver_path_mismatch") from exc
    if stat.S_ISLNK(root_stat.st_mode) or not stat.S_ISDIR(root_stat.st_mode):
        raise TraceError("driver session root is unsafe", "driver_path_mismatch")
    relative = os.path.relpath(session_path, session_root)
    current = session_root
    for component in relative.split(os.sep)[:-1]:
        if component in {"", ".", ".."}:
            raise TraceError("driver session path is malformed", "driver_path_mismatch")
        current = os.path.join(current, component)
        try:
            component_stat = os.lstat(current)
        except OSError as exc:
            raise TraceError("driver session directory is missing", "driver_path_mismatch") from exc
        if stat.S_ISLNK(component_stat.st_mode) or not stat.S_ISDIR(component_stat.st_mode):
            raise TraceError("driver session path contains an unsafe directory", "driver_path_mismatch")
    if not os.path.basename(session_path).endswith(f"_{session_id}.jsonl"):
        raise TraceError("driver session path and id disagree", "session_id_mismatch")
    cwd = value.get("repository_cwd")
    if not isinstance(cwd, str) or contains_control(cwd) or os.path.realpath(cwd) != REPO_ROOT:
        raise TraceError("driver repository cwd provenance is missing", "repository_cwd_mismatch")
    runtime = normalize_runtime_provenance(value.get("runtime"))
    recorded_at = value.get("recorded_at")
    recorded_epoch = parse_timestamp(recorded_at, required=True)
    if recorded_epoch > time.time() + FUTURE_SKEW_SECONDS:
        raise TraceError("driver session record contains a future timestamp", "future_timestamp")
    session_sha256 = value.get("session_sha256")
    if session_sha256 is not None and not re.fullmatch(r"[0-9a-f]{64}", str(session_sha256)):
        raise TraceError("driver session hash is malformed", "driver_record_invalid")
    if "turn" in value and (not isinstance(value["turn"], int) or isinstance(value["turn"], bool) or value["turn"] <= 0):
        raise TraceError("driver turn is malformed", "driver_record_invalid")
    for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
        require_sha(value.get(field), f"driver {field}")
    normalize_lineage(value.get("candidate_lineage"), value["exact_base_sha"], value["exact_head_sha"])
    return {
        "path": session_path,
        "session_id": session_id,
        "session_sha256": session_sha256,
        "repository_cwd": REPO_ROOT,
        "runtime": runtime,
        "recorded_at": recorded_at,
        "recorded_epoch": recorded_epoch,
        "exact_base_sha": value["exact_base_sha"],
        "exact_head_sha": value["exact_head_sha"],
        "exact_head_tree": value["exact_head_tree"],
        "candidate_lineage": value["candidate_lineage"],
        "record_sha256": hashlib.sha256(data).hexdigest(),
    }


def selected_session_path(record):
    if not ARG:
        return record["path"]
    if contains_control(ARG):
        raise TraceError("explicit session selector contains control characters", "driver_path_mismatch")
    if ARG == record["session_id"]:
        return record["path"]
    selected = os.path.abspath(ARG)
    if selected != record["path"]:
        raise TraceError("explicit session is not the exact driver-recorded session", "driver_path_mismatch")
    return record["path"]


def validate_session_binding(session, record):
    if session["path"] != record["path"]:
        raise TraceError("loaded session path differs from the driver record", "driver_path_mismatch")
    if session["session_id"] != record["session_id"]:
        raise TraceError("loaded session id differs from the driver record", "session_id_mismatch")
    if record["session_sha256"] is not None and session["session_sha256"] != record["session_sha256"]:
        raise TraceError("loaded session bytes differ from the driver record", "session_hash_mismatch")
    if os.path.realpath(session["cwd"]) != REPO_ROOT or os.path.realpath(session["cwd"]) != record["repository_cwd"]:
        raise TraceError("session cwd differs from repository provenance", "repository_cwd_mismatch")
    if parse_timestamp(session["ended"], required=True) > record["recorded_epoch"] + FUTURE_SKEW_SECONDS:
        raise TraceError("session events postdate the driver record", "session_record_order_mismatch")


def git_output(args):
    try:
        return subprocess.check_output(
            ["git", "-C", REPO_ROOT, *args], stderr=subprocess.DEVNULL
        )
    except (OSError, subprocess.CalledProcessError) as exc:
        raise TraceError("cannot capture exact candidate identity") from exc


def active_subissue(run_state):
    for subissue in run_state.get("subissues") or []:
        if isinstance(subissue, dict) and subissue.get("stage") not in (None, "not_started", "complete"):
            return subissue
    return None


def identity_nodes(run_state, orchestration_state):
    nodes = []
    for document in (run_state, orchestration_state):
        if not isinstance(document, dict):
            continue
        nodes.append(document)
        for key in ("automated_review", "review_coverage"):
            if isinstance(document.get(key), dict):
                nodes.append(document[key])
    selected = active_subissue(run_state)
    if selected:
        nodes.append(selected)
        for key in ("automated_review", "review_coverage"):
            if isinstance(selected.get(key), dict):
                nodes.append(selected[key])
    return nodes, selected


def normalize_lineage(value, exact_base_sha, exact_head_sha):
    if isinstance(value, dict):
        lineage_id = value.get("id")
        lineage_base = value.get("exact_base_sha")
        replacement_heads = value.get("replacement_heads")
        if lineage_base is not None and lineage_base != exact_base_sha:
            raise TraceError("candidate lineage base disagrees with exact base", "candidate_lineage_mismatch")
        if replacement_heads is not None:
            if not isinstance(replacement_heads, list) or exact_head_sha not in replacement_heads:
                raise TraceError("candidate lineage does not contain exact head", "candidate_lineage_mismatch")
            for head in replacement_heads:
                require_sha(head, "candidate lineage head")
    else:
        lineage_id = value
    if (
        not isinstance(lineage_id, str)
        or not lineage_id
        or len(lineage_id) > 512
        or contains_control(lineage_id)
    ):
        raise TraceError("candidate lineage is missing or malformed", "candidate_lineage_missing")
    return lineage_id


def declared_identity(run_state, orchestration_state, record):
    nodes, selected = identity_nodes(run_state, orchestration_state)
    result = {}
    for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
        values = []
        for node in nodes:
            value = node.get(field)
            if value in (None, ""):
                continue
            values.append(require_sha(value, field))
        if not values:
            raise TraceError(f"state does not declare {field}", "candidate_identity_missing")
        if len(set(values)) > 1:
            raise TraceError(f"conflicting declared {field}", "candidate_identity_mismatch")
        if values[0] != record[field]:
            raise TraceError(f"driver and state disagree on {field}", "candidate_identity_mismatch")
        result[field] = values[0]
    lineage_values = []
    lineage_nodes = ([selected] if selected else []) + [run_state, orchestration_state]
    for node in lineage_nodes:
        if not isinstance(node, dict) or node.get("candidate_lineage") in (None, ""):
            continue
        lineage_values.append(normalize_lineage(node["candidate_lineage"], result["exact_base_sha"], result["exact_head_sha"]))
    if not lineage_values:
        raise TraceError("state does not declare candidate lineage", "candidate_lineage_missing")
    if len(set(lineage_values)) > 1:
        raise TraceError("state candidate lineage declarations conflict", "candidate_lineage_mismatch")
    record_lineage = normalize_lineage(
        record["candidate_lineage"], result["exact_base_sha"], result["exact_head_sha"]
    )
    if lineage_values[0] != record_lineage:
        raise TraceError("driver and state candidate lineage disagree", "candidate_lineage_mismatch")
    result["candidate_lineage_sha256"] = hashlib.sha256(record_lineage.encode("utf-8")).hexdigest()
    return result, selected


def load_optional_state(name):
    path = os.path.join(STATE_DIR, name)
    if not os.path.lexists(path):
        return {}
    return read_json_file(path, name)


def capture_candidate_identity(record):
    exact_head_sha = require_sha(git_output(["rev-parse", "--verify", "HEAD"]).decode().strip(), "Git head")
    exact_head_tree = require_sha(git_output(["rev-parse", "--verify", "HEAD^{tree}"]).decode().strip(), "Git tree")
    if git_output(["status", "--porcelain=v1", "-z", "--untracked-files=all"]):
        raise TraceError("candidate worktree is dirty", "candidate_dirty")
    run_state = load_optional_state("RUN.json")
    orchestration_state = load_optional_state("ORCHESTRATION-STATE.json")
    declared, selected = declared_identity(run_state, orchestration_state, record)
    if declared["exact_head_sha"] != exact_head_sha or record["exact_head_sha"] != exact_head_sha:
        raise TraceError("declared candidate head does not match exact Git head", "candidate_identity_mismatch")
    if declared["exact_head_tree"] != exact_head_tree or record["exact_head_tree"] != exact_head_tree:
        raise TraceError("declared candidate tree does not match exact Git tree", "candidate_identity_mismatch")
    exact_base_sha = declared["exact_base_sha"]
    git_output(["cat-file", "-e", f"{exact_base_sha}^{{commit}}"])
    git_output(["cat-file", "-e", f"{exact_head_sha}^{{commit}}"])
    merge_base = require_sha(git_output(["merge-base", exact_base_sha, exact_head_sha]).decode().strip(), "merge base")
    if merge_base != exact_base_sha:
        raise TraceError("exact base is not the candidate merge base", "candidate_base_mismatch")
    candidate = {
        "schema_version": "polymetrics.ai/loop-trace-candidate/v2",
        "exact_base_sha": exact_base_sha,
        "exact_head_sha": exact_head_sha,
        "exact_head_tree": exact_head_tree,
        "candidate_lineage_sha256": declared["candidate_lineage_sha256"],
        "binding_status": "clean",
    }
    assert_candidate_unchanged(candidate)
    return candidate, run_state, selected


def assert_candidate_unchanged(candidate):
    head = require_sha(git_output(["rev-parse", "--verify", "HEAD"]).decode().strip(), "Git head")
    tree = require_sha(git_output(["rev-parse", "--verify", "HEAD^{tree}"]).decode().strip(), "Git tree")
    dirty = bool(git_output(["status", "--porcelain=v1", "-z", "--untracked-files=all"]))
    if head != candidate["exact_head_sha"] or tree != candidate["exact_head_tree"] or dirty:
        raise TraceError("candidate identity changed during trace capture", "candidate_identity_changed")


def parse_timestamp(value, required=False):
    if not isinstance(value, str) or not value:
        if required:
            raise TraceError("session timestamp is missing", "invalid_timestamp")
        return None
    if not re.fullmatch(r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,9})?(?:Z|[+-][0-9]{2}:[0-9]{2})", value):
        raise TraceError("session contains an invalid timestamp", "invalid_timestamp")
    try:
        parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
        if parsed.tzinfo is None:
            raise ValueError
        return parsed.timestamp()
    except (OverflowError, ValueError) as exc:
        raise TraceError("session contains an invalid timestamp", "invalid_timestamp") from exc


def ts_short(value):
    return value[11:19] if isinstance(value, str) else ""


def metadata_bucket(value, allowed):
    return value if isinstance(value, str) and value in allowed[:-1] else "other"


def empty_counts(keys):
    return {key: 0 for key in keys}


def session_statistics(session):
    event_counts = empty_counts(EVENT_TYPES)
    role_counts = empty_counts(MESSAGE_ROLES)
    content_counts = empty_counts(CONTENT_TYPES)
    tool_counts = empty_counts(TOOL_NAMES)
    content_chars = empty_counts(MESSAGE_ROLES)
    for event in session["events"]:
        event_type = metadata_bucket(event.get("type"), EVENT_TYPES)
        event_counts[event_type] += 1
        if event_type != "message":
            continue
        message = event.get("message")
        if not isinstance(message, dict):
            raise TraceError("session contains an invalid message event", "invalid_session_schema")
        role = metadata_bucket(message.get("role"), MESSAGE_ROLES)
        role_counts[role] += 1
        content = message.get("content") or []
        if not isinstance(content, list):
            raise TraceError("session contains invalid message content", "invalid_session_schema")
        for item in content:
            if not isinstance(item, dict):
                content_counts["other"] += 1
                continue
            item_type = metadata_bucket(item.get("type"), CONTENT_TYPES)
            content_counts[item_type] += 1
            if item_type == "toolCall":
                tool_counts[metadata_bucket(item.get("name"), TOOL_NAMES)] += 1
            elif item_type == "text":
                value = item.get("text")
                content_chars[role] += len(value) if isinstance(value, str) else 0
            elif item_type == "thinking":
                value = item.get("thinking")
                content_chars[role] += len(value) if isinstance(value, str) else 0
    return {
        "event_type_counts": event_counts,
        "message_role_counts": role_counts,
        "content_type_counts": content_counts,
        "tool_call_counts": tool_counts,
        "content_chars_by_role": content_chars,
    }


def digest_session(session, candidate, record):
    started_epoch = parse_timestamp(session["started"], required=True)
    ended_epoch = parse_timestamp(session["ended"], required=True)
    duration = int(ended_epoch - started_epoch)
    stats = session_statistics(session)
    digest = {
        "schema_version": "polymetrics.ai/loop-trace/v2",
        "trace_status": "complete",
        "trajectory_evidence": True,
        "binding": "driver_exact",
        "repository_cwd_match": True,
        "runtime": record["runtime"],
        "driver_record_sha256": record["record_sha256"],
        "driver_recorded_at": record["recorded_at"],
        "session_id": session["session_id"],
        "session_sha256": session["session_sha256"],
        "started": session["started"],
        "ended": session["ended"],
        "duration_s": duration,
        "event_count": len(session["events"]),
        "session_bytes": session["size"],
        "candidate_identity": candidate,
        "metadata": stats,
    }
    markdown = [
        "# Bound loop session metadata",
        "trace_status=complete trajectory_evidence=true binding=driver_exact",
        f"session_id={session['session_id']} session_sha256={session['session_sha256']}",
        f"started={session['started']} ended={session['ended']} duration_s={duration}",
        f"event_count={len(session['events'])} session_bytes={session['size']}",
        "repository_cwd_match=true runtime=pi runtime_role=orchestrator",
        f"driver_record_sha256={record['record_sha256']}",
        f"exact_base_sha={candidate['exact_base_sha']}",
        f"exact_head_sha={candidate['exact_head_sha']}",
        f"exact_head_tree={candidate['exact_head_tree']}",
        f"candidate_lineage_sha256={candidate['candidate_lineage_sha256']}",
        "candidate_worktree=clean",
        "",
        "## Strict allowlisted counters",
    ]
    for group in (
        "event_type_counts", "message_role_counts", "content_type_counts",
        "tool_call_counts", "content_chars_by_role",
    ):
        values = " ".join(f"{key}={value}" for key, value in stats[group].items())
        markdown.append(f"- {group}: {values}")
    return digest, "\n".join(markdown), "run", "metadata"


def open_state_root(create):
    if TRACE_DIR != os.path.join(STATE_DIR, "trace"):
        raise TraceError("trace destination is outside the state root")
    if not os.path.lexists(STATE_DIR):
        if not create:
            raise TraceError("state root is unavailable")
        parent = os.path.dirname(STATE_DIR)
        if not os.path.isdir(parent):
            raise TraceError("state root parent is unavailable")
        try:
            os.mkdir(STATE_DIR, 0o700)
        except OSError as exc:
            raise TraceError("cannot create private state root") from exc
    try:
        before = os.lstat(STATE_DIR)
    except OSError as exc:
        raise TraceError("state root is unavailable") from exc
    if stat.S_ISLNK(before.st_mode) or not stat.S_ISDIR(before.st_mode):
        raise TraceError("state root must be a non-symlink directory")
    try:
        fd = os.open(STATE_DIR, os.O_RDONLY | getattr(os, "O_DIRECTORY", 0) | getattr(os, "O_NOFOLLOW", 0))
    except OSError as exc:
        raise TraceError("cannot open state root safely") from exc
    after = os.fstat(fd)
    if (before.st_dev, before.st_ino) != (after.st_dev, after.st_ino):
        os.close(fd)
        raise TraceError("state root changed during validation")
    if not stat.S_ISDIR(after.st_mode) or after.st_uid != os.geteuid():
        os.close(fd)
        raise TraceError("state root is not an owned directory", "unsafe_state_root")
    return fd


def open_or_create_private_dir(parent_fd, name):
    safe_trace_slug(name, "directory name")
    try:
        os.mkdir(name, 0o700, dir_fd=parent_fd)
    except FileExistsError:
        pass
    except OSError as exc:
        raise TraceError("cannot create private trace directory") from exc
    try:
        fd = os.open(
            name,
            os.O_RDONLY | getattr(os, "O_DIRECTORY", 0) | getattr(os, "O_NOFOLLOW", 0),
            dir_fd=parent_fd,
        )
    except OSError as exc:
        raise TraceError("trace destination is a symlink or unsafe directory") from exc
    destination_stat = os.fstat(fd)
    if not stat.S_ISDIR(destination_stat.st_mode) or destination_stat.st_uid != os.geteuid():
        os.close(fd)
        raise TraceError("trace destination is not an owned directory")
    try:
        os.fchmod(fd, 0o700)
    except OSError as exc:
        os.close(fd)
        raise TraceError("cannot make trace directory private") from exc
    if stat.S_IMODE(os.fstat(fd).st_mode) != 0o700:
        os.close(fd)
        raise TraceError("trace directory is not private")
    return fd


def open_index_for_append(trace_fd):
    flags = os.O_WRONLY | os.O_APPEND | os.O_CREAT | getattr(os, "O_NOFOLLOW", 0)
    try:
        fd = os.open("INDEX.md", flags, 0o600, dir_fd=trace_fd)
    except OSError as exc:
        raise TraceError("trace index is a symlink or unsafe destination") from exc
    index_stat = os.fstat(fd)
    if not stat.S_ISREG(index_stat.st_mode) or index_stat.st_uid != os.geteuid():
        os.close(fd)
        raise TraceError("trace index is not an owned regular file")
    try:
        os.fchmod(fd, 0o600)
    except OSError as exc:
        os.close(fd)
        raise TraceError("cannot make trace index private") from exc
    if stat.S_IMODE(os.fstat(fd).st_mode) != 0o600:
        os.close(fd)
        raise TraceError("trace index is not private")
    return fd


def exclusive_write_at(directory_fd, name, content):
    """Create one private evidence file exclusively; never replace existing evidence."""
    if "/" in name or "\\" in name or contains_control(name) or name in {".", ".."}:
        raise TraceError("unsafe trace filename")
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | getattr(os, "O_NOFOLLOW", 0)
    try:
        fd = os.open(name, flags, 0o600, dir_fd=directory_fd)
    except FileExistsError as exc:
        raise TraceError("trace evidence already exists; refusing overwrite") from exc
    except OSError as exc:
        raise TraceError("cannot create trace evidence exclusively") from exc
    try:
        encoded = content.encode("utf-8")
        offset = 0
        while offset < len(encoded):
            offset += os.write(fd, encoded[offset:])
        os.fsync(fd)
        if stat.S_IMODE(os.fstat(fd).st_mode) != 0o600:
            raise TraceError("new trace evidence is not private")
    except Exception:
        os.close(fd)
        try:
            os.unlink(name, dir_fd=directory_fd)
        except OSError:
            pass
        raise
    os.close(fd)


def remove_at(directory_fd, name):
    try:
        os.unlink(name, dir_fd=directory_fd)
    except OSError:
        pass


def append_index(index_fd, line):
    encoded = one_line(line, 2000).encode("utf-8") + b"\n"
    offset = 0
    while offset < len(encoded):
        offset += os.write(index_fd, encoded[offset:])
    os.fsync(index_fd)


def evidence_basename(session, candidate, action):
    action = safe_trace_slug(action, "trace action")
    return (
        f"turn-{session['session_id']}-{session['session_sha256'][:12]}-"
        f"{candidate['exact_head_sha'][:12]}-{action}"
    )


def persist_digest(session, candidate, digest, markdown, slice_name, action):
    state_fd = open_state_root(create=True)
    trace_fd = slice_fd = index_fd = None
    created = []
    try:
        trace_fd = open_or_create_private_dir(state_fd, "trace")
        slice_fd = open_or_create_private_dir(trace_fd, slice_name)
        index_fd = open_index_for_append(trace_fd)
        base = evidence_basename(session, candidate, action)
        md_name = base + ".md"
        json_name = base + ".json"
        assert_candidate_unchanged(candidate)
        exclusive_write_at(slice_fd, md_name, markdown + "\n")
        created.append(md_name)
        exclusive_write_at(slice_fd, json_name, json.dumps(digest, indent=1, sort_keys=True) + "\n")
        created.append(json_name)
        assert_candidate_unchanged(candidate)
        timestamp = dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        append_index(
            index_fd,
            f"- {timestamp} trace_status=complete binding=driver_exact "
            f"session_id={session['session_id']} session_sha256={session['session_sha256']} "
            f"candidate_head={candidate['exact_head_sha']} {slice_name}/{base} "
            f"event_count={digest['event_count']} duration_s={digest['duration_s']}",
        )
    except Exception:
        if slice_fd is not None:
            for name in created:
                remove_at(slice_fd, name)
        raise
    finally:
        for fd in (index_fd, slice_fd, trace_fd, state_fd):
            if fd is not None:
                try:
                    os.close(fd)
                except OSError:
                    pass
    return os.path.join(TRACE_DIR, slice_name, base)


def event_metadata_line(event):
    event_type = metadata_bucket(event.get("type"), EVENT_TYPES)
    fields = [f"timestamp={event['timestamp']}", f"event_type={event_type}"]
    if event_type != "message":
        return " ".join(fields)
    message = event.get("message")
    if not isinstance(message, dict):
        raise TraceError("session contains an invalid message event", "invalid_session_schema")
    role = metadata_bucket(message.get("role"), MESSAGE_ROLES)
    fields.append(f"role={role}")
    content = message.get("content") or []
    if not isinstance(content, list):
        raise TraceError("session contains invalid message content", "invalid_session_schema")
    counts = empty_counts(CONTENT_TYPES)
    tools = empty_counts(TOOL_NAMES)
    chars = 0
    for item in content:
        if not isinstance(item, dict):
            counts["other"] += 1
            continue
        item_type = metadata_bucket(item.get("type"), CONTENT_TYPES)
        counts[item_type] += 1
        if item_type == "toolCall":
            tools[metadata_bucket(item.get("name"), TOOL_NAMES)] += 1
        elif item_type == "text":
            value = item.get("text")
            chars += len(value) if isinstance(value, str) else 0
        elif item_type == "thinking":
            value = item.get("thinking")
            chars += len(value) if isinstance(value, str) else 0
    fields.append(f"content_chars={chars}")
    fields.extend(f"content_{key}={value}" for key, value in counts.items() if value)
    fields.extend(f"tool_{key}={value}" for key, value in tools.items() if value)
    return " ".join(fields)


def full_markdown(session, candidate, record):
    _, summary, _, _ = digest_session(session, candidate, record)
    output = [summary, "", "## Ordered event metadata"]
    output.extend(f"- {event_metadata_line(event)}" for event in session["events"])
    return "\n".join(output)


def persist_html(session, candidate, markdown):
    state_fd = open_state_root(create=True)
    trace_fd = full_fd = index_fd = None
    name = None
    try:
        trace_fd = open_or_create_private_dir(state_fd, "trace")
        full_fd = open_or_create_private_dir(trace_fd, "full")
        index_fd = open_index_for_append(trace_fd)
        base = evidence_basename(session, candidate, "html")
        name = base + ".html"
        escaped = html_module.escape(markdown)
        document = (
            "<!doctype html><html><head><meta charset=\"utf-8\"><title>Loop trace metadata</title>"
            "<style>body{font-family:ui-monospace,monospace;white-space:pre-wrap;max-width:100ch;margin:2rem auto}</style>"
            f"</head><body>{escaped}</body></html>\n"
        )
        assert_candidate_unchanged(candidate)
        exclusive_write_at(full_fd, name, document)
        assert_candidate_unchanged(candidate)
        timestamp = dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        append_index(
            index_fd,
            f"- {timestamp} session_id={session['session_id']} session_sha256={session['session_sha256']} "
            f"candidate_head={candidate['exact_head_sha']} full/{base} metadata_html=true",
        )
    except Exception:
        if name and full_fd is not None:
            remove_at(full_fd, name)
        raise
    finally:
        for fd in (index_fd, full_fd, trace_fd, state_fd):
            if fd is not None:
                try:
                    os.close(fd)
                except OSError:
                    pass
    return os.path.join(TRACE_DIR, "full", name)


def render_live_event(event):
    parse_timestamp(event.get("timestamp"), required=True)
    return event_metadata_line(event)


def run_live(record, session, candidate):
    print(
        f"trace_status=diagnostic_unbound trajectory_evidence=false reason_code=live_mutable "
        f"session_id={session['session_id']} binding=driver_exact",
        flush=True,
    )
    path = record["path"]
    offset = session["size"]
    inode = session["inode"]
    try:
        while True:
            assert_candidate_unchanged(candidate)
            source_stat = os.lstat(path)
            if (source_stat.st_dev, source_stat.st_ino) != inode:
                raise TraceError("live session source was replaced", "session_source_changed")
            if source_stat.st_mtime > time.time() + FUTURE_SKEW_SECONDS:
                raise TraceError("live session has a future modification time", "future_timestamp")
            if source_stat.st_size < offset:
                raise TraceError("live session source was truncated", "session_source_changed")
            if source_stat.st_size == offset:
                time.sleep(2)
                continue
            data, _, _ = read_regular_bytes(path, "live session source", MAX_SESSION_BYTES)
            new_data = data[offset:]
            if new_data and not new_data.endswith(b"\n"):
                time.sleep(1)
                continue
            now = time.time()
            for line in new_data.decode("utf-8").splitlines():
                try:
                    event = json.loads(line)
                except json.JSONDecodeError as exc:
                    raise TraceError("live session contains invalid JSON", "invalid_session_schema") from exc
                if not isinstance(event, dict):
                    raise TraceError("live session contains a non-object event", "invalid_session_schema")
                if parse_timestamp(event.get("timestamp"), required=True) > now + FUTURE_SKEW_SECONDS:
                    raise TraceError("live session contains a future timestamp", "future_timestamp")
                print(f"session_id={session['session_id']} {render_live_event(event)}", flush=True)
            offset = source_stat.st_size
    except KeyboardInterrupt:
        return


def run_sessions():
    try:
        record = driver_session_record()
    except TraceError:
        record = None
    rejected = 0
    accepted = 0
    for path in session_paths(include_global=True)[:24]:
        try:
            session = load_session(path)
        except TraceError:
            rejected += 1
            continue
        accepted += 1
        age = max(0, int(time.time() - session["mtime"]))
        state = "active" if age < 300 else ("recent" if age < 3600 else "ended")
        driver_bound = bool(
            record
            and session["path"] == record["path"]
            and session["session_id"] == record["session_id"]
            and (record["session_sha256"] is None or session["session_sha256"] == record["session_sha256"])
        )
        print(
            f"state={state} session_id={session['session_id']} "
            f"session_sha256={session['session_sha256']} event_count={len(session['events'])} "
            f"session_bytes={session['size']} repository_cwd_match="
            f"{'true' if os.path.realpath(session['cwd']) == REPO_ROOT else 'false'} "
            f"driver_bound={'true' if driver_bound else 'false'}"
        )
    print(f"session_diagnostic accepted={accepted} rejected={rejected}")


def run_turn():
    if not re.fullmatch(r"[1-9][0-9]*", ARG):
        raise TraceError("turn requires a positive integer", "invalid_turn")
    evidence_files = 0
    if os.path.lexists(TRACE_DIR):
        mode = os.lstat(TRACE_DIR).st_mode
        if stat.S_ISLNK(mode) or not stat.S_ISDIR(mode):
            raise TraceError("trace root is a symlink or unsafe directory", "unsafe_trace_root")
        for current, subdirs, names in os.walk(TRACE_DIR, followlinks=False):
            subdirs[:] = [
                name for name in subdirs
                if not stat.S_ISLNK(os.lstat(os.path.join(current, name)).st_mode)
            ]
            for name in names:
                if re.fullmatch(r"turn-[A-Za-z0-9_-]+-[0-9a-f]{12}-[0-9a-f]{12}-(?:metadata|html)\.(?:md|json|html)", name):
                    evidence_files += 1
    driver_markers = 0
    driver_log = os.path.join(STATE_DIR, "driver.log")
    if os.path.lexists(driver_log):
        data, _, _ = read_regular_bytes(driver_log, "driver log", 16 * 1024 * 1024)
        needle = f"turn {ARG}:"
        driver_markers = sum(1 for line in data.decode("utf-8", errors="replace").splitlines() if needle in line)
    print(f"turn={ARG} evidence_files={evidence_files} driver_markers={driver_markers}")


def main():
    if contains_control(STATE_DIR) or contains_control(TRACE_DIR):
        raise TraceError("state destination contains control characters", "unsafe_state_root")
    if CMD == "sessions":
        if ARG:
            raise TraceError("sessions does not accept an argument", "invalid_arguments")
        run_sessions()
        return
    if CMD == "turn":
        run_turn()
        return
    if CMD not in {"latest", "distill", "full", "html", "live"}:
        print(USAGE, file=sys.stderr)
        raise SystemExit(2)
    if CMD == "live" and ARG:
        raise TraceError("live does not accept an argument", "invalid_arguments")
    record = driver_session_record()
    session = load_session(selected_session_path(record))
    validate_session_binding(session, record)
    candidate, _, _ = capture_candidate_identity(record)
    if CMD == "live":
        run_live(record, session, candidate)
        return
    if CMD in {"latest", "distill"}:
        digest, markdown, slice_name, action = digest_session(session, candidate, record)
        assert_candidate_unchanged(candidate)
        if CMD == "latest":
            print(markdown)
        else:
            persist_digest(session, candidate, digest, markdown, slice_name, action)
            print(
                f"wrote_private_metadata=true session_id={session['session_id']} "
                f"session_sha256={session['session_sha256']}"
            )
        return
    markdown = full_markdown(session, candidate, record)
    assert_candidate_unchanged(candidate)
    if CMD == "full":
        print(markdown)
    else:
        persist_html(session, candidate, markdown)
        print(
            f"wrote_private_metadata_html=true session_id={session['session_id']} "
            f"session_sha256={session['session_sha256']}"
        )


try:
    main()
except TraceError as exc:
    print(
        f"loop-trace: trace_status=diagnostic_unbound trajectory_evidence=false reason_code={exc.code}",
        file=sys.stderr,
    )
    raise SystemExit(1)
PY
