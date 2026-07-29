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


def analyze(cwd, base_ref, current_ref):
    manifest = json.loads((Path(cwd) / ".release-please-manifest.json").read_text())
    current_version = manifest["."]
    if not re.match(r"^[0-9]+\.[0-9]+\.[0-9]+$", current_version):
        raise SystemExit(f"unsupported PM version: {current_version}")

    highest_bump = None
    relevant_commits = []
    ignored_commits = []

    for sha in commit_shas(base_ref, current_ref, cwd):
        files = commit_files(sha, cwd)
        if is_website_only_commit(files):
            ignored_commits.append(sha)
            continue

        message = run_git(["log", "-1", "--format=%B", sha], cwd)
        bump = release_bump(message)
        if not bump:
            continue

        relevant_commits.append(sha)
        if highest_bump is None or BUMP_ORDER[bump] > BUMP_ORDER[highest_bump]:
            highest_bump = bump

    release_as = next_version(current_version, highest_bump) if highest_bump else ""
    return {
        "pm_release_input": "true" if highest_bump else "false",
        "release_as": release_as,
        "bump": highest_bump or "",
        "base_ref": base_ref,
        "current_ref": current_ref,
        "relevant_commits": len(relevant_commits),
        "ignored_website_commits": len(ignored_commits),
    }


def write_github_output(path, result):
    with open(path, "a", encoding="utf-8") as fh:
        for key, value in result.items():
            fh.write(f"{key}={value}\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-ref")
    parser.add_argument("--current-ref", default="HEAD")
    parser.add_argument("--github-output")
    args = parser.parse_args()

    cwd = Path.cwd()
    base_ref = args.base_ref or default_base_ref(cwd)
    result = analyze(cwd, base_ref, args.current_ref)

    if args.github_output:
        write_github_output(args.github_output, result)
    print(json.dumps(result, sort_keys=True))


if __name__ == "__main__":
    main()
