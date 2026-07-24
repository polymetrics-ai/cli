#!/usr/bin/env python3
"""Run one human-like review experiment in a bounded disposable exact-head lab.

The canonical candidate is never writable. Declared lab-local reads, writes, caches, dummy data,
and local services are sandboxed and evidence-recorded. Sensitive capabilities require exact
captain approval; unsupported or unobservable enforcement fails closed.
"""

from __future__ import annotations

import argparse
import base64
import hashlib
import json
import math
import os
import platform
import re
import resource
import shutil
import signal
import socket
import stat
import subprocess
import sys
import tempfile
import time
from pathlib import Path, PurePosixPath
from typing import Any
from urllib.parse import urlparse

LAB_REQUEST_SCHEMA = "polymetrics.ai/pm-review-lab-request/v2"
LAB_EVIDENCE_SCHEMA = "polymetrics.ai/pm-review-lab-evidence/v3"
APPROVAL_PROPOSAL_SCHEMA = "polymetrics.ai/pm-review-lab-approval-proposal/v1"
APPROVAL_SCHEMA = "polymetrics.ai/pm-review-lab-approval/v1"
HEX_SHA = re.compile(r"^[0-9a-f]{40}$")
HEX_DIGEST = re.compile(r"^[0-9a-f]{64}$")
SAFE_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_.-]{0,79}$")
CONTROL = re.compile(r"[\x00-\x1f\x7f]")
EXPERIMENT_MODES = {
    "read_inspect",
    "disposable_write_test",
    "counterfactual_edit",
    "local_service_simulation",
}
APPROVAL_CAPABILITIES = {
    "external_network",
    "package_install",
    "host_credential",
    "live_connector",
    "live_connector_write",
}
LAB_DIRECTORY_NAMES = (
    "home",
    "tmp",
    "xdg-config",
    "xdg-cache",
    "xdg-data",
    "xdg-state",
    "go-cache",
    "go-mod-cache",
    "packages",
    "services",
    "credentials",
)
ROOT_ALIASES = {
    "candidate": "repo",
    "candidate_copy": "repo",
    "workspace": "repo",
    "repo": "repo",
    "home": "home",
    "tmp": "tmp",
    "temp": "tmp",
    "xdg_config": "xdg-config",
    "xdg-config": "xdg-config",
    "xdg_cache": "xdg-cache",
    "xdg-cache": "xdg-cache",
    "cache": "xdg-cache",
    "xdg_data": "xdg-data",
    "xdg-data": "xdg-data",
    "xdg_state": "xdg-state",
    "xdg-state": "xdg-state",
    "go_cache": "go-cache",
    "go-cache": "go-cache",
    "go_mod_cache": "go-mod-cache",
    "go-mod-cache": "go-mod-cache",
    "packages": "packages",
    "services": "services",
    "credentials": "credentials",
    "lab": "lab",
}
DEFAULT_LIMITS = {
    "timeout_seconds": 15 * 60.0,
    "max_processes": 128,
    "max_memory_bytes": 4 * 1024 * 1024 * 1024,
    "max_disk_bytes": 4 * 1024 * 1024 * 1024,
    "max_output_bytes": 16 * 1024 * 1024,
    "max_read_bytes": 512 * 1024 * 1024,
    "max_request_bytes": 1024 * 1024,
    "max_change_bytes": 64 * 1024 * 1024,
    "max_diff_bytes": 128 * 1024 * 1024,
}
MINIMUM_LIMITS = {
    "timeout_seconds": 0.05,
    "max_processes": 1,
    "max_memory_bytes": 64 * 1024 * 1024,
    "max_disk_bytes": 4096,
    "max_output_bytes": 1024,
    "max_read_bytes": 4096,
    "max_diff_bytes": 4096,
}
LEGACY_READ_ROOTS = [
    "repo", "home", "tmp", "xdg_config", "xdg_cache", "xdg_data", "xdg_state",
    "go_cache", "go_mod_cache", "packages", "services",
]
LEGACY_WRITE_ROOTS = [
    "repo", "home", "tmp", "xdg_config", "xdg_cache", "xdg_data", "xdg_state",
    "go_cache", "go_mod_cache", "packages", "services",
]


class LabError(ValueError):
    """Expected fail-closed lab input or safety error."""


def emit(value: dict[str, Any]) -> None:
    json.dump(value, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")


def run_git(
    root: Path,
    args: list[str],
    *,
    env: dict[str, str] | None = None,
    maximum: int = 512 * 1024 * 1024,
) -> str:
    command = ["git", "-C", str(root), "--no-optional-locks", *args]
    with tempfile.TemporaryFile() as stdout, tempfile.TemporaryFile() as stderr:
        proc = subprocess.run(command, check=False, stdout=stdout, stderr=stderr, env=env)
        output_size = stdout.tell()
        error_size = stderr.tell()
        if output_size > maximum or error_size > min(maximum, 1024 * 1024):
            raise LabError("read-only Git identity output exceeded its trusted read bound")
        stdout.seek(0)
        stderr.seek(0)
        output = stdout.read().decode(errors="replace")
        error = stderr.read().decode(errors="replace")
    if proc.returncode != 0:
        detail = error.strip() or output.strip()
        raise LabError(f"read-only Git identity command failed: {detail[:500]}")
    return output


def validate_sha(value: str, label: str) -> str:
    if not isinstance(value, str) or not HEX_SHA.fullmatch(value):
        raise LabError(f"{label} must be exactly 40 lowercase hexadecimal characters")
    return value


def validate_id(value: str, label: str) -> str:
    if not isinstance(value, str) or not SAFE_ID.fullmatch(value):
        raise LabError(f"{label} is malformed")
    return value


def validate_relative(value: str, label: str = "path") -> str:
    if not isinstance(value, str) or not value or CONTROL.search(value) or value.startswith("-") or "\\" in value:
        raise LabError(f"{label} must be a safe repository-relative POSIX path")
    path = PurePosixPath(value)
    if path.is_absolute() or ".." in path.parts or path.as_posix() in {"", "."}:
        raise LabError(f"{label} escapes the lab")
    return path.as_posix()


def candidate_identity(root: Path) -> dict[str, str]:
    head = run_git(root, ["rev-parse", "HEAD"]).strip()
    tree = run_git(root, ["rev-parse", "HEAD^{tree}"]).strip()
    status = run_git(root, ["status", "--porcelain=v1", "--untracked-files=all"])
    return {
        "head": head,
        "tree": tree,
        "status_sha256": hashlib.sha256(status.encode()).hexdigest(),
        "status": "clean" if not status else "dirty",
    }


def canonical_json_bytes(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode()


def is_within(path: Path, roots: list[Path] | set[Path] | tuple[Path, ...]) -> bool:
    resolved = path.resolve(strict=False)
    return any(resolved == root or root in resolved.parents for root in roots)


def safe_environment(lab_root: Path) -> dict[str, str]:
    python_bin = Path(sys.executable).resolve()
    path_parts = [
        str(python_bin.parent),
        "/usr/bin", "/bin", "/usr/sbin", "/sbin", "/opt/homebrew/bin", "/usr/local/bin",
    ]
    return {
        "PATH": ":".join(dict.fromkeys(path_parts)),
        "HOME": str(lab_root / "home"),
        "TMPDIR": str(lab_root / "tmp"),
        "TMP": str(lab_root / "tmp"),
        "TEMP": str(lab_root / "tmp"),
        "XDG_CONFIG_HOME": str(lab_root / "xdg-config"),
        "XDG_CACHE_HOME": str(lab_root / "xdg-cache"),
        "XDG_DATA_HOME": str(lab_root / "xdg-data"),
        "XDG_STATE_HOME": str(lab_root / "xdg-state"),
        "LANG": "C",
        "LC_ALL": "C",
        "TZ": "UTC",
        # CoreFoundation otherwise consults the real user's ~/.CFUserTextEncoding even with a dummy
        # HOME. Build this value from the process UID instead of inheriting host locale state.
        "__CF_USER_TEXT_ENCODING": f"0x{os.getuid():X}:0x0:0x0",
        "PYTHONNOUSERSITE": "1",
        "PYTHONDONTWRITEBYTECODE": "1",
        "PYTHONPYCACHEPREFIX": str(lab_root / "xdg-cache" / "python"),
        "PIP_CONFIG_FILE": os.devnull,
        "PIP_TARGET": str(lab_root / "packages"),
        "PIP_CACHE_DIR": str(lab_root / "xdg-cache" / "pip"),
        "NPM_CONFIG_USERCONFIG": str(lab_root / "xdg-config" / "npmrc"),
        "NPM_CONFIG_CACHE": str(lab_root / "xdg-cache" / "npm"),
        "NPM_CONFIG_PREFIX": str(lab_root / "packages"),
        "CARGO_HOME": str(lab_root / "xdg-data" / "cargo"),
        "GIT_CONFIG_GLOBAL": os.devnull,
        "GIT_CONFIG_SYSTEM": os.devnull,
        "GIT_CONFIG_COUNT": "2",
        "GIT_CONFIG_KEY_0": "user.name",
        "GIT_CONFIG_VALUE_0": "PM Review Lab",
        "GIT_CONFIG_KEY_1": "user.email",
        "GIT_CONFIG_VALUE_1": "pm-review-lab@example.invalid",
        "GIT_AUTHOR_NAME": "PM Review Lab",
        "GIT_AUTHOR_EMAIL": "pm-review-lab@example.invalid",
        "GIT_COMMITTER_NAME": "PM Review Lab",
        "GIT_COMMITTER_EMAIL": "pm-review-lab@example.invalid",
        "GIT_TERMINAL_PROMPT": "0",
        "GIT_ASKPASS": "/usr/bin/false",
        "GIT_SSH_COMMAND": "/usr/bin/false",
        "GIT_PAGER": "cat",
        "PAGER": "cat",
        "GOPROXY": "off",
        "GOSUMDB": "off",
        "GONOSUMDB": "*",
        "GOTOOLCHAIN": "local",
        "GOWORK": "off",
        "GOENV": "off",
        "GOCACHE": str(lab_root / "go-cache"),
        "GOMODCACHE": str(lab_root / "go-mod-cache"),
        "NO_PROXY": "localhost,127.0.0.1,::1",
        "no_proxy": "localhost,127.0.0.1,::1",
        "HTTP_PROXY": "",
        "HTTPS_PROXY": "",
        "ALL_PROXY": "",
        "http_proxy": "",
        "https_proxy": "",
        "all_proxy": "",
        "PM_REVIEW_LAB": "1",
        "PM_REVIEW_PACKAGES": str(lab_root / "packages"),
        "PM_REVIEW_SERVICES": str(lab_root / "services"),
    }


_SANDBOX_BACKEND_CACHE: dict[str, Any] | None = None


def probe_macos_sandbox() -> tuple[bool, str | None]:
    root = Path(tempfile.mkdtemp(prefix="pm-review-sandbox-probe-"))
    try:
        target = root / "denied"
        target.write_text("sentinel")
        profile = root / "probe.sb"
        escaped = str(target.resolve()).replace('"', '\\"')
        profile.write_text(
            "(version 1)\n"
            "(allow default)\n"
            f'(deny file-read* (literal "{escaped}") (with send-signal SIGKILL))\n'
        )
        # The attempted read catches ordinary exceptions in the command process. The sandbox must
        # terminate that process first; child-controlled exception text is never denial evidence.
        code = (
            "import os; "
            "exec(\"try:\\n open(os.environ['PM_REVIEW_DENIED']).read()\\nexcept Exception:\\n print('DENIAL_WAS_CAUGHT')\")"
        )
        proc = subprocess.run(
            ["/usr/bin/sandbox-exec", "-f", str(profile), str(Path(sys.executable).resolve()), "-c", code],
            check=False,
            capture_output=True,
            timeout=5,
            env={"PATH": "/usr/bin:/bin", "PM_REVIEW_DENIED": str(target.resolve())},
        )
        if proc.returncode != -signal.SIGKILL or b"DENIAL_WAS_CAUGHT" in proc.stdout:
            return False, "sandbox denial did not kill the whole descendant job with an uncatchable signal"
        return True, None
    except (OSError, subprocess.SubprocessError) as exc:
        return False, f"sandbox enforcement probe failed: {exc}"
    finally:
        shutil.rmtree(root, ignore_errors=True)


def sandbox_backend() -> dict[str, Any]:
    global _SANDBOX_BACKEND_CACHE
    if _SANDBOX_BACKEND_CACHE is not None:
        return dict(_SANDBOX_BACKEND_CACHE)
    if platform.system() == "Darwin" and Path("/usr/bin/sandbox-exec").is_file():
        proven, reason = probe_macos_sandbox()
        if proven:
            _SANDBOX_BACKEND_CACHE = {
                "status": "available",
                "name": "macos_sandbox_exec",
                "network": "default_deny_exact_exceptions",
                "outside_read": "default_deny",
                "outside_write": "default_deny",
                "denial_detection": "uncatchable_signal_kills_offending_sandbox_process",
                "memory_enforcement": "trusted_aggregate_rss_polling",
            }
            return dict(_SANDBOX_BACKEND_CACHE)
        _SANDBOX_BACKEND_CACHE = {"status": "blocked", "name": "macos_sandbox_exec", "reason": reason}
        return dict(_SANDBOX_BACKEND_CACHE)
    # Bubblewrap availability alone is not proof that namespaces and the intended policy work.
    _SANDBOX_BACKEND_CACHE = {
        "status": "blocked",
        "name": "unavailable",
        "reason": "no proven sandbox-exec or bubblewrap backend; policy-only execution cannot authorize clean evidence",
    }
    return dict(_SANDBOX_BACKEND_CACHE)


def sandbox_escape(value: Path | str) -> str:
    return str(value).replace('"', '\\"')


def executable_runtime_roots(commands: list[list[str]]) -> set[Path]:
    roots = {
        Path(sys.base_prefix).resolve(),
        Path("/System"), Path("/usr"), Path("/bin"), Path("/sbin"), Path("/dev"),
        Path("/Library/Developer/CommandLineTools"),
        Path("/opt/homebrew/Cellar"), Path("/opt/homebrew/opt"), Path("/opt/homebrew/bin"),
        Path("/usr/local/Cellar"), Path("/usr/local/opt"), Path("/usr/local/bin"),
        Path("/private/etc/ssl"), Path("/private/var/db/timezone/zoneinfo"),
        Path("/System/Library/Keychains/SystemRootCertificates.keychain"),
    }
    for command in commands:
        executable = Path(command[0]).resolve(strict=False)
        if executable.is_absolute() and not any(root == executable or root in executable.parents for root in roots):
            roots.add(executable.parent)
    existing = {root for root in roots if root.exists()}
    return existing | {root.resolve(strict=False) for root in existing}


def write_sandbox_profile(
    backend: dict[str, Any],
    lab_root: Path,
    repo: Path,
    read_roots: list[Path],
    write_roots: list[Path],
    commands: list[list[str]],
    local_endpoints: list[dict[str, Any]],
    external_endpoints: list[dict[str, Any]],
    allow_process_fork: bool,
) -> Path:
    if backend["name"] != "macos_sandbox_exec":
        raise LabError("sandbox backend is unavailable")
    profile = lab_root / "sandbox.sb"
    runtime_roots = executable_runtime_roots(commands)
    readable = sorted({path.resolve(strict=False) for path in read_roots} | runtime_roots, key=str)
    outside_filters = " ".join(
        f'(require-not (subpath "{sandbox_escape(value)}"))' for value in readable
    )
    rules = [
        "(version 1)",
        "(allow default)",
        "(deny network* (with send-signal SIGKILL))",
    ]
    if not allow_process_fork:
        rules.append("(deny process-fork (with send-signal SIGKILL))")
    for endpoint in local_endpoints:
        if endpoint["transport"] == "unix":
            socket_path = sandbox_escape(Path(endpoint["path"]).resolve(strict=False))
            rules.extend([
                f'(allow network-bind (local unix-socket (path-literal "{socket_path}")))',
                f'(allow network-inbound (local unix-socket (path-literal "{socket_path}")))',
                f'(allow network-outbound (remote unix-socket (path-literal "{socket_path}")))',
            ])
        else:
            port = int(endpoint["port"])
            rules.extend([
                f'(allow network-bind (local tcp "localhost:{port}"))',
                f'(allow network-inbound (local tcp "localhost:{port}"))',
                f'(allow network-outbound (remote tcp "localhost:{port}"))',
            ])
    for endpoint in external_endpoints:
        rules.append(f'(allow network-outbound (remote tcp "{endpoint["ip"]}:{endpoint["port"]}"))')
    if external_endpoints:
        # Hostname resolution is allowed only after exact external-network approval. The resulting
        # TCP connection is still restricted to the proposal's resolved IP/port set.
        rules.append('(allow network-outbound (literal "/private/var/run/mDNSResponder"))')
    metadata_ancestors = {Path("/"), Path("/private"), Path("/var"), Path("/tmp"), Path("/etc")}
    for readable_root in readable:
        current = readable_root
        while True:
            metadata_ancestors.add(current)
            if current == current.parent:
                break
            current = current.parent
    rules.append(f"(deny file-read* (require-all {outside_filters}) (with send-signal SIGKILL))")
    for ancestor in sorted(metadata_ancestors, key=str):
        rules.append(f'(allow file-read-metadata (literal "{sandbox_escape(ancestor)}"))')
    rules.extend([
        '(allow file-read-data (literal "/"))',
        "(deny file-write* (with send-signal SIGKILL))",
    ])
    for runtime_sink in ("/dev/null", "/dev/dtracehelper"):
        if Path(runtime_sink).exists():
            rules.append(f'(allow file-write-data (literal "{runtime_sink}"))')
    for root in sorted({path.resolve(strict=False) for path in write_roots}, key=str):
        rules.append(f'(allow file-write* (subpath "{sandbox_escape(root)}"))')
    runner_owned = [lab_root / "sandbox.sb", lab_root / "stdout.log", lab_root / "stderr.log"]
    for endpoint in local_endpoints:
        runner_owned.extend([
            lab_root / f"service-{endpoint['id']}.stdout.log",
            lab_root / f"service-{endpoint['id']}.stderr.log",
        ])
    for path in runner_owned:
        rules.append(f'(deny file-write* (literal "{sandbox_escape(path.resolve(strict=False))}") (with send-signal SIGKILL))')
    rules.append(f'(deny file-write* (subpath "{sandbox_escape((lab_root / "credentials").resolve())}") (with send-signal SIGKILL))')
    rules.append(f'(deny file-write* (subpath "{sandbox_escape((repo / ".git").resolve())}") (with send-signal SIGKILL))')
    profile.write_text("\n".join(rules) + "\n")
    return profile


def sandbox_command(backend: dict[str, Any], profile: Path, command: list[str]) -> list[str]:
    if backend["name"] == "macos_sandbox_exec":
        return ["/usr/bin/sandbox-exec", "-f", str(profile), *command]
    raise LabError("sandbox backend is unavailable")


def validate_command_vector(argv: Any, label: str = "command") -> list[str]:
    if not isinstance(argv, list) or not argv or not all(
        isinstance(item, str) and item and not CONTROL.search(item) for item in argv
    ):
        raise LabError(f"{label} must be a non-empty control-free argument vector")
    return list(argv)


def command_capabilities(argv: list[str]) -> set[str]:
    tool = Path(argv[0]).name.lower()
    args = [item.lower() for item in argv[1:]]
    result: set[str] = set()
    if tool in {"curl", "wget", "nc", "netcat", "ssh", "scp", "rsync"} or any(
        item.startswith(("http://", "https://", "ssh://", "git@")) for item in argv[1:]
    ):
        result.add("external_network")
    package_install = (
        tool in {"pip", "pip3", "npm", "npx", "yarn", "pnpm", "gem", "bundle"}
        and any(item in {"install", "add", "update"} for item in args)
    ) or (tool.startswith("python") and args[:2] in (["-m", "pip"], ["-m", "ensurepip"]))
    if package_install:
        result.update({"package_install", "external_network"})
    if tool == "pm":
        result.update({"live_connector", "external_network"})
        if any(item in {"write", "reverse", "delete", "create", "update", "execute"} for item in args):
            result.add("live_connector_write")
    return result


def unconditional_command_denial(argv: list[str]) -> str | None:
    tool = Path(argv[0]).name.lower()
    args = [item.lower() for item in argv[1:]]
    if tool in {"bash", "sh", "zsh", "fish", "dash", "ksh"}:
        return "generic shell execution is unconditionally denied"
    if tool in {"kubectl", "helm", "terraform", "ansible", "ansible-playbook", "podman", "docker", "gh", "gh-axi"}:
        return f"deployment/host-control command is unconditionally denied: {tool}"
    if tool == "git" and (not args or args[0] not in {"log", "show", "diff", "status", "blame", "grep", "rev-parse", "ls-files", "cat-file"}):
        return "Git mutation, commit, push, fetch, and remote operations are unconditionally denied"
    if any(item in {"publish", "deploy", "release"} for item in args) and tool in {
        "npm", "npx", "yarn", "pnpm", "gem", "cargo", "make",
    }:
        return "package publication or deployment is unconditionally denied"
    return None


def command_policy(argv: Any, repo: Path, lab_root: Path, readable_roots: list[Path]) -> list[str]:
    command = validate_command_vector(argv)
    denial = unconditional_command_denial(command)
    if denial:
        raise LabError(denial)
    tool = Path(command[0]).name
    args = command[1:]
    if tool.startswith("python"):
        if not args or args[0] in {"-", "-i"}:
            raise LabError("interactive Python is denied")
        executable = Path(sys.executable).resolve()
    elif tool == "git":
        executable_value = shutil.which("git")
        if not executable_value:
            raise LabError("Git is unavailable")
        executable = Path(executable_value).resolve()
        command = [str(executable), "--no-optional-locks", *args]
        args = command[2:]
    elif "/" in command[0]:
        candidate = Path(command[0])
        candidate = candidate if candidate.is_absolute() else repo / candidate
        try:
            executable = candidate.resolve(strict=True)
        except OSError as exc:
            raise LabError(f"command executable cannot be resolved: {exc}") from exc
        if not executable.is_file() or not os.access(executable, os.X_OK):
            raise LabError("command executable is not an executable file")
    else:
        executable_value = shutil.which(command[0], path=safe_environment(lab_root)["PATH"])
        if not executable_value:
            raise LabError(f"command tool is unavailable: {tool}")
        executable = Path(executable_value).resolve()
    trusted_runtime = executable_runtime_roots([[str(executable)]])
    allowed_arguments = list(readable_roots) + [lab_root.resolve(), *trusted_runtime]
    for arg in args:
        if arg.startswith(("http://", "https://", "ssh://", "git@")) or arg.startswith("-"):
            continue
        candidate_arg = PurePosixPath(arg)
        if candidate_arg.is_absolute():
            if not is_within(Path(arg), allowed_arguments):
                raise LabError("absolute command path argument is outside declared lab/runtime roots")
        elif ".." in candidate_arg.parts and not is_within((repo / arg).resolve(strict=False), [lab_root.resolve()]):
            raise LabError("command path argument escapes the lab")
    if tool == "go" and any(arg.startswith(("-exec", "-toolexec", "-overlay", "-modfile")) for arg in args):
        raise LabError("Go execution/overlay override is denied")
    if tool != "git":
        command = [str(executable), *args]
    return command


def merge_limits(requested: Any) -> dict[str, Any]:
    result = dict(DEFAULT_LIMITS)
    if requested is None:
        return result
    if not isinstance(requested, dict):
        raise LabError("limits must be an object")
    for key, value in requested.items():
        if key not in result or not isinstance(value, (int, float)):
            raise LabError(f"unknown or malformed limit: {key}")
        if value < MINIMUM_LIMITS.get(key, 1) or value > result[key]:
            raise LabError(f"requested {key} must only reduce the configured bound")
        result[key] = value
    return result


def normalize_root_spec(spec: Any, label: str) -> str:
    if isinstance(spec, str):
        raw = spec
    elif isinstance(spec, dict) and set(spec).issubset({"root", "path"}) and isinstance(spec.get("root"), str):
        raw = spec["root"]
        if spec.get("path") not in (None, ""):
            raw = f"{raw}/{validate_relative(spec['path'], label)}"
    else:
        raise LabError(f"{label} entries must be logical root strings or root/path objects")
    if not raw or CONTROL.search(raw) or raw.startswith(("/", "-")) or "\\" in raw:
        raise LabError(f"{label} entry is malformed")
    first, separator, remainder = raw.partition("/")
    canonical = ROOT_ALIASES.get(first)
    if canonical is None:
        raise LabError(f"unknown {label} logical root: {first}")
    if not separator:
        return canonical
    return f"{canonical}/{validate_relative(remainder, label)}"


def normalize_root_specs(value: Any, label: str, *, allow_empty: bool = False) -> list[str]:
    if not isinstance(value, list) or (not value and not allow_empty):
        suffix = "list" if allow_empty else "non-empty list"
        raise LabError(f"{label} must be a {suffix}")
    return sorted({normalize_root_spec(item, label) for item in value})


def resolve_root_spec(lab_root: Path, repo: Path, spec: str) -> Path:
    first, separator, remainder = spec.partition("/")
    base = lab_root if first == "lab" else repo if first == "repo" else lab_root / first
    candidate = base / remainder if separator else base
    resolved_base = base.resolve(strict=False)
    resolved = candidate.resolve(strict=False)
    if resolved != resolved_base and resolved_base not in resolved.parents:
        raise LabError("declared root escapes its logical lab root")
    current = candidate
    while current != base.parent and current != current.parent:
        if current.exists() and current.is_symlink():
            raise LabError("declared root traverses a symlink")
        if current == base:
            break
        current = current.parent
    return resolved


def resolve_root_specs(lab_root: Path, repo: Path, specs: list[str]) -> list[Path]:
    return sorted({resolve_root_spec(lab_root, repo, spec) for spec in specs}, key=str)


def write_roots_are_readable(read_roots: list[Path], write_roots: list[Path]) -> bool:
    return all(is_within(root, read_roots) for root in write_roots)


def request_commands(request: dict[str, Any]) -> list[list[str]]:
    result = [validate_command_vector(request["command"])]
    for service in request.get("services", []):
        result.append(validate_command_vector(service.get("command"), "service command"))
    return result


def declared_capability_details(request: dict[str, Any]) -> dict[str, dict[str, Any]]:
    raw = request.get("capability_requests", [])
    if not isinstance(raw, list):
        raise LabError("capability_requests must be a list")
    result: dict[str, dict[str, Any]] = {}
    for item in raw:
        if not isinstance(item, dict) or item.get("capability") not in APPROVAL_CAPABILITIES:
            raise LabError("capability request is malformed or unknown")
        capability = item["capability"]
        if capability in result:
            raise LabError(f"duplicate capability request: {capability}")
        result[capability] = item
    return result


def requested_capabilities(request: dict[str, Any]) -> tuple[list[str], dict[str, dict[str, Any]]]:
    details = declared_capability_details(request)
    inferred: set[str] = set(details)
    for command in request_commands(request):
        inferred.update(command_capabilities(command))
    return sorted(inferred), details


def inferred_network(command: list[str]) -> dict[str, Any] | None:
    destinations: list[dict[str, Any]] = []
    for arg in command:
        parsed = urlparse(arg) if arg.startswith(("http://", "https://")) else None
        if parsed and parsed.hostname:
            destinations.append({
                "host": parsed.hostname,
                "port": parsed.port or (443 if parsed.scheme == "https" else 80),
                "protocol": "tcp",
            })
    if not destinations:
        return None
    return {"destinations": destinations, "method": "command-declared URL"}


def approval_proposals(
    base: str,
    head: str,
    tree: str,
    packet_id: str,
    request: dict[str, Any],
    limits: dict[str, Any],
) -> list[dict[str, Any]]:
    capabilities, details_by_capability = requested_capabilities(request)
    commands = request_commands(request)
    proposals: list[dict[str, Any]] = []
    for capability in capabilities:
        details = details_by_capability.get(capability, {})
        network = details.get("network")
        if capability == "external_network" and network is None:
            network = inferred_network(request["command"])
        package = details.get("package")
        credential = details.get("credential")
        connector = details.get("connector")
        missing: list[str] = []
        why = details.get("dummy_execution_insufficient_reason", details.get("why_dummy_insufficient"))
        if not isinstance(why, str) or not why.strip():
            missing.append("dummy_execution_insufficient_reason")
            why = "not declared"
        if capability == "external_network":
            if not isinstance(network, dict) or not isinstance(network.get("method"), str) or not isinstance(network.get("destinations"), list) or not network["destinations"]:
                missing.append("network.destination_and_method")
        if capability == "package_install":
            if not isinstance(package, dict) or not all(isinstance(package.get(key), str) and package[key] for key in ("name", "version", "checksum")):
                missing.append("package.name_version_checksum")
        if capability == "host_credential":
            if not isinstance(credential, dict) or not all(isinstance(credential.get(key), str) and credential[key] for key in ("source_path", "mount_name", "scope")):
                missing.append("credential.source_mount_scope")
        if capability in {"live_connector", "live_connector_write"}:
            if not isinstance(connector, dict) or not all(isinstance(connector.get(key), str) and connector[key] for key in ("name", "operation", "scope", "access_mode")):
                missing.append("connector.name_operation_scope_access_mode")
            elif capability == "live_connector" and connector["access_mode"] != "read_only":
                missing.append("connector.read_only_default")
            elif capability == "live_connector_write" and connector["access_mode"] != "write":
                missing.append("connector.explicit_write_scope")
        safer = details.get("safer_dummy_alternative")
        if not isinstance(safer, str) or not safer.strip():
            missing.append("safer_dummy_alternative")
            safer = "not declared"
        minimum = details.get("minimum_privilege")
        if not isinstance(minimum, str) or not minimum.strip():
            missing.append("minimum_privilege")
            minimum = "not declared"
        cleanup = details.get("cleanup")
        if not isinstance(cleanup, str) or not cleanup.strip():
            missing.append("cleanup")
            cleanup = "destroy the whole disposable lab and terminate every tracked process group"
        evidence = details.get("evidence_to_capture", details.get("evidence_captured"))
        if not isinstance(evidence, list) or not evidence or not all(isinstance(item, str) and item for item in evidence):
            missing.append("evidence_to_capture")
            evidence = ["trusted command/effects/diff/services/policy/resource/candidate-cleanup record"]
        proposal = {
            "schema_version": APPROVAL_PROPOSAL_SCHEMA,
            "exact_candidate": {"base_sha": base, "head_sha": head, "head_tree": tree},
            "packet_id": packet_id,
            "experiment_id": request["hypothesis_id"],
            "capability": capability,
            "unresolved_hypothesis": request["claim"],
            "dummy_execution_insufficient_reason": why,
            "package": package,
            "network": network,
            "credential": credential,
            "connector": connector,
            "commands": commands,
            "files": {
                "read_roots": request["allowed_read_roots"],
                "write_roots": request["allowed_write_roots"],
            },
            "duration_seconds": limits["timeout_seconds"],
            "cleanup": cleanup,
            "expected_read_effects": details.get("expected_read_effects", request["allowed_read_roots"]),
            "expected_write_effects": details.get("expected_write_effects", request["allowed_write_roots"]),
            "minimum_privilege": minimum,
            "safer_dummy_alternative": safer,
            "evidence_to_capture": evidence,
            "proposal_ready": not missing,
            "missing_fields": sorted(missing),
        }
        proposal["proposal_sha256"] = hashlib.sha256(canonical_json_bytes(proposal)).hexdigest()
        proposals.append(proposal)
    return proposals


def approval_document(path_value: str, maximum: int, candidate: Path) -> dict[str, Any]:
    path = Path(path_value)
    if path.is_symlink():
        raise LabError("captain approval path must not be a symlink")
    try:
        resolved = path.resolve(strict=True)
        info = resolved.stat()
    except OSError as exc:
        raise LabError(f"cannot read captain approval: {exc}") from exc
    if not stat.S_ISREG(info.st_mode) or info.st_uid != os.getuid() or info.st_mode & 0o022:
        raise LabError("captain approval must be an owner-controlled, non-group/world-writable regular file")
    if is_within(resolved, [candidate.resolve()]):
        raise LabError("captain approval must be supplied outside the candidate")
    if info.st_size > maximum:
        raise LabError("captain approval byte bound exceeded")
    try:
        value = json.loads(read_file_limited(resolved, maximum, "captain approval").decode())
    except (OSError, UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise LabError(f"cannot parse captain approval: {exc}") from exc
    if not isinstance(value, dict):
        raise LabError("captain approval must be an object")
    return value


def validate_approvals(
    paths: list[str],
    proposals: list[dict[str, Any]],
    maximum: int,
    candidate: Path,
) -> list[dict[str, Any]]:
    if not proposals:
        if paths:
            raise LabError("approval object supplied for an experiment with no approval-gated capability")
        return []
    approvals = [approval_document(path, maximum, candidate) for path in paths]
    required_fields = {
        "schema_version", "decision", "exact_candidate", "packet_id", "experiment_id",
        "capability", "duration_seconds", "proposal_sha256",
    }
    by_capability: dict[str, dict[str, Any]] = {}
    for approval in approvals:
        if set(approval) != required_fields or approval.get("schema_version") != APPROVAL_SCHEMA or approval.get("decision") != "approved":
            raise LabError("captain approval object has an incompatible or non-approved exact shape")
        capability = approval.get("capability")
        if capability in by_capability:
            raise LabError("duplicate captain approval capability")
        by_capability[capability] = approval
    missing = [proposal for proposal in proposals if proposal["capability"] not in by_capability]
    if missing:
        return []
    if set(by_capability) != {proposal["capability"] for proposal in proposals}:
        raise LabError("captain approval grants capability outside the minimum requested scope")
    for proposal in proposals:
        if not proposal["proposal_ready"]:
            raise LabError(f"captain proposal is incomplete: {proposal['missing_fields']}")
        approval = by_capability[proposal["capability"]]
        expected = {
            "schema_version": APPROVAL_SCHEMA,
            "decision": "approved",
            "exact_candidate": proposal["exact_candidate"],
            "packet_id": proposal["packet_id"],
            "experiment_id": proposal["experiment_id"],
            "capability": proposal["capability"],
            "duration_seconds": proposal["duration_seconds"],
            "proposal_sha256": proposal["proposal_sha256"],
        }
        if approval != expected:
            raise LabError("captain approval does not exactly bind candidate/packet/experiment/capability/duration/proposal")
    return approvals


def external_network_endpoints(proposals: list[dict[str, Any]]) -> list[dict[str, Any]]:
    endpoints: list[dict[str, Any]] = []
    for proposal in proposals:
        if proposal["capability"] != "external_network":
            continue
        for destination in proposal["network"]["destinations"]:
            if not isinstance(destination, dict) or destination.get("protocol", "tcp") != "tcp":
                raise LabError("approved network destination must be an exact TCP host/port object")
            host = destination.get("host")
            port = destination.get("port")
            if not isinstance(host, str) or not host or not isinstance(port, int) or not (1 <= port <= 65535):
                raise LabError("approved network destination host/port is malformed")
            try:
                rows = socket.getaddrinfo(host, port, family=socket.AF_INET, type=socket.SOCK_STREAM)
            except OSError as exc:
                raise LabError(f"approved network destination cannot be resolved: {host}:{port}: {exc}") from exc
            for row in rows[:16]:
                endpoints.append({"host": host, "ip": row[4][0], "port": port, "protocol": "tcp"})
    unique = {(item["host"], item["ip"], item["port"]): item for item in endpoints}
    return [unique[key] for key in sorted(unique)]


def resolve_lab_file(repo: Path, relative: str, *, must_exist: bool = True) -> Path:
    normalized = validate_relative(relative, "temporary change path")
    if PurePosixPath(normalized).parts[0] == ".git":
        raise LabError("temporary changes to Git administration are denied")
    candidate = repo / normalized
    try:
        resolved = candidate.resolve(strict=must_exist)
    except OSError as exc:
        raise LabError(f"temporary change target cannot be resolved: {exc}") from exc
    if repo.resolve() not in resolved.parents or candidate.is_symlink():
        raise LabError("temporary change target escapes through an outside path or symlink")
    if must_exist and not resolved.is_file():
        raise LabError("temporary change target is not a regular file")
    parent = candidate.parent
    while parent != repo.parent:
        if parent.exists() and parent.is_symlink():
            raise LabError("temporary change parent escapes through a symlink")
        if parent == repo:
            break
        parent = parent.parent
    return resolved


def apply_changes(repo: Path, changes: Any, maximum: int, write_roots: list[Path]) -> list[str]:
    if not isinstance(changes, list):
        raise LabError("changes must be a list")
    changed: list[str] = []
    total = 0
    for item in changes:
        if not isinstance(item, dict):
            raise LabError("each temporary change must be an object")
        operation = item.get("operation", "replace")
        expected = {"path", "find", "replace"} if "operation" not in item else {
            "replace": {"operation", "path", "find", "replace"},
            "write": {"operation", "path", "content"},
            "delete": {"operation", "path"},
        }.get(operation)
        if expected is None or set(item) != expected:
            raise LabError("temporary change must be an exact replace/write/delete operation")
        path = resolve_lab_file(repo, item["path"], must_exist=operation != "write")
        if not is_within(path, write_roots):
            raise LabError("trusted temporary change is outside declared write roots")
        if operation == "replace":
            if path.stat().st_size > maximum:
                raise LabError("temporary change read bound exceeded")
            try:
                before = read_file_limited(path, maximum, "temporary change target").decode()
            except UnicodeDecodeError as exc:
                raise LabError("temporary change target must be UTF-8 text") from exc
            find_text = item["find"]
            replace_text = item["replace"]
            if not isinstance(find_text, str) or not isinstance(replace_text, str) or not find_text:
                raise LabError("temporary find/replace values are malformed")
            if before.count(find_text) != 1:
                raise LabError("temporary replacement must match exactly once")
            payload = before.replace(find_text, replace_text, 1).encode()
        elif operation == "write":
            if not isinstance(item["content"], str):
                raise LabError("temporary write content must be a string")
            payload = item["content"].encode()
            path.parent.mkdir(parents=True, exist_ok=True)
        else:
            payload = b""
        total += len(payload)
        if total > maximum:
            raise LabError("temporary change byte bound exceeded")
        if operation == "delete":
            path.unlink()
        else:
            path.write_bytes(payload)
        changed.append(path.relative_to(repo).as_posix())
    return sorted(set(changed))


def directory_size(root: Path) -> int:
    total = 0
    for base, directories, files in os.walk(root, followlinks=False):
        directories[:] = [name for name in directories if not (Path(base) / name).is_symlink()]
        for name in files:
            path = Path(base) / name
            try:
                if not path.is_symlink():
                    total += path.stat().st_size
            except OSError:
                continue
    return total


def hash_file_bounded(path: Path, remaining: list[int], label: str) -> tuple[str, int]:
    try:
        size = path.stat().st_size
    except OSError as exc:
        raise LabError(f"cannot stat {label}: {exc}") from exc
    if size > remaining[0]:
        raise LabError(f"{label} exceeds bounded trusted read envelope")
    digest = hashlib.sha256()
    read_bytes = 0
    try:
        with path.open("rb") as handle:
            while True:
                chunk = handle.read(min(1024 * 1024, remaining[0] - read_bytes + 1))
                if not chunk:
                    break
                read_bytes += len(chunk)
                if read_bytes > remaining[0]:
                    raise LabError(f"{label} exceeds bounded trusted read envelope")
                digest.update(chunk)
    except OSError as exc:
        raise LabError(f"cannot read {label}: {exc}") from exc
    remaining[0] -= read_bytes
    return digest.hexdigest(), read_bytes


def read_file_limited(path: Path, maximum: int, label: str) -> bytes:
    try:
        size = path.stat().st_size
    except OSError as exc:
        raise LabError(f"cannot stat {label}: {exc}") from exc
    if size > maximum:
        raise LabError(f"{label} exceeds bounded trusted read envelope")
    try:
        with path.open("rb") as handle:
            value = handle.read(maximum + 1)
    except OSError as exc:
        raise LabError(f"cannot read {label}: {exc}") from exc
    if len(value) > maximum:
        raise LabError(f"{label} exceeds bounded trusted read envelope")
    return value


def directory_hash(root: Path, maximum: int) -> str:
    digest = hashlib.sha256()
    remaining = [maximum]
    try:
        paths = sorted(root.rglob("*"))
    except OSError as exc:
        raise LabError(f"cannot enumerate bounded directory hash: {exc}") from exc
    for path in paths:
        relative = path.relative_to(root).as_posix().encode()
        if len(relative) > remaining[0]:
            raise LabError("directory hash path metadata exceeds bounded trusted read envelope")
        remaining[0] -= len(relative)
        digest.update(relative)
        if path.is_symlink():
            digest.update(b"L" + os.readlink(path).encode())
        elif path.is_file():
            file_digest, size = hash_file_bounded(path, remaining, "directory hash")
            digest.update(b"F" + str(size).encode() + file_digest.encode())
        elif path.is_dir():
            digest.update(b"D")
    return digest.hexdigest()


def bounded_capture(
    lab_root: Path,
    command: list[str],
    maximum: int,
    *,
    cwd: Path,
    env: dict[str, str],
    label: str,
) -> bytes:
    stdout_fd, stdout_value = tempfile.mkstemp(prefix=".runner-capture-", dir=lab_root)
    stderr_fd, stderr_value = tempfile.mkstemp(prefix=".runner-error-", dir=lab_root)
    os.close(stdout_fd)
    os.close(stderr_fd)
    stdout = Path(stdout_value)
    stderr = Path(stderr_value)
    try:
        with stdout.open("wb") as out, stderr.open("wb") as err:
            proc = subprocess.run(command, cwd=cwd, env=env, check=False, stdin=subprocess.DEVNULL, stdout=out, stderr=err)
        if stdout.stat().st_size > maximum or stderr.stat().st_size > min(maximum, 1024 * 1024):
            raise LabError(f"{label} output exceeds bounded trusted read envelope")
        output = stdout.read_bytes()
        error = stderr.read_bytes()
        if proc.returncode != 0:
            raise LabError(f"{label} failed: {error[:500].decode(errors='replace')}")
        return output
    finally:
        stdout.unlink(missing_ok=True)
        stderr.unlink(missing_ok=True)


def parse_porcelain_z(value: bytes) -> list[str]:
    fields = value.split(b"\0")
    result: list[str] = []
    index = 0
    while index < len(fields) and fields[index]:
        entry = fields[index]
        if len(entry) < 4:
            raise LabError("Git status emitted a malformed entry")
        status_value = entry[:2]
        path = entry[3:].decode(errors="surrogateescape")
        if CONTROL.search(path) or path.startswith(("/", "-")) or ".." in PurePosixPath(path).parts:
            raise LabError("experiment created an unsafe worktree path")
        result.append(path)
        index += 1
        if b"R" in status_value or b"C" in status_value:
            index += 1
    return sorted(set(result))


def git_effects(repo: Path, lab_root: Path, env: dict[str, str], limits: dict[str, Any]) -> dict[str, Any]:
    status_bytes = bounded_capture(
        lab_root,
        ["git", "-C", str(repo), "--no-optional-locks", "status", "--porcelain=v1", "-z", "--untracked-files=all"],
        int(limits["max_diff_bytes"]),
        cwd=repo,
        env=env,
        label="Git status",
    )
    paths = parse_porcelain_z(status_bytes)
    diff = bounded_capture(
        lab_root,
        ["git", "-C", str(repo), "--no-optional-locks", "diff", "--binary", "--no-ext-diff", "--no-renames", "HEAD", "--"],
        int(limits["max_diff_bytes"]),
        cwd=repo,
        env=env,
        label="Git diff",
    )
    summary = bounded_capture(
        lab_root,
        ["git", "-C", str(repo), "--no-optional-locks", "diff", "--stat", "--no-renames", "HEAD", "--"],
        min(int(limits["max_diff_bytes"]), 1024 * 1024),
        cwd=repo,
        env=env,
        label="Git diff summary",
    ).decode(errors="replace").strip()
    remaining = [int(limits["max_read_bytes"])]
    untracked: list[dict[str, Any]] = []
    for path_value in paths:
        path = repo / path_value
        status_prefix = next((entry[:2] for entry in status_bytes.split(b"\0") if len(entry) >= 4 and entry[3:].decode(errors="surrogateescape") == path_value), b"")
        if status_prefix != b"??":
            continue
        if path.is_symlink():
            target = os.readlink(path)
            untracked.append({"path": path_value, "kind": "symlink", "target": target, "escapes_lab": not is_within(path.resolve(strict=False), [lab_root.resolve()])})
        elif path.is_file():
            digest, size = hash_file_bounded(path, remaining, "untracked experiment output")
            untracked.append({"path": path_value, "kind": "file", "bytes": size, "sha256": digest})
        else:
            untracked.append({"path": path_value, "kind": "special"})
    encoded_diff = base64.b64encode(diff).decode()
    if len(encoded_diff.encode()) > int(limits["max_output_bytes"]):
        raise LabError("exact diff cannot fit the declared bounded evidence output envelope")
    aggregate = hashlib.sha256(diff + canonical_json_bytes(untracked)).hexdigest()
    return {
        "sha256": aggregate,
        "tracked_diff_sha256": hashlib.sha256(diff).hexdigest(),
        "tracked_diff_bytes": len(diff),
        "tracked_diff_base64": encoded_diff,
        "summary": summary,
        "changed_paths": paths,
        "untracked": untracked,
    }


def tree_snapshot(lab_root: Path, repo: Path, roots: list[Path], maximum: int) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    remaining = [maximum]
    ignored = {"sandbox.sb", "stdout.log", "stderr.log"}
    for root in sorted(set(roots), key=str):
        if root == repo or repo in root.parents:
            continue
        if not root.exists():
            continue
        for base, directories, files in os.walk(root, followlinks=False):
            base_path = Path(base)
            if base_path == lab_root:
                directories[:] = [name for name in directories if name != repo.name]
            directories[:] = [name for name in directories if not (base_path / name).is_symlink()]
            for name in files:
                path = base_path / name
                relative = path.relative_to(lab_root).as_posix()
                if relative in ignored or relative.startswith("service-") and relative.endswith((".stdout.log", ".stderr.log")):
                    continue
                if len(relative.encode()) > remaining[0]:
                    raise LabError("lab effect path metadata exceeds bounded trusted read envelope")
                remaining[0] -= len(relative.encode())
                try:
                    mode = path.lstat().st_mode
                except OSError as exc:
                    raise LabError(f"cannot inspect lab effect {relative}: {exc}") from exc
                if stat.S_ISLNK(mode):
                    target = os.readlink(path)
                    result[relative] = {
                        "kind": "symlink",
                        "target": target,
                        "escapes_lab": not is_within(path.resolve(strict=False), [lab_root.resolve()]),
                    }
                elif stat.S_ISREG(mode):
                    digest, size = hash_file_bounded(path, remaining, "lab effect")
                    result[relative] = {"kind": "file", "bytes": size, "sha256": digest}
                elif stat.S_ISSOCK(mode):
                    result[relative] = {"kind": "socket"}
                else:
                    result[relative] = {"kind": "special", "mode": stat.S_IFMT(mode)}
    return result


def compare_snapshots(before: dict[str, Any], after: dict[str, Any]) -> dict[str, Any]:
    created = sorted(set(after) - set(before))
    deleted = sorted(set(before) - set(after))
    modified = sorted(path for path in set(before) & set(after) if before[path] != after[path])
    changed_paths = sorted({*created, *modified, *deleted})
    rows = {path: after.get(path, before.get(path)) for path in changed_paths}
    disclosure = {
        "changed_paths": changed_paths,
        "created": created,
        "modified": modified,
        "deleted": deleted,
        "entries": rows,
    }
    return {**disclosure, "sha256": hashlib.sha256(canonical_json_bytes(disclosure)).hexdigest()}


def process_snapshot() -> dict[int, tuple[int, int, int]]:
    proc = subprocess.run(["/bin/ps", "-axo", "pid=,ppid=,pgid=,rss="], check=False, capture_output=True, text=True)
    result: dict[int, tuple[int, int, int]] = {}
    for line in proc.stdout.splitlines():
        fields = line.split()
        if len(fields) == 4 and all(field.isdigit() for field in fields):
            result[int(fields[0])] = (int(fields[1]), int(fields[2]), int(fields[3]) * 1024)
    return result


def lab_cwd_pids(lab_root: Path) -> set[int]:
    lsof = Path("/usr/sbin/lsof")
    if not lsof.is_file():
        raise LabError("trusted descendant cwd discovery is unavailable")
    proc = subprocess.run(
        [str(lsof), "-a", "-d", "cwd", "-Fpn"],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode not in {0, 1} or len(proc.stdout.encode()) > 16 * 1024 * 1024:
        raise LabError("trusted descendant cwd discovery failed or exceeded its output bound")
    result: set[int] = set()
    current: int | None = None
    for line in proc.stdout.splitlines():
        if line.startswith("p") and line[1:].isdigit():
            current = int(line[1:])
        elif line.startswith("n") and current is not None:
            path = Path(line[1:])
            if path.is_absolute() and is_within(path, [lab_root.resolve()]):
                result.add(current)
    return result


def descendant_pids(roots: set[int], snapshot: dict[int, tuple[int, int, int]]) -> set[int]:
    result = {pid for pid in roots if pid in snapshot}
    changed = True
    while changed:
        changed = False
        for pid, (parent, _, _) in snapshot.items():
            if parent in result and pid not in result:
                result.add(pid)
                changed = True
    return result


def signal_processes(pids: set[int], groups: set[int], requested_signal: int) -> None:
    for group in sorted(groups):
        try:
            os.killpg(group, requested_signal)
        except ProcessLookupError:
            pass
        except PermissionError:
            # Darwin reports EPERM for a group containing only an already-dead zombie. Direct PID
            # cleanup and the final trusted process snapshot below still fail closed on live residue.
            pass
    for pid in sorted(pids, reverse=True):
        try:
            os.kill(pid, requested_signal)
        except ProcessLookupError:
            pass
        except PermissionError:
            pass


def preexec(limits: dict[str, Any]) -> None:
    os.setsid()
    resource.setrlimit(resource.RLIMIT_CORE, (0, 0))
    resource.setrlimit(resource.RLIMIT_NOFILE, (256, 256))
    cpu = max(1, int(math.ceil(float(limits["timeout_seconds"]))) + 1)
    resource.setrlimit(resource.RLIMIT_CPU, (cpu, cpu))
    file_size = int(max(limits["max_output_bytes"], limits["max_disk_bytes"]))
    resource.setrlimit(resource.RLIMIT_FSIZE, (file_size, file_size))
    # Darwin exposes RLIMIT_AS but rejects setting it. Aggregate trusted RSS polling below is the
    # enforceable memory bound there; other Unix hosts also inherit a hard address-space limit.
    if platform.system() != "Darwin" and hasattr(resource, "RLIMIT_AS"):
        memory = int(limits["max_memory_bytes"])
        resource.setrlimit(resource.RLIMIT_AS, (memory, memory))


def service_ready(service: dict[str, Any]) -> bool:
    if service["transport"] == "unix":
        path = Path(service["path"])
        if service["readiness"] == "path_exists":
            try:
                return stat.S_ISSOCK(path.stat().st_mode)
            except FileNotFoundError:
                return False
        client = socket.socket(socket.AF_UNIX)
        client.settimeout(0.1)
        try:
            client.connect(str(path))
            return True
        except OSError:
            return False
        finally:
            client.close()
    try:
        with socket.create_connection((service["host"], service["port"]), timeout=0.1):
            return True
    except OSError:
        return False


def execute(
    lab_root: Path,
    repo: Path,
    argv: list[str],
    env: dict[str, str],
    limits: dict[str, Any],
    backend: dict[str, Any],
    profile: Path,
    services: list[dict[str, Any]],
) -> dict[str, Any]:
    baseline_size = directory_size(lab_root)
    started = time.perf_counter()
    hit: str | None = None
    tracked_pids: set[int] = set()
    tracked_groups: set[int] = set()
    peak_processes = 0
    peak_memory_bytes = 0
    peak_disk_bytes = 0
    peak_output_bytes = 0
    last_disk_sample = 0.0
    jobs: list[dict[str, Any]] = []
    log_paths: list[Path] = []
    handles: list[Any] = []
    unexpected_service_codes: list[dict[str, Any]] = []
    service_failure: str | None = None

    def start_job(name: str, command: list[str]) -> dict[str, Any]:
        stdout_path = lab_root / ("stdout.log" if name == "command" else f"service-{name}.stdout.log")
        stderr_path = lab_root / ("stderr.log" if name == "command" else f"service-{name}.stderr.log")
        stdout_file = stdout_path.open("wb")
        stderr_file = stderr_path.open("wb")
        handles.extend([stdout_file, stderr_file])
        log_paths.extend([stdout_path, stderr_path])
        proc = subprocess.Popen(
            sandbox_command(backend, profile, command),
            cwd=repo,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=stdout_file,
            stderr=stderr_file,
            start_new_session=False,
            preexec_fn=lambda: preexec(limits),
        )
        tracked_pids.add(proc.pid)
        tracked_groups.add(proc.pid)
        job = {"name": name, "argv": command, "proc": proc, "stdout": stdout_path, "stderr": stderr_path}
        jobs.append(job)
        return job

    def sample() -> tuple[set[int], dict[int, tuple[int, int, int]]]:
        nonlocal hit, peak_processes, peak_memory_bytes, peak_disk_bytes, peak_output_bytes, last_disk_sample
        snapshot = process_snapshot()
        tracked_pids.update(descendant_pids(tracked_pids, snapshot))
        live = {pid for pid in tracked_pids if pid in snapshot}
        tracked_groups.update(snapshot[pid][1] for pid in live)
        current_memory = sum(snapshot[pid][2] for pid in live)
        current_output = sum(path.stat().st_size for path in log_paths if path.exists())
        peak_processes = max(peak_processes, len(live))
        peak_memory_bytes = max(peak_memory_bytes, current_memory)
        peak_output_bytes = max(peak_output_bytes, current_output)
        now = time.perf_counter()
        current_disk = peak_disk_bytes
        if now - last_disk_sample >= 0.25:
            current_disk = max(0, directory_size(lab_root) - baseline_size)
            peak_disk_bytes = max(peak_disk_bytes, current_disk)
            last_disk_sample = now
        elapsed = now - started
        if elapsed > limits["timeout_seconds"]:
            hit = hit or "timeout"
        elif len(live) > limits["max_processes"]:
            hit = hit or "process"
        elif current_memory > limits["max_memory_bytes"]:
            hit = hit or "memory"
        elif current_disk > limits["max_disk_bytes"]:
            hit = hit or "disk"
        elif current_output > limits["max_output_bytes"]:
            hit = hit or "output"
        return live, snapshot

    main_job: dict[str, Any] | None = None
    try:
        for service in services:
            service["job"] = start_job(service["id"], service["command"])
            deadline = time.perf_counter() + service["readiness_timeout_seconds"]
            while not service_ready(service):
                live, _ = sample()
                code = service["job"]["proc"].poll()
                if code is not None:
                    unexpected_service_codes.append({"service_id": service["id"], "exit_code": code})
                    service_failure = f"service {service['id']} exited before readiness"
                    break
                if hit or time.perf_counter() > deadline:
                    service_failure = f"service {service['id']} did not become ready within its bound"
                    break
                time.sleep(0.05)
            service["ready"] = service_failure is None
            if service_failure:
                break
        if not service_failure and not hit:
            main_job = start_job("command", argv)
            while main_job["proc"].poll() is None:
                live, current_snapshot = sample()
                for service in services:
                    code = service["job"]["proc"].poll()
                    if code is not None and not any(row["service_id"] == service["id"] for row in unexpected_service_codes):
                        unexpected_service_codes.append({"service_id": service["id"], "exit_code": code})
                        service_failure = f"service {service['id']} exited before command completion"
                if hit or service_failure:
                    signal_processes(live, {current_snapshot[pid][1] for pid in live}, signal.SIGKILL)
                    break
                time.sleep(0.05)
            try:
                return_code = main_job["proc"].wait(timeout=2)
            except subprocess.TimeoutExpired:
                hit = hit or "process_cleanup"
                live, current_snapshot = sample()
                signal_processes(live, {current_snapshot[pid][1] for pid in live}, signal.SIGKILL)
                return_code = main_job["proc"].wait(timeout=2)
        else:
            return_code = 125

        before_cleanup, cleanup_snapshot = sample()
        cwd_residue = lab_cwd_pids(lab_root)
        tracked_pids.update(cwd_residue)
        cleanup_snapshot = process_snapshot()
        before_cleanup.update(pid for pid in cwd_residue if pid in cleanup_snapshot)
        peak_processes = max(peak_processes, len(before_cleanup))
        cleanup_memory = sum(cleanup_snapshot[pid][2] for pid in before_cleanup)
        peak_memory_bytes = max(peak_memory_bytes, cleanup_memory)
        if len(before_cleanup) > limits["max_processes"]:
            hit = hit or "process"
        elif cleanup_memory > limits["max_memory_bytes"]:
            hit = hit or "memory"
        declared_service_pids = {service["job"]["proc"].pid for service in services if "job" in service}
        main_pid = main_job["proc"].pid if main_job else None
        process_residue_detected = any(pid not in declared_service_pids and pid != main_pid for pid in before_cleanup)
        signal_processes(before_cleanup, {cleanup_snapshot[pid][1] for pid in before_cleanup}, signal.SIGTERM)
        time.sleep(0.1)
        live, current_snapshot = sample()
        if live:
            signal_processes(live, {current_snapshot[pid][1] for pid in live}, signal.SIGKILL)
        for job in jobs:
            try:
                job["proc"].wait(timeout=2)
            except subprocess.TimeoutExpired:
                hit = hit or "process_cleanup"
        tracked_pids.update(lab_cwd_pids(lab_root))
        final_snapshot = process_snapshot()
        live_residue = {pid for pid in tracked_pids if pid in final_snapshot}
        if live_residue:
            signal_processes(live_residue, {final_snapshot[pid][1] for pid in live_residue}, signal.SIGKILL)
            for _ in range(20):
                time.sleep(0.02)
                final_snapshot = process_snapshot()
                live_residue = {pid for pid in tracked_pids if pid in final_snapshot}
                if not live_residue:
                    break
        processes_remaining = len(live_residue)
    finally:
        try:
            tracked_pids.update(lab_cwd_pids(lab_root))
            emergency_snapshot = process_snapshot()
            emergency_live = {pid for pid in tracked_pids if pid in emergency_snapshot}
            if emergency_live:
                signal_processes(
                    emergency_live,
                    {emergency_snapshot[pid][1] for pid in emergency_live},
                    signal.SIGKILL,
                )
            for job in jobs:
                try:
                    job["proc"].wait(timeout=1)
                except subprocess.TimeoutExpired:
                    pass
        finally:
            for handle in handles:
                handle.close()

    final_disk_bytes = max(0, directory_size(lab_root) - baseline_size)
    final_output_bytes = sum(path.stat().st_size for path in log_paths if path.exists())
    peak_disk_bytes = max(peak_disk_bytes, final_disk_bytes)
    peak_output_bytes = max(peak_output_bytes, final_output_bytes)
    if final_disk_bytes > limits["max_disk_bytes"]:
        hit = hit or "disk"
    if final_output_bytes > limits["max_output_bytes"]:
        hit = hit or "output"
    policy_violations: list[dict[str, Any]] = []
    if hit is None and return_code == -signal.SIGKILL:
        policy_violations.append({
            "category": "sandbox_policy_denial",
            "mechanism": "uncatchable_signal_killed_offending_sandbox_process",
            "signal": signal.SIGKILL,
        })
    for row in unexpected_service_codes:
        if hit is None and row["exit_code"] == -signal.SIGKILL:
            policy_violations.append({
                "category": "sandbox_policy_denial",
                "service_id": row["service_id"],
                "mechanism": "uncatchable_signal_killed_offending_sandbox_process",
                "signal": signal.SIGKILL,
            })
    output_max = int(limits["max_output_bytes"])
    stdout_bytes = read_file_limited(main_job["stdout"], output_max, "command stdout") if main_job and main_job["stdout"].exists() else b""
    stderr_bytes = read_file_limited(main_job["stderr"], output_max, "command stderr") if main_job and main_job["stderr"].exists() else b""
    stdout = stdout_bytes[:output_max].decode(errors="replace")
    remaining_output = max(0, output_max - len(stdout.encode()))
    stderr = stderr_bytes[:remaining_output].decode(errors="replace")
    service_records: list[dict[str, Any]] = []
    for service in services:
        job = service.get("job")
        service_records.append({
            "id": service["id"],
            "transport": service["transport"],
            "endpoint": service["evidence_endpoint"],
            "local_only": True,
            "command": service["command"],
            "ready": service.get("ready", False),
            "exit_code_before_cleanup": next((row["exit_code"] for row in unexpected_service_codes if row["service_id"] == service["id"]), None),
            "cleanup_verified": bool(job) and job["proc"].poll() is not None,
        })
    return {
        "argv": argv,
        "exit_code": return_code,
        "stdout": stdout,
        "stderr": stderr,
        "duration_ms": round((time.perf_counter() - started) * 1000, 3),
        "limit_hit": hit,
        "process_residue_detected": process_residue_detected,
        "processes_remaining": processes_remaining,
        "peak_processes": peak_processes,
        "peak_memory_bytes": peak_memory_bytes,
        "peak_disk_bytes": peak_disk_bytes,
        "final_disk_bytes": final_disk_bytes,
        "peak_output_bytes": peak_output_bytes,
        "final_output_bytes": final_output_bytes,
        "sandbox_denial_observed": bool(policy_violations),
        "policy_violations": policy_violations,
        "service_failure": service_failure,
        "services": service_records,
        "sandbox": backend,
        "profile_sha256": hashlib.sha256(profile.read_bytes()).hexdigest(),
    }


def clone_exact(candidate: Path, head: str, lab_root: Path, env: dict[str, str]) -> Path:
    repo = lab_root / "repo"
    proc = subprocess.run(
        ["git", "-c", "protocol.file.allow=always", "clone", "--no-hardlinks", "--no-local", "--quiet", "--no-checkout", str(candidate), str(repo)],
        check=False,
        capture_output=True,
        text=True,
        env=env,
    )
    if proc.returncode != 0:
        raise LabError(f"exact-head lab clone failed: {(proc.stderr or proc.stdout).strip()[:500]}")
    checkout = subprocess.run(["git", "-C", str(repo), "checkout", "--quiet", "--detach", head], check=False, capture_output=True, text=True, env=env)
    if checkout.returncode != 0:
        raise LabError(f"exact-head lab checkout failed: {(checkout.stderr or checkout.stdout).strip()[:500]}")
    subprocess.run(["git", "-C", str(repo), "remote", "remove", "origin"], check=False, capture_output=True, env=env)
    subprocess.run(["git", "-C", str(repo), "config", "core.hooksPath", os.devnull], check=True, capture_output=True, env=env)
    return repo


def normalize_service_request(value: Any) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise LabError("each requested service must be an object")
    allowed = {
        "id", "transport", "command", "path", "host", "port", "readiness",
        "readiness_timeout_seconds",
    }
    if set(value) - allowed:
        raise LabError("requested service contains unknown fields")
    identifier = validate_id(value.get("id"), "service id")
    transport = value.get("transport")
    if transport not in {"unix", "tcp"}:
        raise LabError("service transport must be unix or tcp")
    command = validate_command_vector(value.get("command"), "service command")
    timeout = value.get("readiness_timeout_seconds", 15.0)
    if not isinstance(timeout, (int, float)) or timeout <= 0 or timeout > 60:
        raise LabError("service readiness timeout must be in (0, 60] seconds")
    if transport == "unix":
        path = value.get("path", f"services/{identifier}.sock")
        normalized_path = normalize_root_spec(path, "service socket path")
        if not normalized_path.startswith("services/"):
            raise LabError("Unix service socket must be under the declared services root")
        readiness = value.get("readiness", "path_exists")
        if readiness not in {"path_exists", "connect"}:
            raise LabError("Unix service readiness must be path_exists or connect")
        return {
            "id": identifier,
            "transport": transport,
            "command": command,
            "path_spec": normalized_path,
            "readiness": readiness,
            "readiness_timeout_seconds": float(timeout),
        }
    host = value.get("host", "127.0.0.1")
    port = value.get("port", 0)
    if host not in {"127.0.0.1", "::1", "localhost"} or not isinstance(port, int) or not (0 <= port <= 65535):
        raise LabError("TCP service must use loopback and a valid bounded port")
    readiness = value.get("readiness", "connect")
    if readiness != "connect":
        raise LabError("TCP service readiness must be connect")
    return {
        "id": identifier,
        "transport": transport,
        "command": command,
        "host": "127.0.0.1" if host == "localhost" else host,
        "port": port,
        "readiness": readiness,
        "readiness_timeout_seconds": float(timeout),
    }


def request_document(path: Path, maximum: int) -> dict[str, Any]:
    try:
        size = path.stat().st_size
    except OSError as exc:
        raise LabError(f"cannot stat lab request: {exc}") from exc
    if size > maximum:
        raise LabError("request byte bound exceeded")
    try:
        value = json.loads(read_file_limited(path, maximum, "lab request").decode())
    except (OSError, UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise LabError(f"cannot read lab request: {exc}") from exc
    if not isinstance(value, dict) or value.get("schema_version") != LAB_REQUEST_SCHEMA:
        raise LabError(f"lab request migration required: {value.get('schema_version') if isinstance(value, dict) else None!r}")
    if "captain_approval" in value or "captain_approvals" in value:
        raise LabError("captain approvals must be supplied separately through trusted --approval paths")
    if "hypothesis_id" not in value or "alternative" not in value or "command" not in value:
        raise LabError("lab request lacks required hypothesis/alternative/command fields")
    validate_id(value["hypothesis_id"], "hypothesis id")
    claim = value.get("hypothesis", value.get("claim"))
    alternative = value.get("alternative")
    expected = value.get("discriminator", value.get("expected_discriminator"))
    if not isinstance(claim, str) or not claim.strip() or not isinstance(alternative, str) or not alternative.strip():
        raise LabError("hypothesis/claim and alternative must be non-empty")
    if not isinstance(expected, dict) or not expected:
        raise LabError("discriminator/expected_discriminator must be a non-empty object")
    value["claim"] = claim
    value["expected_discriminator"] = expected
    value["impact_edges_examined"] = value.get("impact_edges_examined", [])
    if not isinstance(value["impact_edges_examined"], list) or not all(isinstance(item, str) for item in value["impact_edges_examined"]):
        raise LabError("impact_edges_examined must be a string list")
    value["changes"] = value.get("changes", [])
    if not isinstance(value["changes"], list):
        raise LabError("changes must be a list")
    value["temporary_change"] = value.get(
        "temporary_change",
        "none; experiment uses only declared disposable lab effects",
    )
    if not isinstance(value["temporary_change"], str) or not value["temporary_change"].strip():
        raise LabError("temporary_change must be a non-empty description")
    validate_command_vector(value["command"])
    explicit = "mode" in value
    if explicit:
        if value["mode"] not in EXPERIMENT_MODES:
            raise LabError("experiment mode is unknown")
        value["allowed_read_roots"] = normalize_root_specs(value.get("allowed_read_roots"), "allowed_read_roots")
        value["allowed_write_roots"] = normalize_root_specs(value.get("allowed_write_roots"), "allowed_write_roots", allow_empty=True)
    else:
        # Existing v2 callers are a counterfactual-edit contract. Keep them operational while all
        # new human-like requests carry explicit mode/root declarations.
        value["mode"] = "counterfactual_edit"
        value["allowed_read_roots"] = normalize_root_specs(LEGACY_READ_ROOTS, "allowed_read_roots")
        value["allowed_write_roots"] = normalize_root_specs(LEGACY_WRITE_ROOTS, "allowed_write_roots")
    raw_services = value.get("services", [])
    if not isinstance(raw_services, list):
        raise LabError("services must be a list")
    value["services"] = [normalize_service_request(item) for item in raw_services]
    if len({service["id"] for service in value["services"]}) != len(value["services"]):
        raise LabError("service ids must be unique")
    if value["mode"] == "local_service_simulation" and not value["services"]:
        raise LabError("local_service_simulation requires at least one declared service")
    if value["mode"] != "local_service_simulation" and value["services"]:
        raise LabError("services are allowed only in local_service_simulation mode")
    if value["mode"] == "read_inspect" and value["changes"]:
        raise LabError("read_inspect cannot apply trusted temporary changes")
    if value.get("change_scope", "lab") != "lab":
        raise LabError("canonical candidate changes are denied; only the disposable lab is writable")
    if "resource_envelope" in value:
        if "limits" in value and value["limits"] != value["resource_envelope"]:
            raise LabError("limits and resource_envelope conflict")
        value["limits"] = value["resource_envelope"]
    declared_capability_details(value)
    for command in request_commands(value):
        denial = unconditional_command_denial(command)
        if denial:
            raise LabError(denial)
    value["legacy_contract"] = not explicit
    return value


def expand_command(command: list[str], placeholders: dict[str, str]) -> list[str]:
    result: list[str] = []
    for item in command:
        expanded = item
        for key, replacement in placeholders.items():
            expanded = expanded.replace("{" + key + "}", replacement)
        if "{" in expanded or "}" in expanded:
            raise LabError("command contains an unresolved lab/service placeholder")
        if CONTROL.search(expanded):
            raise LabError("expanded command contains a control character")
        result.append(expanded)
    return result


def prepare_services(
    request: dict[str, Any],
    lab_root: Path,
    repo: Path,
    env: dict[str, str],
    readable_roots: list[Path],
    writable_roots: list[Path],
) -> tuple[list[dict[str, Any]], dict[str, str]]:
    placeholders = {
        "repo": str(repo),
        "home": str(lab_root / "home"),
        "tmp": str(lab_root / "tmp"),
        "xdg_config": str(lab_root / "xdg-config"),
        "xdg_cache": str(lab_root / "xdg-cache"),
        "xdg_data": str(lab_root / "xdg-data"),
        "xdg_state": str(lab_root / "xdg-state"),
        "go_cache": str(lab_root / "go-cache"),
        "go_mod_cache": str(lab_root / "go-mod-cache"),
        "packages": str(lab_root / "packages"),
        "services": str(lab_root / "services"),
        "credentials": str(lab_root / "credentials"),
    }
    prepared: list[dict[str, Any]] = []
    for raw in request["services"]:
        service = dict(raw)
        if service["transport"] == "unix":
            path = resolve_root_spec(lab_root, repo, service["path_spec"])
            if len(os.fsencode(path)) >= 100:
                raise LabError("Unix service socket path exceeds the platform-safe bound")
            if not is_within(path, readable_roots) or not is_within(path, writable_roots):
                raise LabError("service endpoint is outside declared read/write roots")
            service["path"] = str(path)
            service["evidence_endpoint"] = f"unix:{service['path_spec']}"
            endpoint = str(path)
            placeholders[f"service.{service['id']}.socket"] = endpoint
            placeholders[f"service.{service['id']}.endpoint"] = endpoint
        else:
            port = service["port"]
            if port == 0:
                family = socket.AF_INET6 if service["host"] == "::1" else socket.AF_INET
                with socket.socket(family, socket.SOCK_STREAM) as allocator:
                    allocator.bind((service["host"], 0))
                    port = allocator.getsockname()[1]
            service["port"] = port
            endpoint = f"{service['host']}:{port}"
            service["evidence_endpoint"] = f"tcp://{endpoint}"
            placeholders[f"service.{service['id']}.host"] = service["host"]
            placeholders[f"service.{service['id']}.port"] = str(port)
            placeholders[f"service.{service['id']}.endpoint"] = endpoint
        env_key = re.sub(r"[^A-Za-z0-9]", "_", service["id"]).upper()
        env[f"PM_REVIEW_SERVICE_{env_key}_ENDPOINT"] = endpoint
        if service["transport"] == "unix":
            env[f"PM_REVIEW_SERVICE_{env_key}_SOCKET"] = endpoint
        else:
            env[f"PM_REVIEW_SERVICE_{env_key}_HOST"] = service["host"]
            env[f"PM_REVIEW_SERVICE_{env_key}_PORT"] = str(service["port"])
        service["command"] = command_policy(
            expand_command(service["command"], placeholders), repo, lab_root, readable_roots,
        )
        prepared.append(service)
    return prepared, placeholders


def materialize_credentials(
    proposals: list[dict[str, Any]],
    candidate: Path,
    lab_root: Path,
    env: dict[str, str],
    readable_roots: list[Path],
    maximum: int,
) -> tuple[list[dict[str, str]], list[bytes]]:
    records: list[dict[str, str]] = []
    sensitive_values: list[bytes] = []
    remaining = [maximum]
    for proposal in proposals:
        if proposal["capability"] != "host_credential":
            continue
        credential = proposal["credential"]
        source_raw = Path(credential["source_path"])
        if source_raw.is_symlink():
            raise LabError("approved credential source must not be a symlink")
        try:
            source = source_raw.resolve(strict=True)
        except OSError as exc:
            raise LabError(f"approved credential source cannot be resolved: {exc}") from exc
        if is_within(source, [candidate.resolve()]):
            raise LabError("host credential source must not be inside the primary candidate")
        if not source.is_file():
            raise LabError("approved credential source must be a regular file")
        mount_name = validate_relative(credential["mount_name"], "credential mount name")
        target = (lab_root / "credentials" / mount_name).resolve(strict=False)
        if not is_within(target, [lab_root / "credentials"]) or not is_within(target, readable_roots):
            raise LabError("credential mount is outside declared credential read root")
        payload = read_file_limited(source, remaining[0], "approved host credential")
        remaining[0] -= len(payload)
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(payload)
        if payload:
            sensitive_values.append(payload)
        del payload  # Credential values never enter evidence.
        target.chmod(0o400)
        env_key = re.sub(r"[^A-Za-z0-9]", "_", Path(mount_name).stem).upper()
        env[f"PM_REVIEW_CREDENTIAL_{env_key}"] = str(target)
        records.append({"mount_name": mount_name, "scope": credential["scope"], "access": "read_only"})
    return records, sensitive_values


def redact_sensitive_text(value: str, sensitive_values: list[bytes]) -> tuple[str, bool]:
    payload = value.encode(errors="replace")
    matched = False
    for secret in sensitive_values:
        if secret and secret in payload:
            payload = payload.replace(secret, b"[REDACTED_CREDENTIAL]")
            matched = True
    return payload.decode(errors="replace"), matched


def sanitize_sensitive_effects(
    lab_root: Path,
    repo: Path,
    worktree: dict[str, Any],
    dummy: dict[str, Any],
    stdout: str,
    stderr: str,
    sensitive_values: list[bytes],
    maximum: int,
) -> tuple[str, str, list[str]]:
    if not sensitive_values:
        return stdout, stderr, []
    matches: list[str] = []
    stdout, stdout_match = redact_sensitive_text(stdout, sensitive_values)
    stderr, stderr_match = redact_sensitive_text(stderr, sensitive_values)
    if stdout_match:
        matches.append("stdout")
    if stderr_match:
        matches.append("stderr")
    patch = base64.b64decode(worktree.get("tracked_diff_base64", ""), validate=True)
    if any(secret and secret in patch for secret in sensitive_values):
        matches.append("worktree_diff")
        worktree["tracked_diff_base64"] = None
        worktree["tracked_diff_sha256"] = None
        worktree["sha256"] = None
        worktree["redacted_sensitive_content"] = True
    remaining = [maximum]
    worktree_output_match = False
    for row in worktree.get("untracked", []):
        if row.get("kind") != "file":
            continue
        path = repo / row["path"]
        payload = read_file_limited(path, remaining[0], "sensitive untracked-output scan")
        remaining[0] -= len(payload)
        if any(secret and secret in payload for secret in sensitive_values):
            matches.append(f"worktree:{row['path']}")
            row.pop("sha256", None)
            row["redacted_sensitive_content"] = True
            worktree_output_match = True
    if worktree_output_match:
        worktree["sha256"] = None
    dummy_output_match = False
    for path_value in dummy.get("created", []) + dummy.get("modified", []):
        row = dummy.get("entries", {}).get(path_value, {})
        path = lab_root / path_value
        if row.get("kind") != "file" or not path.exists():
            continue
        payload = read_file_limited(path, remaining[0], "sensitive lab-output scan")
        remaining[0] -= len(payload)
        if any(secret and secret in payload for secret in sensitive_values):
            matches.append(f"lab:{path_value}")
            row.pop("sha256", None)
            row["redacted_sensitive_content"] = True
            dummy_output_match = True
    if dummy_output_match:
        dummy["sha256"] = None
    # Exact matching catches direct disclosure attempts. For any credentialled experiment, suppress
    # all child-controlled text and content-derived hashes as well so encoded/transformed secrets
    # cannot enter the evidence envelope.
    stdout = "[SUPPRESSED_CREDENTIALLED_EXPERIMENT_OUTPUT]" if stdout else ""
    stderr = "[SUPPRESSED_CREDENTIALLED_EXPERIMENT_OUTPUT]" if stderr else ""
    worktree["tracked_diff_base64"] = None
    worktree["tracked_diff_sha256"] = None
    worktree["sha256"] = None
    worktree["credential_redaction_applied"] = True
    for row in worktree.get("untracked", []):
        row.pop("sha256", None)
    dummy["sha256"] = None
    dummy["credential_redaction_applied"] = True
    for row in dummy.get("entries", {}).values():
        if isinstance(row, dict):
            row.pop("sha256", None)
    return stdout, stderr, sorted(matches)


def validate_external_command_scope(command: list[str], proposals: list[dict[str, Any]]) -> None:
    if not any(proposal["capability"] == "external_network" for proposal in proposals):
        return
    tool = Path(command[0]).name.lower()
    raw_args = command[1:]
    args = [item.lower() for item in raw_args]
    package_approved = any(proposal["capability"] == "package_install" for proposal in proposals)
    live_approved = any(proposal["capability"] in {"live_connector", "live_connector_write"} for proposal in proposals)
    if tool.startswith("python") and args[:2] in (["-m", "pip"], ["-m", "ensurepip"]) and package_approved:
        return
    if tool in {"pip", "pip3", "npm", "npx", "yarn", "pnpm", "gem", "bundle"} and package_approved:
        return
    if tool == "curl":
        write_flags = {"-d", "-F", "-T", "--data", "--data-raw", "--data-binary", "--form", "--upload-file"}
        if any(arg in write_flags for arg in raw_args) or any(arg.startswith(("--data=", "--form=", "--upload-file=")) for arg in args):
            raise LabError("approved curl command attempts an external write")
        for index, arg in enumerate(args[:-1]):
            if arg in {"-x", "--request"} and args[index + 1].upper() not in {"GET", "HEAD"}:
                raise LabError("approved curl method is not read-only")
        return
    if tool == "wget":
        if any(arg.startswith(("--post-data", "--post-file", "--method=")) for arg in args):
            raise LabError("approved wget command attempts an external write")
        return
    if tool == "pm" and live_approved:
        return
    raise LabError("approved external network command lacks an enforceable minimum-scope command adapter")


def validate_approved_capability_scope(
    proposals: list[dict[str, Any]],
    lab_root: Path,
    writable_roots: list[Path],
) -> None:
    for proposal in proposals:
        if proposal["capability"] == "package_install":
            package = proposal["package"]
            checksum = package["checksum"]
            if not re.fullmatch(r"sha256:[0-9a-f]{64}", checksum):
                raise LabError("approved package checksum must be an exact lowercase sha256 digest")
            pinned_forms = {f"{package['name']}=={package['version']}", f"{package['name']}@{package['version']}"}
            if not any(item in pinned_forms for command in proposal["commands"] for item in command):
                raise LabError("approved package command does not use the exact proposed name/version pin")
            if not is_within(lab_root / "packages", writable_roots):
                raise LabError("approved package installation requires the declared lab-local packages write root")
        if proposal["capability"] == "live_connector" and proposal["connector"]["access_mode"] != "read_only":
            raise LabError("live connector approval is read-only by default")


def blocked_envelope(base: str | None, head: str | None, packet: str | None, claim: str) -> dict[str, Any]:
    return {
        "schema_version": LAB_EVIDENCE_SCHEMA,
        "status": "blocked",
        "exact_base_sha": base,
        "exact_head_sha": head,
        "packet_id": packet,
        "candidate_unchanged": False,
        "lab_cleanup_verified": False,
        "blockers": [{"category": "lab_safety", "claim": claim}],
    }


def approval_required_envelope(
    base: str,
    head: str,
    tree: str,
    packet: str,
    candidate_identity_value: dict[str, str],
    proposals: list[dict[str, Any]],
) -> dict[str, Any]:
    return {
        "schema_version": LAB_EVIDENCE_SCHEMA,
        "status": "blocked",
        "decision": "captain_approval_required",
        "exact_base_sha": base,
        "exact_head_sha": head,
        "exact_head_tree": tree,
        "packet_id": packet,
        "candidate_identity_before": candidate_identity_value,
        "candidate_identity_after": candidate_identity_value,
        "candidate_unchanged": True,
        "lab_cleanup_verified": True,
        "captain_approval_required": proposals,
        "blockers": [{
            "category": "captain_approval_required",
            "claim": "approval-gated capability was not executed; supply exact trusted --approval object(s)",
        }],
    }


def command_probe(_: argparse.Namespace) -> int:
    backend = sandbox_backend()
    emit({"schema_version": LAB_EVIDENCE_SCHEMA, "status": "ready" if backend["status"] == "available" else "blocked", "backend": backend})
    return 0 if backend["status"] == "available" else 1


def command_run(args: argparse.Namespace) -> int:
    base = head = packet_id = None
    candidate: Path | None = None
    lab_root: Path | None = None
    candidate_before: dict[str, str] | None = None
    evidence: dict[str, Any] | None = None
    cleanup_verified = False
    try:
        base = validate_sha(args.base, "exact base")
        head = validate_sha(args.head, "exact head")
        packet_id = validate_id(args.packet_id, "packet id")
        candidate = Path(args.repo_root).resolve(strict=True)
        temp_root = Path(args.temp_root).resolve(strict=True)
        if not candidate.is_dir() or not temp_root.is_dir() or temp_root == candidate or candidate in temp_root.parents or temp_root in candidate.parents:
            raise LabError("private temp root must be an existing directory outside the candidate")
        temp_root.chmod(0o700)
        candidate_before = candidate_identity(candidate)
        if candidate_before["head"] != head or candidate_before["status"] != "clean":
            raise LabError("candidate is not the clean exact reviewed head")
        actual_base = run_git(candidate, ["merge-base", base, head]).strip()
        if actual_base != base:
            raise LabError("exact base is not the merge base of the candidate")

        configured_limits = merge_limits(None)
        request_path = Path(args.request).resolve(strict=True)
        request = request_document(request_path, int(configured_limits["max_request_bytes"]))
        limits = merge_limits(request.get("limits"))
        proposals = approval_proposals(base, head, candidate_before["tree"], packet_id, request, limits)
        approvals = validate_approvals(args.approval, proposals, int(limits["max_request_bytes"]), candidate)
        if proposals and not approvals:
            evidence = approval_required_envelope(
                base, head, candidate_before["tree"], packet_id, candidate_before, proposals,
            )
        else:
            backend = sandbox_backend()
            if backend["status"] != "available":
                raise LabError(backend["reason"])
            lab_root = Path(tempfile.mkdtemp(prefix=f"pm-review-{packet_id}-", dir=temp_root))
            lab_root.chmod(0o700)
            for name in LAB_DIRECTORY_NAMES:
                (lab_root / name).mkdir(mode=0o700)
            env = safe_environment(lab_root)
            repo = clone_exact(candidate, head, lab_root, env)
            if run_git(repo, ["rev-parse", "HEAD"]).strip() != head or run_git(repo, ["rev-parse", "HEAD^{tree}"]).strip() != candidate_before["tree"]:
                raise LabError("lab snapshot identity does not match candidate")

            readable_roots = resolve_root_specs(lab_root, repo, request["allowed_read_roots"])
            writable_roots = resolve_root_specs(lab_root, repo, request["allowed_write_roots"])
            if not write_roots_are_readable(readable_roots, writable_roots):
                raise LabError("every declared write root must be covered by a declared read root")
            validate_approved_capability_scope(proposals, lab_root, writable_roots)
            credentials, sensitive_values = materialize_credentials(
                proposals, candidate, lab_root, env, readable_roots, int(limits["max_read_bytes"]),
            )
            services, placeholders = prepare_services(
                request, lab_root, repo, env, readable_roots, writable_roots,
            )
            command = command_policy(
                expand_command(request["command"], placeholders), repo, lab_root, readable_roots,
            )
            validate_external_command_scope(command, proposals)
            for service in services:
                validate_external_command_scope(service["command"], proposals)
            if any(proposal["capability"] == "package_install" for proposal in proposals):
                env["PYTHONPATH"] = str(lab_root / "packages")
            external_endpoints = external_network_endpoints(proposals)
            local_endpoints = [
                {"id": service["id"], "transport": service["transport"], "path": service.get("path"), "port": service.get("port")}
                for service in services
            ]
            profile = write_sandbox_profile(
                backend,
                lab_root,
                repo,
                readable_roots,
                writable_roots,
                [command, *[service["command"] for service in services]],
                local_endpoints,
                external_endpoints,
                allow_process_fork=not request["legacy_contract"],
            )
            git_admin_before = directory_hash(repo / ".git", int(limits["max_read_bytes"]))
            changed_paths = apply_changes(
                repo, request["changes"], int(limits["max_change_bytes"]), writable_roots,
            )
            temporary_diff = git_effects(repo, lab_root, env, limits)
            dummy_before = tree_snapshot(
                lab_root, repo, writable_roots, int(limits["max_read_bytes"]),
            )
            execution = execute(
                lab_root, repo, command, env, limits, backend, profile, services,
            )
            actual_worktree_effects = git_effects(repo, lab_root, env, limits)
            dummy_after = tree_snapshot(
                lab_root, repo, writable_roots, int(limits["max_read_bytes"]),
            )
            dummy_effects = compare_snapshots(dummy_before, dummy_after)
            execution["stdout"], execution["stderr"], sensitive_matches = sanitize_sensitive_effects(
                lab_root,
                repo,
                actual_worktree_effects,
                dummy_effects,
                execution["stdout"],
                execution["stderr"],
                sensitive_values,
                int(limits["max_read_bytes"]),
            )
            del sensitive_values
            lab_head = run_git(repo, ["rev-parse", "HEAD"]).strip()
            git_admin_after = directory_hash(repo / ".git", int(limits["max_read_bytes"]))
            candidate_after = candidate_identity(candidate)
            if os.environ.get("PM_REVIEW_LAB_TEST_FORCE_IDENTITY_DRIFT") == "1":
                candidate_after = {**candidate_after, "tree": "f" * 40}
            candidate_unchanged = candidate_after == candidate_before

            policy_violations = list(execution["policy_violations"])
            if sensitive_matches:
                policy_violations.append({
                    "category": "credential_exposure_attempt",
                    "locations": sensitive_matches,
                    "mechanism": "trusted_output_and_effect_scan_with_redaction",
                })
            for row in actual_worktree_effects["untracked"]:
                if row.get("escapes_lab"):
                    policy_violations.append({
                        "category": "lab_symlink_escape",
                        "path": row["path"],
                        "mechanism": "trusted_final_state_inspection",
                    })
            for path_value, row in dummy_effects["entries"].items():
                if isinstance(row, dict) and row.get("escapes_lab"):
                    policy_violations.append({
                        "category": "lab_symlink_escape",
                        "path": path_value,
                        "mechanism": "trusted_final_state_inspection",
                    })
            for path_value in actual_worktree_effects["changed_paths"]:
                if not is_within((repo / path_value).resolve(strict=False), writable_roots):
                    policy_violations.append({
                        "category": "undeclared_write_effect",
                        "path": path_value,
                        "mechanism": "trusted_final_state_inspection",
                    })

            expected = request["expected_discriminator"]
            observed = {
                "exit_code": execution["exit_code"],
                "limit_hit": execution["limit_hit"],
                "process_residue_detected": execution["process_residue_detected"],
                "processes_remaining": execution["processes_remaining"],
                "sandbox_denial_observed": execution["sandbox_denial_observed"],
                "policy_violation_count": len(policy_violations),
                "services_ready": all(service["ready"] for service in execution["services"]),
            }
            discriminator_matched = all(key in observed and observed[key] == value for key, value in expected.items())
            blockers: list[dict[str, str]] = []
            if not discriminator_matched:
                blockers.append({"category": "hypothesis_inconclusive", "claim": "observed result did not match the declared discriminator"})
            if execution["limit_hit"]:
                blockers.append({"category": "lab_limit", "claim": f"experiment hit {execution['limit_hit']} limit"})
            if policy_violations:
                blockers.append({"category": "lab_policy_violation", "claim": "trusted enforcement detected a denied or undeclared operation; evidence is unusable"})
            if execution["service_failure"]:
                blockers.append({"category": "lab_service", "claim": execution["service_failure"]})
            if execution["process_residue_detected"] or execution["processes_remaining"]:
                blockers.append({"category": "lab_cleanup", "claim": "experiment spawned residual processes; every tracked process group was terminated"})
            if request["legacy_contract"] and set(actual_worktree_effects["changed_paths"]) != set(changed_paths):
                blockers.append({"category": "lab_final_state", "claim": "legacy counterfactual final worktree differs from its declared temporary change set"})
            if lab_head != head or git_admin_after != git_admin_before:
                blockers.append({"category": "lab_git_mutation", "claim": "lab Git administration changed; commit and Git mutation are prohibited"})
            if not candidate_unchanged:
                blockers.append({"category": "candidate_identity", "claim": "canonical candidate identity changed"})

            evidence = {
                "schema_version": LAB_EVIDENCE_SCHEMA,
                "status": "blocked" if blockers else "evidence",
                "exact_base_sha": base,
                "exact_head_sha": head,
                "exact_head_tree": candidate_before["tree"],
                "packet_id": packet_id,
                "hypothesis_id": request["hypothesis_id"],
                "candidate_identity_before": candidate_before,
                "candidate_identity_after": candidate_after,
                "candidate_unchanged": candidate_unchanged,
                "approval": {
                    "required_capabilities": [proposal["capability"] for proposal in proposals],
                    "validated_approval_sha256": [hashlib.sha256(canonical_json_bytes(item)).hexdigest() for item in approvals],
                    "minimum_scope_validated": bool(approvals) if proposals else True,
                },
                "experiment": {
                    "mode": request["mode"],
                    "hypothesis_id": request["hypothesis_id"],
                    "claim": request["claim"],
                    "alternative": request["alternative"],
                    "impact_edges_examined": request["impact_edges_examined"],
                    "temporary_change": request["temporary_change"],
                    "allowed_read_roots": request["allowed_read_roots"],
                    "allowed_write_roots": request["allowed_write_roots"],
                    "command": execution["argv"],
                    "expected_discriminator": expected,
                    "observed": observed,
                    "discriminator_matched": discriminator_matched,
                    "exit_code": execution["exit_code"],
                    "stdout": execution["stdout"],
                    "stderr": execution["stderr"],
                    "duration_ms": execution["duration_ms"],
                    "temporary_diff": temporary_diff,
                    "actual_effects": {
                        "worktree": actual_worktree_effects,
                        "lab_owned_roots": dummy_effects,
                    },
                    "services": execution["services"],
                    "credentials": credentials,
                    "credential_output_suppressed": bool(credentials),
                    "external_network_policy": {
                        "approved": bool(external_endpoints),
                        "resolved_exact_endpoints": external_endpoints,
                    },
                    "policy_violations": policy_violations,
                    "sandbox": execution["sandbox"],
                    "sandbox_profile_sha256": execution["profile_sha256"],
                    "git_admin_sha256_before": git_admin_before,
                    "git_admin_sha256_after": git_admin_after,
                    "limits": limits,
                    "resource_usage": {
                        "peak_processes": execution["peak_processes"],
                        "peak_memory_bytes": execution["peak_memory_bytes"],
                        "peak_disk_bytes": execution["peak_disk_bytes"],
                        "final_disk_bytes": execution["final_disk_bytes"],
                        "peak_output_bytes": execution["peak_output_bytes"],
                        "final_output_bytes": execution["final_output_bytes"],
                    },
                },
                "final_state": {
                    "candidate_unchanged": candidate_unchanged,
                    "git_admin_unchanged": git_admin_after == git_admin_before,
                    "declared_worktree_state_only": not any(row["category"] == "undeclared_write_effect" for row in policy_violations),
                    "policy_violations_absent": not policy_violations,
                    "services_cleaned": all(service["cleanup_verified"] for service in execution["services"]),
                    "processes_remaining": execution["processes_remaining"],
                    "resource_bounds_satisfied": execution["limit_hit"] is None,
                    "cleanup_verified": False,
                },
                "lab_cleanup_verified": False,
                "blockers": blockers,
            }
    except (LabError, OSError, json.JSONDecodeError, subprocess.SubprocessError) as exc:
        evidence = blocked_envelope(base, head, packet_id, str(exc))
        if candidate_before is not None and candidate is not None:
            try:
                after = candidate_identity(candidate)
                if os.environ.get("PM_REVIEW_LAB_TEST_FORCE_IDENTITY_DRIFT") == "1":
                    after = {**after, "tree": "f" * 40}
                evidence["candidate_identity_before"] = candidate_before
                evidence["candidate_identity_after"] = after
                evidence["candidate_unchanged"] = after == candidate_before
            except (LabError, OSError):
                pass
    finally:
        if lab_root is not None:
            try:
                shutil.rmtree(lab_root)
                cleanup_verified = not lab_root.exists()
            except OSError:
                cleanup_verified = False
        else:
            cleanup_verified = True
        if os.environ.get("PM_REVIEW_LAB_TEST_FORCE_CLEANUP_FAILURE") == "1":
            cleanup_verified = False
        if evidence is None:
            evidence = blocked_envelope(base, head, packet_id, "lab failed without evidence")
        evidence["lab_cleanup_verified"] = cleanup_verified
        final_state = evidence.setdefault("final_state", {})
        final_state["cleanup_verified"] = cleanup_verified
        final_state["candidate_unchanged"] = evidence.get("candidate_unchanged") is True
        final_state.setdefault("resource_bounds_satisfied", False)
        if not cleanup_verified:
            evidence["status"] = "blocked"
            evidence.setdefault("blockers", []).append({"category": "lab_cleanup", "claim": "whole-lab destruction was not verified"})
        if evidence.get("candidate_unchanged") is not True:
            evidence["status"] = "blocked"
        emit(evidence)
    return 0 if evidence.get("status") == "evidence" else 1


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description="Run a bounded disposable exact-head PM review hypothesis lab")
    subparsers = result.add_subparsers(dest="command", required=True)
    probe = subparsers.add_parser("probe", help="report whether a proven OS sandbox backend is available")
    probe.set_defaults(func=command_probe)
    run = subparsers.add_parser("run", help="run one exact-head packet experiment and destroy its lab")
    run.add_argument("--repo-root", required=True)
    run.add_argument("--base", required=True)
    run.add_argument("--head", required=True)
    run.add_argument("--packet-id", required=True)
    run.add_argument("--request", required=True)
    run.add_argument("--approval", action="append", default=[], help="trusted path to one exact captain approval object; repeat per capability")
    run.add_argument("--temp-root", required=True)
    run.set_defaults(func=command_run)
    return result


def main() -> int:
    try:
        args = parser().parse_args()
        return args.func(args)
    except LabError as exc:
        emit(blocked_envelope(None, None, None, str(exc)))
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
