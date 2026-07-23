#!/usr/bin/env python3
"""Run one bounded counterfactual review experiment in a disposable exact-head lab.

The canonical candidate is never the writable experiment root. A clean experiment requires an OS
sandbox that denies network and writes outside the lab; unsupported platforms fail closed.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import platform
import re
import resource
import shutil
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path, PurePosixPath
from typing import Any

LAB_REQUEST_SCHEMA = "polymetrics.ai/pm-review-lab-request/v1"
LAB_EVIDENCE_SCHEMA = "polymetrics.ai/pm-review-lab-evidence/v1"
HEX_SHA = re.compile(r"^[0-9a-f]{40}$")
SAFE_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_.-]{0,79}$")
CONTROL = re.compile(r"[\x00-\x1f\x7f]")
DEFAULT_LIMITS = {
    "timeout_seconds": 30.0,
    "max_processes": 16,
    "max_disk_bytes": 64 * 1024 * 1024,
    "max_output_bytes": 256 * 1024,
    "max_request_bytes": 64 * 1024,
    "max_change_bytes": 1024 * 1024,
}
MINIMUM_LIMITS = {
    "timeout_seconds": 0.05,
    "max_processes": 1,
    "max_disk_bytes": 4096,
    "max_output_bytes": 1024,
}


class LabError(ValueError):
    """Expected fail-closed lab input or safety error."""


def emit(value: dict[str, Any]) -> None:
    json.dump(value, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")


def run_git(root: Path, args: list[str], *, env: dict[str, str] | None = None) -> str:
    command = ["git", "-C", str(root), "--no-optional-locks", *args]
    proc = subprocess.run(command, check=False, capture_output=True, text=True, env=env)
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise LabError(f"read-only Git identity command failed: {detail[:500]}")
    return proc.stdout


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


def safe_environment(lab_root: Path) -> dict[str, str]:
    python_bin = Path(sys.executable).resolve()
    path_parts = [str(python_bin.parent), "/usr/bin", "/bin", "/usr/sbin", "/sbin", "/opt/homebrew/bin", "/usr/local/bin"]
    return {
        "PATH": ":".join(dict.fromkeys(path_parts)),
        "HOME": str(lab_root / "home"),
        "TMPDIR": str(lab_root / "tmp"),
        "LANG": "C",
        "LC_ALL": "C",
        "PYTHONNOUSERSITE": "1",
        "PYTHONDONTWRITEBYTECODE": "1",
        "GIT_CONFIG_GLOBAL": os.devnull,
        "GIT_CONFIG_SYSTEM": os.devnull,
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
        "NO_PROXY": "",
        "no_proxy": "",
    }


def sandbox_backend() -> dict[str, Any]:
    if platform.system() == "Darwin" and Path("/usr/bin/sandbox-exec").is_file():
        return {"status": "available", "name": "macos_sandbox_exec", "network": "deny", "outside_write": "deny"}
    # Bubblewrap availability alone is not proof that user/network namespaces and the intended
    # mount policy work on this host. Until a backend-specific probe is implemented, Linux blocks
    # rather than downgrading to a policy-only claim.
    return {
        "status": "blocked",
        "name": "unavailable",
        "reason": "no proven sandbox-exec or bubblewrap backend; policy-only execution cannot authorize clean evidence",
    }


def sandbox_command(backend: dict[str, Any], lab_root: Path, repo: Path, command: list[str]) -> tuple[list[str], Path | None]:
    if backend["name"] == "macos_sandbox_exec":
        profile = lab_root / "sandbox.sb"
        escaped_lab = str(lab_root).replace('"', '\\"')
        denied_read_roots = {str(Path.home().resolve()), "/Volumes", "/private/tmp", "/tmp", "/var/tmp", "/Applications"}
        rules = [
            "(version 1)",
            "(allow default)",
            "(deny network*)",
            "(deny file-write*)",
            f'(allow file-write* (subpath "{escaped_lab}"))',
        ]
        for value in sorted(denied_read_roots):
            escaped = value.replace('"', '\\"')
            rules.append(f'(deny file-read* (subpath "{escaped}"))')
        # The lab is more specific than its private temporary parent and is the only reviewer data
        # root re-enabled. System runtime/framework reads remain available to execute allowlisted
        # compilers and parsers; candidate/home/temp/volume/application reads stay denied.
        rules.append(f'(allow file-read* (subpath "{escaped_lab}"))')
        profile.write_text("\n".join(rules) + "\n")
        return ["/usr/bin/sandbox-exec", "-f", str(profile), *command], profile
    if backend["name"] == "linux_bwrap":
        executable_roots = [path for path in ("/usr", "/bin", "/lib", "/lib64", "/sbin", "/opt") if Path(path).exists()]
        wrapped = ["bwrap", "--unshare-all", "--die-with-parent", "--new-session", "--proc", "/proc", "--dev", "/dev"]
        for path in executable_roots:
            wrapped.extend(["--ro-bind", path, path])
        wrapped.extend(["--bind", str(lab_root), str(lab_root), "--chdir", str(repo), *command])
        return wrapped, None
    raise LabError("sandbox backend is unavailable")


def command_policy(argv: Any, repo: Path) -> list[str]:
    if not isinstance(argv, list) or not argv or not all(isinstance(item, str) and item and not CONTROL.search(item) for item in argv):
        raise LabError("command must be a non-empty control-free argument vector")
    tool = Path(argv[0]).name
    args = argv[1:]
    denied_tools = {
        "bash", "sh", "zsh", "fish", "curl", "wget", "nc", "netcat", "ssh", "scp", "rsync",
        "kubectl", "helm", "terraform", "podman", "docker", "npm", "npx", "pip", "pip3", "make",
        "pm", "gh", "gh-axi",
    }
    if tool in denied_tools:
        raise LabError(f"command category is denied: {tool}")
    for arg in args:
        if arg.startswith(("http://", "https://", "ssh://", "git@")) or CONTROL.search(arg):
            raise LabError("network/control argument is denied")
        if not arg.startswith("-") and (PurePosixPath(arg).is_absolute() or ".." in PurePosixPath(arg).parts):
            raise LabError("command path argument escapes the lab")
    if tool in {"python", "python3"}:
        if not args:
            raise LabError("interactive Python is denied")
        if args[0] == "-c" or args[0] in {"-", "-i"}:
            raise LabError("inline/interactive Python is denied")
        if args[:2] in (["-m", "pip"], ["-m", "venv"], ["-m", "ensurepip"]):
            raise LabError("dependency/environment installation is denied")
        if args[0] == "-m":
            if len(args) < 2 or args[1] not in {"py_compile", "unittest"}:
                raise LabError("Python module is not allowlisted")
        else:
            script = validate_relative(args[0], "Python script")
            path = (repo / script).resolve(strict=True)
            if repo.resolve() not in path.parents or not path.is_file() or path.suffix != ".py":
                raise LabError("Python script is outside the lab or not a file")
        child_python = shutil.which("python3", path="/usr/bin:/bin:/opt/homebrew/bin:/usr/local/bin")
        if not child_python:
            raise LabError("allowlisted system Python is unavailable")
        return [child_python, *args]
    if tool == "go":
        if not args or args[0] not in {"test", "vet", "build"}:
            raise LabError("Go command is not allowlisted")
        if any(arg.startswith(("-exec", "-toolexec", "-overlay", "-modfile")) for arg in args):
            raise LabError("Go execution/overlay override is denied")
        executable = shutil.which("go")
        if not executable:
            raise LabError("allowlisted Go tool is unavailable")
        return [executable, *args]
    if tool == "git":
        if not args or args[0] not in {"log", "show", "diff", "status", "blame", "grep", "rev-parse"}:
            raise LabError("Git mutation/remote command is denied")
        return ["git", "--no-optional-locks", *args]
    if tool == "shellcheck":
        executable = shutil.which("shellcheck")
        if not executable:
            raise LabError("allowlisted Shellcheck tool is unavailable")
        return [executable, *args]
    if tool == "ruby":
        if not args or args[0] != "-c":
            raise LabError("only Ruby syntax checking is allowlisted")
        executable = shutil.which("ruby")
        if not executable:
            raise LabError("allowlisted Ruby tool is unavailable")
        return [executable, *args]
    raise LabError(f"command tool is not allowlisted: {tool}")


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


def resolve_lab_file(repo: Path, relative: str) -> Path:
    normalized = validate_relative(relative, "temporary change path")
    candidate = repo / normalized
    try:
        resolved = candidate.resolve(strict=True)
    except OSError as exc:
        raise LabError(f"temporary change target cannot be resolved: {exc}") from exc
    if repo.resolve() not in resolved.parents or not resolved.is_file() or candidate.is_symlink():
        raise LabError("temporary change target escapes through an outside path or symlink")
    return resolved


def apply_changes(repo: Path, changes: Any, maximum: int) -> list[str]:
    if not isinstance(changes, list):
        raise LabError("changes must be a list")
    changed: list[str] = []
    total = 0
    for item in changes:
        if not isinstance(item, dict) or set(item) != {"path", "find", "replace"}:
            raise LabError("each temporary change must contain only path/find/replace")
        path = resolve_lab_file(repo, item["path"])
        before = path.read_text()
        find_text = item["find"]
        replace_text = item["replace"]
        if not isinstance(find_text, str) or not isinstance(replace_text, str) or not find_text:
            raise LabError("temporary find/replace values are malformed")
        if before.count(find_text) != 1:
            raise LabError("temporary replacement must match exactly once")
        after = before.replace(find_text, replace_text, 1)
        total += len(after.encode())
        if total > maximum:
            raise LabError("temporary change byte bound exceeded")
        path.write_text(after)
        changed.append(path.relative_to(repo).as_posix())
    return sorted(changed)


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


def process_group_count(group: int) -> int:
    proc = subprocess.run(["/bin/ps", "-axo", "pid=,pgid="], check=False, capture_output=True, text=True)
    count = 0
    for line in proc.stdout.splitlines():
        fields = line.split()
        if len(fields) == 2 and fields[1] == str(group):
            count += 1
    return count


def kill_group(group: int) -> None:
    try:
        os.killpg(group, signal.SIGKILL)
    except ProcessLookupError:
        pass


def preexec(limits: dict[str, Any]) -> None:
    os.setsid()
    resource.setrlimit(resource.RLIMIT_CORE, (0, 0))
    resource.setrlimit(resource.RLIMIT_NOFILE, (64, 64))
    cpu = max(1, int(math.ceil(float(limits["timeout_seconds"]))) + 1)
    resource.setrlimit(resource.RLIMIT_CPU, (cpu, cpu))
    file_size = int(max(limits["max_output_bytes"], limits["max_disk_bytes"]))
    resource.setrlimit(resource.RLIMIT_FSIZE, (file_size, file_size))


def execute(
    lab_root: Path,
    repo: Path,
    argv: list[str],
    env: dict[str, str],
    limits: dict[str, Any],
    backend: dict[str, Any],
) -> dict[str, Any]:
    wrapped, profile = sandbox_command(backend, lab_root, repo, argv)
    stdout_path = lab_root / "stdout.log"
    stderr_path = lab_root / "stderr.log"
    baseline_size = directory_size(lab_root)
    started = time.perf_counter()
    hit: str | None = None
    with stdout_path.open("wb") as stdout_file, stderr_path.open("wb") as stderr_file:
        proc = subprocess.Popen(
            wrapped,
            cwd=repo,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=stdout_file,
            stderr=stderr_file,
            start_new_session=False,
            preexec_fn=lambda: preexec(limits),
        )
        while proc.poll() is None:
            elapsed = time.perf_counter() - started
            if elapsed > limits["timeout_seconds"]:
                hit = "timeout"
            elif process_group_count(proc.pid) > limits["max_processes"]:
                hit = "process"
            elif directory_size(lab_root) - baseline_size > limits["max_disk_bytes"]:
                hit = "disk"
            elif stdout_path.stat().st_size + stderr_path.stat().st_size > limits["max_output_bytes"]:
                hit = "output"
            if hit:
                kill_group(proc.pid)
                break
            time.sleep(0.03)
        try:
            return_code = proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            hit = hit or "process_cleanup"
            kill_group(proc.pid)
            return_code = proc.wait(timeout=2)
    process_residue_detected = process_group_count(proc.pid) > 0
    if process_residue_detected:
        kill_group(proc.pid)
        for _ in range(20):
            if process_group_count(proc.pid) == 0:
                break
            time.sleep(0.02)
    processes_remaining = process_group_count(proc.pid)
    duration_ms = round((time.perf_counter() - started) * 1000, 3)
    output_max = int(limits["max_output_bytes"])
    stdout_bytes = stdout_path.read_bytes()
    stderr_bytes = stderr_path.read_bytes()
    if len(stdout_bytes) + len(stderr_bytes) > output_max:
        hit = hit or "output"
    stdout = stdout_bytes[:output_max].decode(errors="replace")
    remaining = max(0, output_max - len(stdout.encode()))
    stderr = stderr_bytes[:remaining].decode(errors="replace")
    return {
        "argv": argv,
        "exit_code": return_code,
        "stdout": stdout,
        "stderr": stderr,
        "duration_ms": duration_ms,
        "limit_hit": hit,
        "process_residue_detected": process_residue_detected,
        "processes_remaining": processes_remaining,
        "sandbox": backend,
        "profile_sha256": hashlib.sha256(profile.read_bytes()).hexdigest() if profile and profile.exists() else None,
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


def request_document(path: Path, maximum: int) -> dict[str, Any]:
    if path.stat().st_size > maximum:
        raise LabError("request byte bound exceeded")
    try:
        value = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise LabError(f"cannot read lab request: {exc}") from exc
    if not isinstance(value, dict) or value.get("schema_version") != LAB_REQUEST_SCHEMA:
        raise LabError(f"lab request migration required: {value.get('schema_version') if isinstance(value, dict) else None!r}")
    required = {"hypothesis_id", "claim", "alternative", "impact_edges_examined", "temporary_change", "changes", "command", "expected_discriminator"}
    if not required.issubset(value):
        raise LabError("lab request lacks required hypothesis fields")
    validate_id(value["hypothesis_id"], "hypothesis id")
    if not all(isinstance(value.get(field), str) and value[field].strip() for field in ("claim", "alternative", "temporary_change")):
        raise LabError("claim, alternative, and temporary change must be non-empty")
    if not isinstance(value["impact_edges_examined"], list) or not all(isinstance(item, str) for item in value["impact_edges_examined"]):
        raise LabError("impact_edges_examined must be a string list")
    return value


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


def command_probe(_: argparse.Namespace) -> int:
    backend = sandbox_backend()
    emit({"schema_version": LAB_EVIDENCE_SCHEMA, "status": "ready" if backend["status"] == "available" else "blocked", "backend": backend})
    return 0 if backend["status"] == "available" else 1


def command_run(args: argparse.Namespace) -> int:
    base = head = packet_id = None
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
        if not temp_root.is_dir() or temp_root == candidate or candidate in temp_root.parents or temp_root in candidate.parents:
            raise LabError("private temp root must be an existing directory outside the candidate")
        temp_root.chmod(0o700)
        candidate_before = candidate_identity(candidate)
        if candidate_before["head"] != head or candidate_before["status"] != "clean":
            raise LabError("candidate is not the clean exact reviewed head")
        actual_base = run_git(candidate, ["merge-base", base, head]).strip()
        if actual_base != base:
            raise LabError("exact base is not the merge base of the candidate")
        backend = sandbox_backend()
        if backend["status"] != "available":
            raise LabError(backend["reason"])
        request_path = Path(args.request).resolve(strict=True)
        limits = merge_limits(None)
        request = request_document(request_path, int(limits["max_request_bytes"]))
        limits = merge_limits(request.get("limits"))
        if request.get("change_scope", "lab") != "lab":
            raise LabError("canonical candidate changes are denied; only the disposable lab is writable")
        lab_root = Path(tempfile.mkdtemp(prefix=f"pm-review-{packet_id}-", dir=temp_root))
        lab_root.chmod(0o700)
        for name in ("home", "tmp", "go-cache", "go-mod-cache"):
            (lab_root / name).mkdir(mode=0o700)
        env = safe_environment(lab_root)
        repo = clone_exact(candidate, head, lab_root, env)
        if run_git(repo, ["rev-parse", "HEAD"]).strip() != head or run_git(repo, ["rev-parse", "HEAD^{tree}"]).strip() != candidate_before["tree"]:
            raise LabError("lab snapshot identity does not match candidate")
        object_count_before = run_git(repo, ["count-objects", "-v"])
        changed_paths = apply_changes(repo, request["changes"], int(limits["max_change_bytes"]))
        command = command_policy(request["command"], repo)
        diff_bytes = subprocess.check_output(["git", "-C", str(repo), "diff", "--binary", "--no-ext-diff", "--"], env=env)
        diff_stat = subprocess.check_output(["git", "-C", str(repo), "diff", "--stat", "--no-renames", "--"], env=env, text=True).strip()
        execution = execute(lab_root, repo, command, env, limits, backend)
        lab_head = run_git(repo, ["rev-parse", "HEAD"]).strip()
        object_count_after = run_git(repo, ["count-objects", "-v"])
        candidate_after = candidate_identity(candidate)
        if os.environ.get("PM_REVIEW_LAB_TEST_FORCE_IDENTITY_DRIFT") == "1":
            candidate_after = {**candidate_after, "tree": "f" * 40}
        candidate_unchanged = candidate_after == candidate_before
        expected = request["expected_discriminator"]
        discriminator_matched = isinstance(expected, dict) and expected.get("exit_code") == execution["exit_code"]
        blockers: list[dict[str, str]] = []
        if not discriminator_matched:
            blockers.append({"category": "hypothesis_inconclusive", "claim": "observed result did not match the declared discriminator"})
        if execution["limit_hit"]:
            blockers.append({"category": "lab_limit", "claim": f"experiment hit {execution['limit_hit']} limit"})
        if execution["process_residue_detected"] or execution["processes_remaining"]:
            blockers.append({"category": "lab_cleanup", "claim": "experiment spawned residual processes; process group was terminated"})
        if lab_head != head or object_count_after != object_count_before:
            blockers.append({"category": "lab_git_mutation", "claim": "lab Git identity/object store changed"})
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
            "experiment": {
                "hypothesis_id": request["hypothesis_id"],
                "claim": request["claim"],
                "alternative": request["alternative"],
                "impact_edges_examined": request["impact_edges_examined"],
                "temporary_change": request["temporary_change"],
                "command": execution["argv"],
                "expected_discriminator": expected,
                "observed": {
                    "exit_code": execution["exit_code"],
                    "limit_hit": execution["limit_hit"],
                    "process_residue_detected": execution["process_residue_detected"],
                    "processes_remaining": execution["processes_remaining"],
                },
                "discriminator_matched": discriminator_matched,
                "exit_code": execution["exit_code"],
                "stdout": execution["stdout"],
                "stderr": execution["stderr"],
                "duration_ms": execution["duration_ms"],
                "temporary_diff": {
                    "sha256": hashlib.sha256(diff_bytes).hexdigest(),
                    "summary": diff_stat,
                    "changed_paths": changed_paths,
                },
                "sandbox": execution["sandbox"],
                "sandbox_profile_sha256": execution["profile_sha256"],
                "limits": limits,
            },
            "lab_cleanup_verified": False,
            "blockers": blockers,
        }
    except (LabError, OSError, json.JSONDecodeError, subprocess.SubprocessError) as exc:
        evidence = blocked_envelope(base, head, packet_id, str(exc))
        if candidate_before is not None:
            try:
                after = candidate_identity(Path(args.repo_root).resolve(strict=True))
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
