# Rebased binary measurements — issue #4299

Machine state: the same worktree, Go toolchain, dependency cache, and fixed
release-like link metadata. Baseline commit was rebased `origin/main`
`51dd6d468` before any production edit.

## Required local build

Command on the rebased baseline and again after the implementation:

```sh
go build ./cmd/pm && ls -l pm
```

| State | `pm` bytes |
| --- | ---: |
| Before | 180,124,498 |
| After | 180,124,498 |
| Delta | 0 |

## Release-like archive

Both builds used the same command except the explicit output path, with fixed
metadata so the linked source revision and build time cannot influence the
comparison:

```sh
go build -trimpath -ldflags '-s -w -X polymetrics.ai/internal/cli.version=0.0.0-embed-budget -X polymetrics.ai/internal/cli.commit=embed-budget-baseline -X polymetrics.ai/internal/cli.buildDate=2026-08-20T00:00:00Z' -o <output> ./cmd/pm
tar -czf <output>.tar.gz <output>
```

| State | Stripped `pm` bytes | `.tar.gz` bytes |
| --- | ---: | ---: |
| Before | 142,600,210 | 32,640,924 |
| After | 142,600,210 | 32,640,937 |
| Delta | 0 | +13 |

The 13-byte archive difference is from the intentionally different archive
entry names (`pm-release-before` and `pm-release-after`); the installed binary
is byte-identical in size. Current `main` carries only the retained GitHub
lock, so the immediate size delta is correctly zero. The explicit source-lock
allowlist and the installed-binary ceiling prevent the pending provider locks
from entering a shipped binary.

Both artifacts passed:

```text
release-size-report kind=archive subject=pm-release-before.tar.gz bytes=32640924 budget=50331648
release-size-report kind=installed_binary subject=pm-release-before.tar.gz!pm-release-before bytes=142600210 budget=167772160
release-size-report kind=archive subject=pm-release-after.tar.gz bytes=32640937 budget=50331648
release-size-report kind=installed_binary subject=pm-release-after.tar.gz!pm-release-after bytes=142600210 budget=167772160
```
