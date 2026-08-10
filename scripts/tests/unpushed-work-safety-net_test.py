#!/usr/bin/env python3
"""End-to-end contracts for the opt-in unpushed-work observer.

Every Git state below is created by Git itself. In particular, the rebase test
leaves a genuine multi-commit rebase paused on a conflict; it never fakes a
marker file.
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
OBSERVER = REPO_ROOT / "scripts" / "unpushed-work-safety-net.py"
STATE_DIR_NAME = "unpushed-work-safety-net"


class GitFixture(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp_dir = Path(tempfile.mkdtemp(prefix="unpushed-work-safety-net."))
        self.origin = self.create_bare("origin", "main")
        self.primary = self.tmp_dir / "primary"
        self.git(self.tmp_dir, "clone", str(self.origin), str(self.primary))
        self.configure_identity(self.primary)
        self.git(self.primary, "switch", "-c", "main")
        self.write(self.primary, "conflict.txt", "base\n")
        self.git(self.primary, "add", "conflict.txt")
        self.git(self.primary, "commit", "-m", "base")
        self.git(self.primary, "push", "-u", "origin", "main")

    def tearDown(self) -> None:
        shutil.rmtree(self.tmp_dir, ignore_errors=True)

    def git(
        self, cwd: Path, *args: str, check: bool = True
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["git", *args],
            cwd=cwd,
            check=check,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env={
                **os.environ,
                "GIT_AUTHOR_NAME": "Safety Net Test",
                "GIT_AUTHOR_EMAIL": "safety-net@example.test",
                "GIT_COMMITTER_NAME": "Safety Net Test",
                "GIT_COMMITTER_EMAIL": "safety-net@example.test",
            },
        )

    def create_bare(self, name: str, default_branch: str) -> Path:
        path = self.tmp_dir / f"{name}.git"
        self.git(self.tmp_dir, "init", "--bare", str(path))
        self.git(path, "symbolic-ref", "HEAD", f"refs/heads/{default_branch}")
        return path

    def configure_identity(self, worktree: Path) -> None:
        self.git(worktree, "config", "user.name", "Safety Net Test")
        self.git(worktree, "config", "user.email", "safety-net@example.test")

    def write(self, worktree: Path, filename: str, content: str) -> None:
        (worktree / filename).write_text(content, encoding="utf-8")

    def commit(self, worktree: Path, message: str, *paths: str) -> str:
        self.git(worktree, "add", *paths)
        self.git(worktree, "commit", "-m", message)
        return self.git(worktree, "rev-parse", "HEAD").stdout.strip()

    def feature_worktree(self, branch: str) -> Path:
        path = self.tmp_dir / branch.replace("/", "-")
        self.git(
            self.primary, "worktree", "add", "-b", branch, str(path), "origin/main"
        )
        self.configure_identity(path)
        return path

    def observer(
        self,
        *args: str,
        check: bool = False,
        environment: dict[str, str] | None = None,
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [sys.executable, str(OBSERVER), *args],
            cwd=self.primary,
            check=check,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env={**os.environ, **(environment or {})},
        )

    def enable(self) -> None:
        completed = self.observer("enable")
        self.assertEqual(completed.returncode, 0, completed.stderr + completed.stdout)

    def remote_sha(self, remote: Path, branch: str) -> str | None:
        completed = self.git(
            remote, "rev-parse", "--verify", f"refs/heads/{branch}", check=False
        )
        return completed.stdout.strip() if completed.returncode == 0 else None

    def common_dir(self) -> Path:
        raw = self.git(self.primary, "rev-parse", "--git-common-dir").stdout.strip()
        path = Path(raw)
        return path if path.is_absolute() else (self.primary / path).resolve()

    def state_dir(self) -> Path:
        return self.common_dir() / STATE_DIR_NAME

    def assert_event(
        self, completed: subprocess.CompletedProcess[str], event: str
    ) -> None:
        self.assertIn(
            f"event={event}", completed.stdout, completed.stdout + completed.stderr
        )


class UnpushedWorkSafetyNetTest(GitFixture):
    def test_opt_in_opt_out_and_dirty_recovery(self) -> None:
        feature = self.feature_worktree("feature/dirty")
        self.write(feature, "feature.txt", "committed\n")
        self.commit(feature, "feature commit", "feature.txt")

        disabled = self.observer("run")
        self.assertNotEqual(disabled.returncode, 0)
        self.assert_event(disabled, "not_enabled")
        self.assertIsNone(self.remote_sha(self.origin, "feature/dirty"))

        self.enable()
        self.write(feature, "feature.txt", "uncommitted\n")
        dirty = self.observer("run")
        self.assertEqual(dirty.returncode, 0, dirty.stderr)
        self.assert_event(dirty, "dirty_worktree")
        self.assertIsNone(self.remote_sha(self.origin, "feature/dirty"))

        local_sha = self.commit(feature, "finish dirty work", "feature.txt")
        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(self.origin, "feature/dirty"), local_sha)

        disabled_again = self.observer("disable")
        self.assertEqual(disabled_again.returncode, 0, disabled_again.stderr)
        self.write(feature, "after-disable.txt", "local only\n")
        self.commit(feature, "after disable", "after-disable.txt")
        skipped = self.observer("run")
        self.assertNotEqual(skipped.returncode, 0)
        self.assert_event(skipped, "not_enabled")
        self.assertEqual(self.remote_sha(self.origin, "feature/dirty"), local_sha)

    def test_default_branch_is_never_pushed(self) -> None:
        self.enable()
        self.write(self.primary, "main-only.txt", "must remain local\n")
        local_main = self.commit(self.primary, "local main commit", "main-only.txt")
        before = self.remote_sha(self.origin, "main")

        completed = self.observer("run")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assert_event(completed, "default_branch")
        self.assertNotEqual(local_main, before)
        self.assertEqual(self.remote_sha(self.origin, "main"), before)

    def test_real_multi_commit_rebase_is_deferred_then_recovers_without_blocking_other_worktree(
        self,
    ) -> None:
        rebasing = self.feature_worktree("feature/rebasing")
        independent = self.feature_worktree("feature/independent")
        self.write(rebasing, "conflict.txt", "feature first\n")
        self.commit(rebasing, "feature first", "conflict.txt")
        self.write(rebasing, "second.txt", "second commit\n")
        self.commit(rebasing, "feature second", "second.txt")
        self.write(independent, "independent.txt", "independent\n")
        independent_sha = self.commit(
            independent, "independent commit", "independent.txt"
        )

        self.write(self.primary, "conflict.txt", "main change\n")
        self.commit(self.primary, "main change", "conflict.txt")
        self.git(self.primary, "push", "origin", "main")

        paused = self.git(rebasing, "rebase", "origin/main", check=False)
        self.assertNotEqual(paused.returncode, 0, paused.stdout + paused.stderr)
        self.assertTrue((rebasing / ".git").exists())

        self.enable()
        deferred = self.observer("run")
        self.assertEqual(deferred.returncode, 0, deferred.stderr)
        self.assert_event(deferred, "operation_active")
        self.assert_event(deferred, "pushed")
        self.assertIsNone(self.remote_sha(self.origin, "feature/rebasing"))
        self.assertEqual(
            self.remote_sha(self.origin, "feature/independent"), independent_sha
        )

        self.write(rebasing, "conflict.txt", "resolved\n")
        self.git(rebasing, "add", "conflict.txt")
        continued = self.git(
            rebasing, "-c", "core.editor=true", "rebase", "--continue", check=False
        )
        self.assertEqual(continued.returncode, 0, continued.stdout + continued.stderr)
        rebased_sha = self.git(rebasing, "rev-parse", "HEAD").stdout.strip()

        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(self.origin, "feature/rebasing"), rebased_sha)

    def test_real_merge_cherry_pick_revert_bisect_and_squash_states_are_deferred_then_recover(
        self,
    ) -> None:
        self.enable()

        merge = self.feature_worktree("feature/merge")
        self.write(merge, "conflict.txt", "merge feature\n")
        merge_sha = self.commit(merge, "merge feature", "conflict.txt")
        self.write(self.primary, "conflict.txt", "merge main\n")
        self.commit(self.primary, "merge main", "conflict.txt")
        self.git(self.primary, "push", "origin", "main")
        self.assertNotEqual(
            self.git(merge, "merge", "origin/main", check=False).returncode, 0
        )
        deferred = self.observer("run")
        self.assert_event(deferred, "operation_active")
        self.assertIn("operation=merge", deferred.stdout)
        self.assertIsNone(self.remote_sha(self.origin, "feature/merge"))
        self.git(merge, "merge", "--abort")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(self.remote_sha(self.origin, "feature/merge"), merge_sha)

        cherry_source = self.feature_worktree("feature/cherry-source")
        self.write(cherry_source, "conflict.txt", "cherry source\n")
        cherry_sha = self.commit(cherry_source, "cherry source", "conflict.txt")
        cherry_target = self.feature_worktree("feature/cherry-target")
        self.write(cherry_target, "conflict.txt", "cherry target\n")
        cherry_target_sha = self.commit(cherry_target, "cherry target", "conflict.txt")
        self.assertNotEqual(
            self.git(cherry_target, "cherry-pick", cherry_sha, check=False).returncode,
            0,
        )
        deferred = self.observer("run")
        self.assert_event(deferred, "operation_active")
        self.assertIn("operation=cherry_pick", deferred.stdout)
        self.assertIsNone(self.remote_sha(self.origin, "feature/cherry-target"))
        self.git(cherry_target, "cherry-pick", "--abort")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(
            self.remote_sha(self.origin, "feature/cherry-target"), cherry_target_sha
        )

        reverting = self.feature_worktree("feature/revert")
        self.write(reverting, "revert.txt", "first\n")
        first_sha = self.commit(reverting, "revert first", "revert.txt")
        self.write(reverting, "revert.txt", "second\n")
        revert_tip = self.commit(reverting, "revert second", "revert.txt")
        self.assertNotEqual(
            self.git(
                reverting, "revert", "--no-edit", first_sha, check=False
            ).returncode,
            0,
        )
        deferred = self.observer("run")
        self.assert_event(deferred, "operation_active")
        self.assertIn("operation=revert", deferred.stdout)
        self.assertIsNone(self.remote_sha(self.origin, "feature/revert"))
        self.git(reverting, "revert", "--abort")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(self.remote_sha(self.origin, "feature/revert"), revert_tip)

        bisecting = self.feature_worktree("feature/bisect")
        for index in range(3):
            filename = f"bisect-{index}.txt"
            self.write(bisecting, filename, f"{index}\n")
            self.commit(bisecting, f"bisect {index}", filename)
        bisect_tip = self.git(bisecting, "rev-parse", "HEAD").stdout.strip()
        self.assertEqual(
            self.git(bisecting, "bisect", "start", "HEAD", "HEAD~2").returncode, 0
        )
        deferred = self.observer("run")
        self.assert_event(deferred, "operation_active")
        self.assertIn("operation=bisect", deferred.stdout)
        self.assertIsNone(self.remote_sha(self.origin, "feature/bisect"))
        self.git(bisecting, "bisect", "reset")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(self.remote_sha(self.origin, "feature/bisect"), bisect_tip)

        squash_source = self.feature_worktree("feature/squash-source")
        self.write(squash_source, "squash-source.txt", "source\n")
        squash_source_sha = self.commit(
            squash_source, "squash source", "squash-source.txt"
        )
        squash_target = self.feature_worktree("feature/squash-target")
        self.write(squash_target, "squash-target.txt", "target\n")
        squash_target_sha = self.commit(
            squash_target, "squash target", "squash-target.txt"
        )
        self.assertEqual(
            self.git(squash_target, "merge", "--squash", squash_source_sha).returncode,
            0,
        )
        deferred = self.observer("run")
        self.assert_event(deferred, "operation_active")
        self.assertIn("operation=squash_merge", deferred.stdout)
        self.assertIsNone(self.remote_sha(self.origin, "feature/squash-target"))
        self.git(squash_target, "reset", "--merge")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(
            self.remote_sha(self.origin, "feature/squash-target"), squash_target_sha
        )

    def test_effective_pushurl_default_branch_wins_over_fetch_remote(self) -> None:
        fetch_remote = self.create_bare("push-target-fetch", "main")
        push_remote = self.create_bare("push-target-push", "feature/target")
        initial = self.remote_sha(self.origin, "main")
        self.git(
            fetch_remote, "fetch", str(self.origin), "refs/heads/main:refs/heads/main"
        )
        self.git(
            push_remote, "fetch", str(self.origin), "refs/heads/main:refs/heads/main"
        )
        self.git(push_remote, "update-ref", "refs/heads/feature/target", initial)

        feature = self.feature_worktree("feature/target")
        self.write(feature, "target.txt", "local target\n")
        local_sha = self.commit(feature, "target commit", "target.txt")
        self.git(feature, "remote", "add", "safety-target", str(fetch_remote))
        self.git(
            feature, "remote", "set-url", "--push", "safety-target", str(push_remote)
        )
        self.git(feature, "config", "branch.feature/target.pushRemote", "safety-target")
        self.enable()

        completed = self.observer("run")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assert_event(completed, "default_branch")
        self.assertNotEqual(local_sha, initial)
        self.assertEqual(self.remote_sha(push_remote, "feature/target"), initial)

        self.git(push_remote, "symbolic-ref", "HEAD", "refs/heads/main")
        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(push_remote, "feature/target"), local_sha)
        self.assertIsNone(self.remote_sha(fetch_remote, "feature/target"))

    def test_divergence_never_rewrites_remote_work(self) -> None:
        feature = self.feature_worktree("feature/diverged")
        self.write(feature, "feature.txt", "local base\n")
        self.commit(feature, "local base", "feature.txt")
        self.git(feature, "push", "-u", "origin", "feature/diverged")

        other = self.tmp_dir / "other"
        self.git(self.tmp_dir, "clone", str(self.origin), str(other))
        self.configure_identity(other)
        self.git(other, "switch", "feature/diverged")
        self.write(other, "remote.txt", "remote wins\n")
        remote_tip = self.commit(other, "remote advancement", "remote.txt")
        self.git(other, "push", "origin", "feature/diverged")

        self.write(feature, "local.txt", "local divergent\n")
        self.commit(feature, "local divergence", "local.txt")
        self.enable()
        completed = self.observer("run")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assert_event(completed, "diverged")
        self.assertEqual(self.remote_sha(self.origin, "feature/diverged"), remote_tip)

    def test_detached_and_multiple_push_urls_are_deferred_then_recover(self) -> None:
        detached = self.feature_worktree("feature/detached")
        self.write(detached, "detached.txt", "detached commit\n")
        detached_sha = self.commit(detached, "detached commit", "detached.txt")
        self.git(detached, "switch", "--detach")
        self.enable()

        deferred = self.observer("run")
        self.assertEqual(deferred.returncode, 0, deferred.stderr)
        self.assert_event(deferred, "detached_head")
        self.assertIsNone(self.remote_sha(self.origin, "feature/detached"))

        self.git(detached, "switch", "feature/detached")
        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(self.origin, "feature/detached"), detached_sha)

        ambiguous = self.feature_worktree("feature/ambiguous-pushurl")
        self.write(ambiguous, "ambiguous.txt", "local\n")
        ambiguous_sha = self.commit(ambiguous, "ambiguous commit", "ambiguous.txt")
        second_push_url = self.create_bare("second-push-url", "main")
        self.git(
            ambiguous,
            "remote",
            "set-url",
            "--add",
            "--push",
            "origin",
            str(self.origin),
        )
        self.git(
            ambiguous,
            "remote",
            "set-url",
            "--add",
            "--push",
            "origin",
            str(second_push_url),
        )

        blocked = self.observer("run")
        self.assertNotEqual(blocked.returncode, 0)
        self.assert_event(blocked, "multiple_push_urls")
        self.assertIsNone(self.remote_sha(self.origin, "feature/ambiguous-pushurl"))

        self.git(ambiguous, "config", "--unset-all", "remote.origin.pushurl")
        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(
            self.remote_sha(self.origin, "feature/ambiguous-pushurl"), ambiguous_sha
        )

    def test_remote_change_during_final_recheck_is_deferred_then_recovers(self) -> None:
        branch = "feature/remote-recheck"
        feature = self.feature_worktree(branch)
        self.write(feature, "feature.txt", "local candidate\n")
        local_sha = self.commit(feature, "local candidate", "feature.txt")
        base_sha = self.remote_sha(self.origin, "main")
        assert base_sha is not None
        self.enable()

        wrapper_dir = self.tmp_dir / "remote-change-wrapper"
        wrapper_dir.mkdir()
        fired = self.tmp_dir / "remote-change-fired"
        real_git = shutil.which("git")
        self.assertIsNotNone(real_git)
        wrapper = wrapper_dir / "git"
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env python3
                import os
                import subprocess
                import sys
                from pathlib import Path

                real_git = {real_git!r}
                branch_ref = "refs/heads/{branch}"
                fired = Path({str(fired)!r})
                arguments = sys.argv[1:]
                completed = subprocess.run([real_git, *arguments])
                if (
                    "ls-remote" in arguments
                    and branch_ref in arguments
                    and not fired.exists()
                ):
                    fired.write_text("fired\\n", encoding="utf-8")
                    subprocess.run(
                        [
                            real_git,
                            "--git-dir={self.origin}",
                            "update-ref",
                            branch_ref,
                            {base_sha!r},
                        ],
                        check=True,
                    )
                raise SystemExit(completed.returncode)
                """
            ),
            encoding="utf-8",
        )
        wrapper.chmod(0o755)

        deferred = self.observer(
            "run",
            environment={"PATH": f"{wrapper_dir}{os.pathsep}{os.environ['PATH']}"},
        )

        self.assertEqual(deferred.returncode, 0, deferred.stderr + deferred.stdout)
        self.assert_event(deferred, "remote_changed")
        self.assertEqual(self.remote_sha(self.origin, branch), base_sha)
        self.assertTrue(fired.exists())

        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr + recovered.stdout)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(self.origin, branch), local_sha)

    def test_actual_non_fast_forward_rejection_never_rewrites_remote(self) -> None:
        feature = self.feature_worktree("feature/rejected")
        initial = self.remote_sha(self.origin, "main")
        self.git(self.origin, "update-ref", "refs/heads/feature/rejected", initial)
        self.write(feature, "local.txt", "local candidate\n")
        self.commit(feature, "local candidate", "local.txt")

        other = self.tmp_dir / "racing-remote-writer"
        self.git(self.tmp_dir, "clone", str(self.origin), str(other))
        self.configure_identity(other)
        self.git(other, "switch", "feature/rejected")
        hook_dir = self.tmp_dir / "pre-push-hooks"
        hook_dir.mkdir()
        hook = hook_dir / "pre-push"
        hook.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env python3
                import subprocess
                from pathlib import Path

                worktree = Path({str(other)!r})
                path = worktree / "remote-race.txt"
                path.write_text("remote wins\\n", encoding="utf-8")
                subprocess.run(["git", "-C", str(worktree), "add", "remote-race.txt"], check=True)
                subprocess.run(["git", "-C", str(worktree), "commit", "-m", "remote wins"], check=True)
                subprocess.run(["git", "-C", str(worktree), "push", "origin", "feature/rejected"], check=True)
                """
            ),
            encoding="utf-8",
        )
        hook.chmod(0o755)
        self.git(feature, "config", "core.hooksPath", str(hook_dir))
        self.enable()

        rejected = self.observer("run")

        self.assertNotEqual(rejected.returncode, 0)
        self.assert_event(rejected, "push_rejected")
        remote_tip = self.git(other, "rev-parse", "HEAD").stdout.strip()
        self.assertEqual(self.remote_sha(self.origin, "feature/rejected"), remote_tip)

    def test_every_observer_git_invocation_rejects_force_arguments(self) -> None:
        feature = self.feature_worktree("feature/no-force")
        self.write(feature, "feature.txt", "local\n")
        local_sha = self.commit(feature, "feature commit", "feature.txt")
        self.enable()
        wrapper_dir = self.tmp_dir / "git-wrapper"
        wrapper_dir.mkdir()
        real_git = shutil.which("git")
        self.assertIsNotNone(real_git)
        wrapper = wrapper_dir / "git"
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env python3
                import os
                import sys

                for argument in sys.argv[1:]:
                    if argument == "--force" or argument.startswith("--force=") or argument.startswith("+"):
                        raise SystemExit("force push argument refused by test wrapper")
                os.execv({real_git!r}, [{real_git!r}, *sys.argv[1:]])
                """
            ),
            encoding="utf-8",
        )
        wrapper.chmod(0o755)

        completed = self.observer(
            "run",
            environment={"PATH": f"{wrapper_dir}{os.pathsep}{os.environ['PATH']}"},
        )

        self.assertEqual(completed.returncode, 0, completed.stderr + completed.stdout)
        self.assert_event(completed, "pushed")
        self.assertEqual(self.remote_sha(self.origin, "feature/no-force"), local_sha)

    def test_missing_remote_default_and_corrupt_state_report_then_recover(self) -> None:
        feature = self.feature_worktree("feature/unknown-default")
        self.write(feature, "feature.txt", "local\n")
        local_sha = self.commit(feature, "feature commit", "feature.txt")
        self.enable()
        self.git(self.origin, "symbolic-ref", "HEAD", "refs/heads/does-not-exist")

        unknown = self.observer("run")
        self.assertNotEqual(unknown.returncode, 0)
        self.assert_event(unknown, "default_branch_unknown")
        self.assertIsNone(self.remote_sha(self.origin, "feature/unknown-default"))
        failed_status = self.observer("status")
        self.assertNotEqual(failed_status.returncode, 0)
        self.assert_event(failed_status, "last_run_failed")

        self.git(self.origin, "symbolic-ref", "HEAD", "refs/heads/main")
        state_path = self.state_dir() / "state.json"
        state_path.write_text("not json", encoding="utf-8")
        corrupt = self.observer("run")
        self.assertNotEqual(corrupt.returncode, 0)
        self.assert_event(corrupt, "state_invalid")
        self.assertIsNone(self.remote_sha(self.origin, "feature/unknown-default"))

        state_path.unlink()
        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(
            self.remote_sha(self.origin, "feature/unknown-default"), local_sha
        )

    def test_kernel_lock_contention_is_visible_and_releases_after_holder_exits(
        self,
    ) -> None:
        feature = self.feature_worktree("feature/locked")
        self.write(feature, "feature.txt", "local\n")
        local_sha = self.commit(feature, "feature commit", "feature.txt")
        self.enable()
        self.state_dir().mkdir(parents=True, exist_ok=True)
        lock_path = self.state_dir() / "observer.lock"
        helper = subprocess.Popen(
            [
                sys.executable,
                "-c",
                textwrap.dedent(
                    """
                    import fcntl
                    import sys
                    import time
                    with open(sys.argv[1], "a+", encoding="utf-8") as handle:
                        fcntl.flock(handle.fileno(), fcntl.LOCK_EX)
                        print("locked", flush=True)
                        time.sleep(30)
                    """
                ),
                str(lock_path),
            ],
            stdout=subprocess.PIPE,
            text=True,
        )
        self.assertEqual(helper.stdout.readline().strip(), "locked")
        try:
            busy = self.observer("run")
            self.assertEqual(busy.returncode, 75, busy.stderr)
            self.assert_event(busy, "observer_busy")
            self.assertIsNone(self.remote_sha(self.origin, "feature/locked"))
        finally:
            helper.terminate()
            helper.wait(timeout=10)
            assert helper.stdout is not None
            helper.stdout.close()

        recovered = self.observer("run")
        self.assertEqual(recovered.returncode, 0, recovered.stderr)
        self.assert_event(recovered, "pushed")
        self.assertEqual(self.remote_sha(self.origin, "feature/locked"), local_sha)

    def test_rate_floor_reports_instead_of_triggering_a_second_push(self) -> None:
        feature = self.feature_worktree("feature/rate-limited")
        self.write(feature, "first.txt", "first\n")
        first_sha = self.commit(feature, "first", "first.txt")
        self.enable()
        first = self.observer("run")
        self.assertEqual(first.returncode, 0, first.stderr)
        self.assert_event(first, "pushed")
        self.assertEqual(
            self.remote_sha(self.origin, "feature/rate-limited"), first_sha
        )

        self.write(feature, "second.txt", "second\n")
        second_sha = self.commit(feature, "second", "second.txt")
        limited = self.observer("run")
        self.assertEqual(limited.returncode, 0, limited.stderr)
        self.assert_event(limited, "rate_limited")
        self.assertNotEqual(second_sha, first_sha)
        self.assertEqual(
            self.remote_sha(self.origin, "feature/rate-limited"), first_sha
        )

    def test_status_exposes_missing_and_healthy_heartbeats_and_plist_keeps_the_floor(
        self,
    ) -> None:
        disabled = self.observer("status")
        self.assertNotEqual(disabled.returncode, 0)
        self.assert_event(disabled, "not_enabled")

        below_floor = self.observer("enable", "--interval-seconds", "599")
        self.assertNotEqual(below_floor.returncode, 0)
        self.assertIn("event=interval_below_floor", below_floor.stderr)

        self.enable()
        never_run = self.observer("status")
        self.assertNotEqual(never_run.returncode, 0)
        self.assert_event(never_run, "status_no_successful_run")

        completed = self.observer("run")
        self.assertEqual(completed.returncode, 0, completed.stderr)
        healthy = self.observer("status")
        self.assertEqual(healthy.returncode, 0, healthy.stderr)
        self.assert_event(healthy, "status_healthy")

        state_path = self.state_dir() / "state.json"
        state = json.loads(state_path.read_text(encoding="utf-8"))
        state["last_run"]["finished_at"] = 0
        state_path.write_text(json.dumps(state), encoding="utf-8")
        stale = self.observer("status")
        self.assertNotEqual(stale.returncode, 0)
        self.assert_event(stale, "stale_heartbeat")
        self.assertEqual(self.observer("run").returncode, 0)
        self.assertEqual(self.observer("status").returncode, 0)

        label = self.observer("launchd-label")
        self.assertEqual(label.returncode, 0, label.stderr)
        self.assertTrue(
            label.stdout.strip().startswith("ai.polymetrics.unpushed-work-safety-net.")
        )
        plist = self.observer("launchd-plist")
        self.assertEqual(plist.returncode, 0, plist.stderr)
        self.assertIn("<integer>600</integer>", plist.stdout)
        self.assertIn("<false/>", plist.stdout)


if __name__ == "__main__":
    unittest.main(verbosity=2)
