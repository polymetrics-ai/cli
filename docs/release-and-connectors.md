# PM v0.1.0 release and connector shipping

This guide covers the `pm` binary release path and how connector changes ship in
that binary. It does not authorize a tag, GitHub Release, website deploy, or
merge.

## Product boundary

Website deployment and PM binary releases are independent and never trigger each
other. Website checks and deploys are owned by `.github/workflows/website.yml`;
PM binary releases are owned by `.github/workflows/release.yml`, Release Please,
and GoReleaser.

Do not dispatch or deploy the website for a PM binary release. Do not cut a PM
binary release because the website deploys.

## First release target

The first PM binary release target is `v0.1.0`, not `v1.0.0`.

Use Release Please's supported one-shot version selection: a commit body
containing `Release-As: 0.1.0`. Do not add a persistent `release-as` field to
`release-please-config.json`; that would keep forcing later releases back to
`0.1.0`.

If this preparation PR is squash-merged, the squash commit body must preserve
`Release-As: 0.1.0`; otherwise Release Please will not see the one-shot override
on `main`.

The release flow is:

1. Release Please runs from `main` and opens or updates a release PR using
   `release-please-config.json` and `.release-please-manifest.json`. For the
   first release, that PR must be `chore: release 0.1.0`; if it proposes any
   other version, stop and fix the release preparation before merge.
2. The captain owns the release-PR merge after the mandatory Bahmni corrective
   PR is merged, the intended `main` commit is green, and the generated
   changelog/manifest update is reviewed.
3. Release Please creates the `v0.1.0` tag and GitHub Release from the merged
   release PR. Do not create the tag or release assets by hand.
4. The release workflow uses GoReleaser to build the publication set from the
   release tag, verifies it, and uploads these assets:
   - `pm_0.1.0_darwin_amd64.tar.gz`
   - `pm_0.1.0_darwin_arm64.tar.gz`
   - `pm_0.1.0_linux_amd64.tar.gz`
   - `pm_0.1.0_linux_arm64.tar.gz`
   - `pm_0.1.0_windows_amd64.zip`
   - `pm_0.1.0_windows_arm64.zip`
   - `checksums.txt`

## Install on macOS or Linux

After `v0.1.0` is cut, this example downloads the archive for the current
macOS/Linux host, verifies its SHA-256 checksum, installs `pm`, and runs
`pm version` as a smoke test.

```bash
release=v0.1.0
version=${release#v}
repo=polymetrics-ai/cli

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

archive="pm_${version}_${os}_${arch}.tar.gz"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

gh release download "$release" --repo "$repo" \
  --pattern "$archive" \
  --pattern checksums.txt \
  --dir "$tmpdir"

awk -v archive="$archive" '$2 == archive { print }' \
  "$tmpdir/checksums.txt" > "$tmpdir/checksums.selected.txt"
test -s "$tmpdir/checksums.selected.txt"

if command -v sha256sum >/dev/null 2>&1; then
  (cd "$tmpdir" && sha256sum --check checksums.selected.txt)
else
  (cd "$tmpdir" && shasum -a 256 --check checksums.selected.txt)
fi

tar -xzf "$tmpdir/$archive" -C "$tmpdir"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"
install -m 0755 "$tmpdir/pm" "$install_dir/pm"
"$install_dir/pm" version
```

Use an `INSTALL_DIR` already on your `PATH`, or add it before running `pm` from a
new shell.

## Connector release model

Connectors are embedded in the PM binary. There is currently no separately
versioned connector package, registry artifact, or downloadable connector
archive.

A released PM binary contains exactly the connector code and definitions merged
into the commit tagged for that release. For `v0.1.0`, only connectors merged
into the exact `v0.1.0` commit are included; `pm connectors list` from that
binary is the release truth.

After `v0.1.0`, a compatible connector fix normally ships as the next patch
release, such as `v0.1.1`. A new connector or user-facing connector feature
normally ships as the next pre-1.0 minor release, such as `v0.2.0`.

Do not document any connector as included in a PM release until its change is
merged into the exact release commit. WhatsApp is not part of `v0.1.0` unless its
PR is merged into that exact release commit.
