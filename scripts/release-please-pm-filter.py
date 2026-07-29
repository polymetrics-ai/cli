#!/usr/bin/env python3
import argparse
import json
import re
import subprocess
from pathlib import Path

WEBSITE_PREFIXES = ("website/",)
WEBSITE_EXACT_PATHS = frozenset((".github/workflows/website.yml", ".gitlab-ci.yml"))
BUMP_ORDER = {"patch": 1, "minor": 2, "major": 3}


def run_git(args, cwd):
    return subprocess.check_output(["git", *args], cwd=cwd, text=True).strip()


def git_ref_exists(ref, cwd):
    try:
        run_git(["rev-parse", "--verify", f"{ref}^{{commit}}"], cwd)
    except subprocess.CalledProcessError:
        return False
    return True


def default_base_ref(cwd):
    manifest = json.loads((Path(cwd) / ".release-please-manifest.json").read_text())
    tag = f"v{manifest['.']}"
    if git_ref_exists(tag, cwd):
        return tag

    config = json.loads((Path(cwd) / "release-please-config.json").read_text())
    bootstrap_sha = config.get("bootstrap-sha")
    if bootstrap_sha and git_ref_exists(bootstrap_sha, cwd):
        return bootstrap_sha

    raise SystemExit(f"could not resolve release base ref {tag}")


def commit_shas(base_ref, current_ref, cwd):
    raw = run_git(["log", "--format=%H", f"{base_ref}..{current_ref}", "--reverse"], cwd)
    return [line for line in raw.splitlines() if line]


def commit_files(sha, cwd):
    raw = run_git(["diff-tree", "--no-commit-id", "--name-only", "-r", sha], cwd)
    return [line for line in raw.splitlines() if line]


def is_website_only_path(path):
    return path in WEBSITE_EXACT_PATHS or any(path.startswith(prefix) for prefix in WEBSITE_PREFIXES)


def is_website_only_commit(files):
    return bool(files) and all(is_website_only_path(path) for path in files)


def release_bump(message):
    header, _, body = message.partition("\n")
    match = re.match(r"^(?P<type>[A-Za-z]+)(?:\([^)]+\))?(?P<breaking>!)?:\s+.+", header)
    if not match:
        return None
    if match.group("breaking") or re.search(r"(?m)^BREAKING[- ]CHANGE:\s+", body):
        return "major"
    kind = match.group("type")
    if kind == "feat":
        return "minor"
    if kind in {"fix", "perf"}:
        return "patch"
    return None


def next_version(version, bump):
    major, minor, patch = (int(part) for part in version.split("."))
    if bump == "major":
        return f"{major + 1}.0.0"
    if bump == "minor":
        return f"{major}.{minor + 1}.0"
    if bump == "patch":
        return f"{major}.{minor}.{patch + 1}"
    raise ValueError(f"unsupported bump: {bump}")


def included_release_commits(cwd, base_ref, current_ref):
    included_commits = []
    ignored_commits = []
    highest_bump = None

    for sha in commit_shas(base_ref, current_ref, cwd):
        files = commit_files(sha, cwd)
        if is_website_only_commit(files):
            ignored_commits.append(sha)
            continue

        included_commits.append(sha)
        message = run_git(["log", "-1", "--format=%B", sha], cwd)
        bump = release_bump(message)
        if not bump:
            continue
        if highest_bump is None or BUMP_ORDER[bump] > BUMP_ORDER[highest_bump]:
            highest_bump = bump

    return included_commits, ignored_commits, highest_bump


def create_filtered_repo(cwd, base_ref, target_branch, output_path, included_commits):
    output_path = output_path.resolve()
    source_path = Path(cwd).resolve()
    if output_path == source_path or source_path in output_path.parents:
        raise SystemExit("filtered repo must be outside the source checkout")
    if output_path.exists():
        if not output_path.is_dir() or any(output_path.iterdir()):
            raise SystemExit(f"filtered repo path is not empty: {output_path}")

    output_path.parent.mkdir(parents=True, exist_ok=True)
    run_git(["clone", "--quiet", "--no-local", "--no-hardlinks", str(source_path), str(output_path)], cwd)
    base_sha = run_git(["rev-parse", f"{base_ref}^{{commit}}"], output_path)
    visible_head = included_commits[-1] if included_commits else base_sha

    previous_visible = base_sha
    for sha in included_commits:
        parents = run_git(["rev-list", "--parents", "-n", "1", sha], output_path).split()[1:]
        if parents != [previous_visible]:
            run_git(["replace", "--graft", sha, previous_visible], output_path)
        previous_visible = sha

    run_git(["checkout", "--quiet", "-B", target_branch, visible_head], output_path)
    run_git(["remote", "remove", "origin"], output_path)
    run_git(["remote", "add", "origin", "."], output_path)
    run_git(["update-ref", f"refs/remotes/origin/{target_branch}", "HEAD"], output_path)
    return output_path


def analyze(cwd, base_ref, current_ref):
    manifest = json.loads((Path(cwd) / ".release-please-manifest.json").read_text())
    current_version = manifest["."]
    if not re.match(r"^[0-9]+\.[0-9]+\.[0-9]+$", current_version):
        raise SystemExit(f"unsupported PM version: {current_version}")

    included_commits, ignored_commits, highest_bump = included_release_commits(cwd, base_ref, current_ref)

    release_as = next_version(current_version, highest_bump) if highest_bump else ""
    return {
        "pm_release_input": "true" if highest_bump else "false",
        "release_as": release_as,
        "bump": highest_bump or "",
        "base_ref": base_ref,
        "current_ref": current_ref,
        "relevant_commits": len(included_commits),
        "ignored_website_commits": len(ignored_commits),
        "included_commit_shas": included_commits,
    }


def write_github_output(path, result):
    with open(path, "a", encoding="utf-8") as fh:
        for key, value in result.items():
            if isinstance(value, list):
                value = ",".join(value)
            fh.write(f"{key}={value}\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-ref")
    parser.add_argument("--current-ref", default="HEAD")
    parser.add_argument("--target-branch", default="main")
    parser.add_argument("--filtered-repo")
    parser.add_argument("--github-output")
    args = parser.parse_args()

    cwd = Path.cwd()
    base_ref = args.base_ref or default_base_ref(cwd)
    result = analyze(cwd, base_ref, args.current_ref)
    if args.filtered_repo:
        filtered_repo = create_filtered_repo(
            cwd,
            base_ref,
            args.target_branch,
            Path(args.filtered_repo),
            result["included_commit_shas"],
        )
        result["filtered_repo"] = str(filtered_repo)
    result.pop("included_commit_shas")

    if args.github_output:
        write_github_output(args.github_output, result)
    print(json.dumps(result, sort_keys=True))


if __name__ == "__main__":
    main()
