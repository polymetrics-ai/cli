#!/usr/bin/env python3
"""Opt-in, out-of-band protection against feature-branch work staying local.

This is deliberately not a Git hook.  A scheduler invokes ``run`` at a bounded
cadence; the observer checks the *current* state of each linked worktree and
pushes only a pinned, non-default feature branch by an explicit non-force
refspec.  It never guesses how a past commit was made.
"""

from __future__ import annotations

import argparse
import fcntl
import hashlib
import json
import math
import os
import re
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterable


CONFIG_ENABLED = "unpushed-work-safety-net.enabled"
CONFIG_INTERVAL = "unpushed-work-safety-net.interval-seconds"
DEFAULT_INTERVAL_SECONDS = 600
LOCK_EXIT_CODE = 75
STATE_DIRECTORY = "unpushed-work-safety-net"
STATE_FILENAME = "state.json"
EVENTS_FILENAME = "events.jsonl"
LOCK_FILENAME = "observer.lock"
STATE_SCHEMA = 1
REMOTE_TIMEOUT_SECONDS = 30

# Every entry is resolved with ``git rev-parse --git-path`` inside the relevant
# worktree.  Linked worktrees therefore do not mistake another worktree's
# operation state for their own.
OPERATION_PATHS = (
    ("rebase", "rebase-apply"),
    ("rebase", "rebase-merge"),
    ("cherry_pick", "CHERRY_PICK_HEAD"),
    ("revert", "REVERT_HEAD"),
    ("sequencer", "sequencer"),
    ("bisect", "BISECT_LOG"),
    ("bisect", "BISECT_START"),
    ("merge", "MERGE_HEAD"),
    ("squash_merge", "SQUASH_MSG"),
    # MERGE_MSG is deliberately last: clean --no-commit cherry-picks and
    # reverts can leave only a message, while a real specific marker is a more
    # honest operation label when both are present.
    ("merge", "MERGE_MSG"),
)


class SafetyNetError(RuntimeError):
    """Expected observer error that must be reported without raw Git stderr."""


class StateError(SafetyNetError):
    """The durable state cannot be trusted for a safe scan."""


@dataclass(frozen=True)
class RepoContext:
    root: Path
    common_dir: Path

    @property
    def state_dir(self) -> Path:
        return self.common_dir / STATE_DIRECTORY


@dataclass(frozen=True)
class Worktree:
    path: Path
    branch: str | None


@dataclass(frozen=True)
class PushTarget:
    remote: str
    push_url: str
    default_branch: str


@dataclass(frozen=True)
class PushSnapshot:
    worktree: Worktree
    branch: str
    local_sha: str
    target: PushTarget
    remote_sha: str | None
    attempt_key: str


class EventReporter:
    """Print each outcome and append it to a durable, non-secret event log."""

    def __init__(self, state_dir: Path | None) -> None:
        self.state_dir = state_dir
        self.log_failed = False

    def emit(self, event: str, *, level: str = "info", **fields: object) -> None:
        record: dict[str, object] = {
            "event": event,
            "level": level,
            "at": int(time.time()),
        }
        record.update(
            {
                key: str(value) if isinstance(value, Path) else value
                for key, value in fields.items()
            }
        )
        # Values in this utility are ref names or fixed reasons. Push URLs,
        # filesystem paths, and Git stderr are intentionally never emitted:
        # any of them can contain a credential.
        rendered_fields = " ".join(
            f"{key}={self._render(value)}"
            for key, value in record.items()
            if key != "at"
        )
        print(f"unpushed-work-safety-net {rendered_fields}", flush=True)
        if self.state_dir is None:
            return
        try:
            self._append(record)
        except (OSError, TypeError):
            self.log_failed = True
            # stdout/stderr remains a visible failure channel even when the
            # repository's state volume is unavailable.
            print(
                "unpushed-work-safety-net event=event_log_write_failed level=error",
                file=sys.stderr,
                flush=True,
            )

    @staticmethod
    def _render(value: object) -> str:
        rendered: list[str] = []
        for character in str(value):
            if character == " ":
                rendered.append("_")
            elif ord(character) < 32 or ord(character) == 127:
                rendered.append(f"\\x{ord(character):02x}")
            else:
                rendered.append(character)
        return "".join(rendered)

    def _append(self, record: dict[str, object]) -> None:
        assert self.state_dir is not None
        path = self.state_dir / EVENTS_FILENAME
        with path.open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(record, sort_keys=True) + "\n")
            handle.flush()
            os.fsync(handle.fileno())


class AdvisoryLock:
    """A process-scoped lock that is released by the kernel if its owner dies."""

    def __init__(self, path: Path) -> None:
        self.path = path
        self.handle: Any | None = None

    def __enter__(self) -> "AdvisoryLock":
        try:
            self.handle = self.path.open("a+", encoding="utf-8")
        except OSError as exc:
            raise StateError("observer lock could not be opened") from exc
        try:
            fcntl.flock(self.handle.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            self.handle.close()
            self.handle = None
            raise
        return self

    def __exit__(self, exc_type: object, exc: object, traceback: object) -> None:
        if self.handle is not None:
            fcntl.flock(self.handle.fileno(), fcntl.LOCK_UN)
            self.handle.close()
            self.handle = None


def git(
    cwd: Path,
    args: Iterable[str],
    *,
    check: bool = True,
    timeout: int | None = None,
    capture_stderr: bool = True,
) -> subprocess.CompletedProcess[str]:
    """Run Git without shell interpolation or credential-bearing diagnostic output."""

    environment = {**os.environ, "GIT_TERMINAL_PROMPT": "0"}
    completed = subprocess.run(
        ["git", "-C", str(cwd), *args],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE if capture_stderr else subprocess.DEVNULL,
        env=environment,
        timeout=timeout,
    )
    if check and completed.returncode != 0:
        raise SafetyNetError(f"git command failed: {args!r}")
    return completed


def git_from_common_dir(
    common_dir: Path,
    args: Iterable[str],
    *,
    timeout: int | None = None,
) -> subprocess.CompletedProcess[str]:
    """Run a bare Git command against the shared object/config database.

    Pushing an already-pinned object from the common Git directory avoids
    operating through the mutable worktree whose state was inspected.
    """

    return subprocess.run(
        ["git", f"--git-dir={common_dir}", *args],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env={**os.environ, "GIT_TERMINAL_PROMPT": "0"},
        timeout=timeout,
    )


def git_config_one(cwd: Path, key: str) -> str | None:
    completed = git(cwd, ("config", "--get", key), check=False)
    if completed.returncode != 0:
        return None
    value = completed.stdout.strip()
    return value or None


def git_config_all(cwd: Path, key: str) -> list[str]:
    completed = git(cwd, ("config", "--get-all", key), check=False)
    if completed.returncode != 0:
        return []
    return [line for line in completed.stdout.splitlines() if line]


def repo_context(cwd: Path) -> RepoContext:
    root = Path(git(cwd, ("rev-parse", "--show-toplevel")).stdout.strip()).resolve()
    common_raw = git(root, ("rev-parse", "--git-common-dir")).stdout.strip()
    common_path = Path(common_raw)
    if not common_path.is_absolute():
        common_path = (root / common_path).resolve()
    return RepoContext(root=root, common_dir=common_path)


def worktrees(context: RepoContext) -> list[Worktree]:
    output = git(context.root, ("worktree", "list", "--porcelain")).stdout
    result: list[Worktree] = []
    current_path: Path | None = None
    current_branch: str | None = None

    def finish() -> None:
        nonlocal current_path, current_branch
        if current_path is not None:
            result.append(Worktree(path=current_path, branch=current_branch))
        current_path = None
        current_branch = None

    for line in output.splitlines():
        if not line:
            finish()
            continue
        if line.startswith("worktree "):
            current_path = Path(line.removeprefix("worktree ")).resolve()
        elif line.startswith("branch refs/heads/"):
            current_branch = line.removeprefix("branch refs/heads/")
        elif line == "detached":
            current_branch = None
    finish()
    return result


def operation_name(worktree: Worktree) -> str | None:
    for name, git_path in OPERATION_PATHS:
        resolved = git(
            worktree.path, ("rev-parse", "--git-path", git_path)
        ).stdout.strip()
        path = Path(resolved)
        if not path.is_absolute():
            path = worktree.path / path
        if path.exists():
            return name
    return None


def is_dirty(worktree: Worktree) -> bool:
    return bool(git(worktree.path, ("status", "--porcelain=v1", "-z")).stdout)


def local_sha(worktree: Worktree, branch: str) -> str | None:
    completed = git(
        worktree.path,
        ("rev-parse", "--verify", f"refs/heads/{branch}^{{commit}}"),
        check=False,
    )
    if completed.returncode != 0:
        return None
    return completed.stdout.strip() or None


def configured_push_remote(worktree: Worktree, branch: str) -> str | None:
    # This is Git's push target precedence: per-branch pushRemote, then the
    # repository's pushDefault, then the branch's tracking remote.  We fail
    # closed rather than guessing origin when no tracking relation exists.
    for key in (
        f"branch.{branch}.pushRemote",
        "remote.pushDefault",
        f"branch.{branch}.remote",
    ):
        value = git_config_one(worktree.path, key)
        if value:
            return value
    return None


def push_urls(worktree: Worktree, remote: str) -> list[str]:
    urls = git_config_all(worktree.path, f"remote.{remote}.pushurl")
    if urls:
        return urls
    return git_config_all(worktree.path, f"remote.{remote}.url")


def live_default_branch(push_url: str) -> str | None:
    try:
        completed = git(
            Path.cwd(),
            ("ls-remote", "--symref", "--", push_url, "HEAD"),
            check=False,
            timeout=REMOTE_TIMEOUT_SECONDS,
            capture_stderr=False,
        )
    except (OSError, subprocess.TimeoutExpired):
        return None
    if completed.returncode != 0:
        return None
    for line in completed.stdout.splitlines():
        match = re.fullmatch(r"ref:\s+(refs/heads/\S+)\s+HEAD", line)
        if match:
            return match.group(1).removeprefix("refs/heads/")
    return None


def live_remote_sha(push_url: str, branch: str) -> tuple[str | None, bool]:
    """Return (sha, transport_ok); a missing branch is a successful query."""

    try:
        completed = git(
            Path.cwd(),
            (
                "ls-remote",
                "--exit-code",
                "--",
                push_url,
                f"refs/heads/{branch}",
            ),
            check=False,
            timeout=REMOTE_TIMEOUT_SECONDS,
            capture_stderr=False,
        )
    except (OSError, subprocess.TimeoutExpired):
        return None, False
    if completed.returncode == 2:
        return None, True
    if completed.returncode != 0:
        return None, False
    for line in completed.stdout.splitlines():
        fields = line.split()
        if len(fields) == 2 and fields[1] == f"refs/heads/{branch}":
            return fields[0], True
    return None, False


def remote_is_ancestor(worktree: Worktree, remote_sha: str, branch_sha: str) -> bool:
    completed = git(
        worktree.path,
        ("merge-base", "--is-ancestor", remote_sha, branch_sha),
        check=False,
    )
    return completed.returncode == 0


def ensure_state_directory(context: RepoContext) -> None:
    context.state_dir.mkdir(mode=0o700, parents=True, exist_ok=True)


def new_state() -> dict[str, Any]:
    return {"schema": STATE_SCHEMA, "attempts": {}, "last_run": {}}


def load_state(context: RepoContext) -> dict[str, Any]:
    path = context.state_dir / STATE_FILENAME
    if not path.exists():
        return new_state()
    try:
        raw = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise StateError("state is invalid") from exc
    if not isinstance(raw, dict) or raw.get("schema") != STATE_SCHEMA:
        raise StateError("state has an unsupported schema")
    attempts = raw.get("attempts")
    last_run = raw.get("last_run")
    if not isinstance(attempts, dict) or not isinstance(last_run, dict):
        raise StateError("state has an invalid shape")
    if not all(
        isinstance(key, str) and isinstance(value, (int, float))
        for key, value in attempts.items()
    ):
        raise StateError("state has invalid attempt timestamps")
    return raw


def write_state(context: RepoContext, state: dict[str, Any]) -> None:
    """Replace state atomically; a failed write is visible and blocks a push."""

    ensure_state_directory(context)
    target = context.state_dir / STATE_FILENAME
    fd, temporary = tempfile.mkstemp(
        prefix="state.", suffix=".tmp", dir=context.state_dir
    )
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(state, handle, sort_keys=True)
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        os.chmod(temporary, 0o600)
        os.replace(temporary, target)
        directory_fd = os.open(context.state_dir, os.O_RDONLY)
        try:
            os.fsync(directory_fd)
        finally:
            os.close(directory_fd)
    except OSError as exc:
        try:
            os.unlink(temporary)
        except FileNotFoundError:
            pass
        raise StateError("state could not be written") from exc


def enabled(context: RepoContext) -> bool:
    return git_config_one(context.root, CONFIG_ENABLED) == "true"


def configured_interval(context: RepoContext) -> int:
    raw = git_config_one(context.root, CONFIG_INTERVAL)
    if raw is None:
        return DEFAULT_INTERVAL_SECONDS
    try:
        value = int(raw)
    except ValueError as exc:
        raise StateError("configured interval is invalid") from exc
    if value < DEFAULT_INTERVAL_SECONDS:
        raise StateError("configured interval is below the 600-second safety floor")
    return value


def configured_target(
    worktree: Worktree,
    branch: str,
    reporter: EventReporter,
) -> PushTarget | None:
    remote = configured_push_remote(worktree, branch)
    if not remote or remote == ".":
        reporter.emit("push_remote_unknown", level="error", branch=branch)
        return None
    known_remotes = set(git(worktree.path, ("remote",)).stdout.splitlines())
    if remote not in known_remotes:
        reporter.emit(
            "push_remote_unknown", level="error", branch=branch, remote=remote
        )
        return None
    urls = push_urls(worktree, remote)
    if len(urls) != 1:
        reporter.emit(
            "multiple_push_urls" if len(urls) > 1 else "push_url_unknown",
            level="error",
            branch=branch,
            remote=remote,
        )
        return None
    default_branch = live_default_branch(urls[0])
    if default_branch is None:
        reporter.emit(
            "default_branch_unknown", level="error", branch=branch, remote=remote
        )
        return None
    return PushTarget(remote=remote, push_url=urls[0], default_branch=default_branch)


def snapshot_for_push(
    worktree: Worktree,
    state: dict[str, Any],
    interval: int,
    reporter: EventReporter,
    now: float,
) -> tuple[PushSnapshot | None, bool]:
    """Inspect a worktree.  The bool is true only for a visible infrastructure error."""

    operation = operation_name(worktree)
    if operation is not None:
        reporter.emit(
            "operation_active",
            branch=worktree.branch or "detached",
            operation=operation,
        )
        return None, False
    if worktree.branch is None:
        reporter.emit("detached_head", level="attention")
        return None, False
    branch = worktree.branch
    if is_dirty(worktree):
        reporter.emit("dirty_worktree", level="attention", branch=branch)
        return None, False
    target = configured_target(worktree, branch, reporter)
    if target is None:
        return None, True
    if branch == target.default_branch:
        reporter.emit("default_branch", branch=branch, remote=target.remote)
        return None, False
    branch_sha = local_sha(worktree, branch)
    if branch_sha is None:
        reporter.emit("branch_tip_unknown", level="error", branch=branch)
        return None, True
    remote_sha, transport_ok = live_remote_sha(target.push_url, branch)
    if not transport_ok:
        reporter.emit(
            "remote_ref_unknown", level="error", branch=branch, remote=target.remote
        )
        return None, True
    if remote_sha == branch_sha:
        reporter.emit("synchronized", branch=branch, remote=target.remote)
        return None, False
    if remote_sha is not None and not remote_is_ancestor(
        worktree, remote_sha, branch_sha
    ):
        # If the remote tip is not locally present, ancestry is unknowable; it
        # is equally unsafe to push.  Report the same operator action: resolve
        # the divergent branch deliberately rather than letting a bot rewrite it.
        reporter.emit(
            "diverged", level="attention", branch=branch, remote=target.remote
        )
        return None, False
    attempt_key = f"{target.remote}\t{branch}"
    previous = state["attempts"].get(attempt_key)
    if isinstance(previous, (int, float)) and now - previous < interval:
        reporter.emit(
            "rate_limited",
            branch=branch,
            remote=target.remote,
            retry_in_seconds=math.ceil(interval - (now - previous)),
        )
        return None, False
    return (
        PushSnapshot(
            worktree=worktree,
            branch=branch,
            local_sha=branch_sha,
            target=target,
            remote_sha=remote_sha,
            attempt_key=attempt_key,
        ),
        False,
    )


def still_safe_to_push(
    snapshot: PushSnapshot, reporter: EventReporter
) -> tuple[bool, bool]:
    """Recheck mutable local and remote state immediately before the push."""

    operation = operation_name(snapshot.worktree)
    if operation is not None:
        reporter.emit("operation_active", branch=snapshot.branch, operation=operation)
        return False, False
    if is_dirty(snapshot.worktree):
        reporter.emit("dirty_worktree", level="attention", branch=snapshot.branch)
        return False, False
    current_sha = local_sha(snapshot.worktree, snapshot.branch)
    if current_sha != snapshot.local_sha:
        reporter.emit("branch_changed", branch=snapshot.branch)
        return False, False
    refreshed_target = configured_target(snapshot.worktree, snapshot.branch, reporter)
    if refreshed_target != snapshot.target:
        if refreshed_target is not None:
            reporter.emit(
                "push_target_changed",
                branch=snapshot.branch,
                remote=snapshot.target.remote,
            )
        return False, True
    if snapshot.branch == refreshed_target.default_branch:
        reporter.emit(
            "default_branch", branch=snapshot.branch, remote=refreshed_target.remote
        )
        return False, False
    remote_sha, transport_ok = live_remote_sha(
        refreshed_target.push_url, snapshot.branch
    )
    if not transport_ok:
        reporter.emit(
            "remote_ref_unknown",
            level="error",
            branch=snapshot.branch,
            remote=refreshed_target.remote,
        )
        return False, True
    if remote_sha != snapshot.remote_sha:
        reporter.emit(
            "remote_changed",
            level="attention",
            branch=snapshot.branch,
            remote=refreshed_target.remote,
        )
        return False, False
    return True, False


def push(context: RepoContext, snapshot: PushSnapshot, reporter: EventReporter) -> bool:
    """Push one exact SHA with no force option or force refspec."""

    refspec = f"{snapshot.local_sha}:refs/heads/{snapshot.branch}"
    try:
        completed = git_from_common_dir(
            context.common_dir,
            ("push", "--porcelain", "--", snapshot.target.push_url, refspec),
            timeout=REMOTE_TIMEOUT_SECONDS,
        )
    except (OSError, subprocess.TimeoutExpired):
        reporter.emit(
            "push_failed",
            level="error",
            branch=snapshot.branch,
            remote=snapshot.target.remote,
        )
        return False
    if completed.returncode != 0:
        diagnostic = f"{completed.stdout}\n{completed.stderr}".lower()
        rejected_markers = (
            "[rejected]",
            "non-fast-forward",
            "failed to push some refs",
            "remote rejected",
        )
        event = (
            "push_rejected"
            if any(marker in diagnostic for marker in rejected_markers)
            else "push_failed"
        )
        reporter.emit(
            event, level="error", branch=snapshot.branch, remote=snapshot.target.remote
        )
        return False
    reporter.emit(
        "pushed",
        branch=snapshot.branch,
        remote=snapshot.target.remote,
        sha=snapshot.local_sha,
    )
    return True


def run_observer(context: RepoContext) -> int:
    reporter = EventReporter(None)
    if not enabled(context):
        reporter.emit("not_enabled", level="attention")
        return 2
    try:
        ensure_state_directory(context)
        interval = configured_interval(context)
    except (OSError, StateError):
        reporter.emit("configuration_or_state_failed", level="error")
        return 1
    reporter = EventReporter(context.state_dir)
    try:
        with AdvisoryLock(context.state_dir / LOCK_FILENAME):
            state = new_state()
            try:
                state = load_state(context)
                state["last_run"] = {
                    "started_at": int(time.time()),
                    "result": "running",
                }
                write_state(context, state)
            except StateError:
                reporter.emit("state_invalid", level="error")
                return 1

            reporter.emit("scan_started", interval_seconds=interval)
            had_error = reporter.log_failed
            for worktree in worktrees(context):
                snapshot, inspection_error = snapshot_for_push(
                    worktree,
                    state,
                    interval,
                    reporter,
                    time.time(),
                )
                had_error = had_error or inspection_error or reporter.log_failed
                if snapshot is None:
                    continue
                safe_to_push, recheck_error = still_safe_to_push(snapshot, reporter)
                had_error = had_error or recheck_error or reporter.log_failed
                if not safe_to_push:
                    continue
                # Persist before the external side effect. A crash cannot cause
                # a rapid retry storm; an unwritable state is a visible stop.
                state["attempts"][snapshot.attempt_key] = time.time()
                try:
                    write_state(context, state)
                except StateError:
                    reporter.emit(
                        "state_write_failed", level="error", branch=snapshot.branch
                    )
                    had_error = True
                    continue
                if not push(context, snapshot, reporter):
                    had_error = True
                had_error = had_error or reporter.log_failed

            state["last_run"] = {
                "finished_at": int(time.time()),
                "result": "error" if had_error else "ok",
            }
            try:
                write_state(context, state)
            except StateError:
                reporter.emit("state_write_failed", level="error")
                return 1
            reporter.emit(
                "scan_finished",
                level="error" if had_error else "info",
                result=state["last_run"]["result"],
            )
            return 1 if had_error or reporter.log_failed else 0
    except BlockingIOError:
        reporter.emit("observer_busy", level="attention")
        return LOCK_EXIT_CODE
    except (OSError, SafetyNetError, subprocess.TimeoutExpired):
        reporter.emit("scan_failed", level="error")
        return 1


def set_enabled(context: RepoContext, value: bool, interval: int | None = None) -> int:
    if value:
        selected_interval = (
            interval if interval is not None else DEFAULT_INTERVAL_SECONDS
        )
        if selected_interval < DEFAULT_INTERVAL_SECONDS:
            print(
                "unpushed-work-safety-net event=interval_below_floor level=error",
                file=sys.stderr,
                flush=True,
            )
            return 2
        git(context.root, ("config", "--local", CONFIG_ENABLED, "true"))
        git(
            context.root, ("config", "--local", CONFIG_INTERVAL, str(selected_interval))
        )
        try:
            ensure_state_directory(context)
        except OSError:
            print(
                "unpushed-work-safety-net event=state_write_failed level=error",
                file=sys.stderr,
                flush=True,
            )
            return 1
        reporter = EventReporter(context.state_dir)
        reporter.emit("enabled", interval_seconds=selected_interval)
        return 1 if reporter.log_failed else 0
    git(context.root, ("config", "--local", CONFIG_ENABLED, "false"))
    reporter = EventReporter(context.state_dir if context.state_dir.exists() else None)
    reporter.emit("disabled")
    return 1 if reporter.log_failed else 0


def status(context: RepoContext) -> int:
    reporter = EventReporter(context.state_dir if context.state_dir.exists() else None)
    if not enabled(context):
        reporter.emit("not_enabled", level="attention")
        return 2
    try:
        interval = configured_interval(context)
        state = load_state(context)
    except StateError:
        reporter.emit("state_invalid", level="error")
        return 1
    last_run = state.get("last_run", {})
    finished_at = last_run.get("finished_at") if isinstance(last_run, dict) else None
    if not isinstance(finished_at, int):
        reporter.emit(
            "status_no_successful_run", level="attention", interval_seconds=interval
        )
        return 1
    age = max(0, int(time.time()) - finished_at)
    if age > interval * 2 + 60:
        reporter.emit(
            "stale_heartbeat", level="error", age_seconds=age, interval_seconds=interval
        )
        return 1
    if last_run.get("result") != "ok":
        reporter.emit(
            "last_run_failed",
            level="error",
            age_seconds=age,
            interval_seconds=interval,
        )
        return 1
    reporter.emit(
        "status_healthy",
        age_seconds=age,
        interval_seconds=interval,
        last_result=last_run.get("result", "unknown"),
    )
    return 0 if not reporter.log_failed else 1


def launchd_label(context: RepoContext) -> str:
    digest = hashlib.sha256(str(context.common_dir).encode("utf-8")).hexdigest()[:12]
    return f"ai.polymetrics.unpushed-work-safety-net.{digest}"


def plist_escape(value: str) -> str:
    return value.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def launchd_plist(context: RepoContext) -> int:
    try:
        interval = configured_interval(context)
    except StateError:
        EventReporter(context.state_dir if context.state_dir.exists() else None).emit(
            "state_invalid", level="error"
        )
        return 1
    script = Path(__file__).resolve()
    stdout_log = context.state_dir / "launchd.stdout.log"
    stderr_log = context.state_dir / "launchd.stderr.log"
    arguments = [sys.executable, str(script), "run"]
    argument_xml = "\n".join(
        f"      <string>{plist_escape(argument)}</string>" for argument in arguments
    )
    print(
        "\n".join(
            (
                '<?xml version="1.0" encoding="UTF-8"?>',
                '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" '
                '"http://www.apple.com/DTDs/PropertyList-1.0.dtd">',
                '<plist version="1.0">',
                "<dict>",
                "  <key>Label</key>",
                f"  <string>{launchd_label(context)}</string>",
                "  <key>ProgramArguments</key>",
                "  <array>",
                argument_xml,
                "  </array>",
                "  <key>StartInterval</key>",
                f"  <integer>{interval}</integer>",
                "  <key>RunAtLoad</key>",
                "  <false/>",
                "  <key>StandardOutPath</key>",
                f"  <string>{plist_escape(str(stdout_log))}</string>",
                "  <key>StandardErrorPath</key>",
                f"  <string>{plist_escape(str(stderr_log))}</string>",
                "</dict>",
                "</plist>",
            )
        )
    )
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    subcommands = parser.add_subparsers(dest="command", required=True)
    enable_parser = subcommands.add_parser(
        "enable", help="explicitly enable the local observer"
    )
    enable_parser.add_argument(
        "--interval-seconds", type=int, default=DEFAULT_INTERVAL_SECONDS
    )
    subcommands.add_parser(
        "disable", help="disable observer scans without touching Git hooks"
    )
    subcommands.add_parser("run", help="perform one scheduled observer scan")
    subcommands.add_parser("status", help="report durable observer health")
    subcommands.add_parser(
        "launchd-label", help="print this repository's unique launchd label"
    )
    subcommands.add_parser(
        "launchd-plist", help="print an explicit 600-second launchd plist"
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        context = repo_context(Path.cwd())
        if args.command == "enable":
            return set_enabled(context, True, args.interval_seconds)
        if args.command == "disable":
            return set_enabled(context, False)
        if args.command == "run":
            return run_observer(context)
        if args.command == "status":
            return status(context)
        if args.command == "launchd-label":
            print(launchd_label(context))
            return 0
        if args.command == "launchd-plist":
            return launchd_plist(context)
    except SafetyNetError:
        print(
            "unpushed-work-safety-net event=repository_error level=error",
            file=sys.stderr,
            flush=True,
        )
        return 1
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
