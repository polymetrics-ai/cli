#!/usr/bin/env python3
"""Append the Crisp captain-policy addendum to issue bodies idempotently via gh-axi."""
from __future__ import annotations

import ast
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent
ADDENDUM = (ROOT / "captain-policy-addendum.md").read_text()
MARKER = "fm-crisp-captain-policy-addendum-v1"
ISSUES = [204, 205, 206, 207, 208, 209, 210, 211]
REPO = "polymetrics-ai/cli"
TMP = ROOT / ".issue-body-tmp"
TMP.mkdir(exist_ok=True)
LOG = ROOT / "captain-policy-addendum.log"


def run(args: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)


def issue_body(number: int) -> str:
    cp = run(["gh-axi", "issue", "view", str(number), "-R", REPO, "--full"])
    match = re.search(r'^  body: (".*")$', cp.stdout, re.MULTILINE)
    if not match:
        raise RuntimeError(f"could not parse body for issue {number}")
    return ast.literal_eval(match.group(1))


def main() -> int:
    lines: list[str] = []
    for number in ISSUES:
        body = issue_body(number)
        if MARKER in body:
            lines.append(f"#{number}: already-present")
            continue
        new_body = body.rstrip() + "\n\n" + ADDENDUM.rstrip() + "\n"
        body_file = TMP / f"issue-{number}.md"
        body_file.write_text(new_body)
        run(["gh-axi", "issue", "edit", str(number), "-R", REPO, "--body-file", str(body_file)])
        lines.append(f"#{number}: appended")
    LOG.write_text("\n".join(lines) + "\n")
    print(LOG.read_text(), end="")
    return 0


if __name__ == "__main__":
    sys.exit(main())
