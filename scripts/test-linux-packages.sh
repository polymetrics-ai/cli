#!/usr/bin/env bash
set -euo pipefail

DIST_DIR=${1:-dist}

if ! command -v docker >/dev/null 2>&1; then
  printf 'docker is required for Linux package install tests\n' >&2
  exit 1
fi
if [[ ! -d "$DIST_DIR" ]]; then
  printf 'release asset directory not found: %s\n' "$DIST_DIR" >&2
  exit 1
fi

abs_dist=$(cd "$DIST_DIR" && pwd -P)

container_arch=$(docker run --rm ubuntu:24.04 uname -m)
case "$container_arch" in
  x86_64|amd64)
    deb_arch=amd64
    rpm_arch=x86_64
    ;;
  aarch64|arm64)
    deb_arch=arm64
    rpm_arch=aarch64
    ;;
  *)
    printf 'unsupported Docker architecture for package install test: %s\n' "$container_arch" >&2
    exit 1
    ;;
esac

shopt -s nullglob
debs=("$abs_dist"/pm_*_linux_"$deb_arch".deb)
rpms=("$abs_dist"/pm_*_linux_"$rpm_arch".rpm)
if [[ ${#debs[@]} -ne 1 ]]; then
  printf 'expected one %s deb package, found %d\n' "$deb_arch" "${#debs[@]}" >&2
  exit 1
fi
if [[ ${#rpms[@]} -ne 1 ]]; then
  printf 'expected one %s rpm package, found %d\n' "$rpm_arch" "${#rpms[@]}" >&2
  exit 1
fi

deb_pkg=/dist/$(basename "${debs[0]}")
rpm_pkg=/dist/$(basename "${rpms[0]}")

# Debian/Ubuntu-family coverage. Install from the standalone package, exercise
# non-credentialed commands, reinstall the same package to exercise package
# manager replacement/upgrade semantics, then uninstall.
docker run --rm \
  --volume "$abs_dist:/dist:ro" \
  --network bridge \
  ubuntu:24.04 \
  bash -euo pipefail -c '
    export DEBIAN_FRONTEND=noninteractive
    apt-get update >/dev/null
    apt-get install -y --no-install-recommends ca-certificates >/dev/null
    apt-get install -y "$1" >/dev/null
    pm version --json >/tmp/pm-version.json
    grep -q "\"version\"" /tmp/pm-version.json
    pm connectors inspect sample --json >/tmp/pm-sample.json
    grep -q "\"name\"" /tmp/pm-sample.json
    apt-get install -y --reinstall "$1" >/dev/null
    pm version --json >/dev/null
    apt-get remove -y pm >/dev/null
    hash -r || true
    if dpkg -s pm >/dev/null 2>&1 || [ -e /usr/bin/pm ]; then
      echo "pm still present after deb uninstall" >&2
      exit 1
    fi
  ' _ "$deb_pkg"

# Fedora/RHEL-family coverage. Use dnf against the standalone package without
# configuring any signed repository; issue #553 owns repository metadata/signing.
docker run --rm \
  --volume "$abs_dist:/dist:ro" \
  --network bridge \
  fedora:44 \
  bash -euo pipefail -c '
    dnf install -y ca-certificates >/dev/null
    dnf install -y "$1" >/dev/null
    pm version --json >/tmp/pm-version.json
    grep -q "\"version\"" /tmp/pm-version.json
    pm connectors inspect sample --json >/tmp/pm-sample.json
    grep -q "\"name\"" /tmp/pm-sample.json
    rpm -Uvh --replacepkgs "$1" >/dev/null
    pm version --json >/dev/null
    dnf remove -y pm >/dev/null
    hash -r || true
    if rpm -q pm >/dev/null 2>&1 || [ -e /usr/bin/pm ]; then
      echo "pm still present after rpm uninstall" >&2
      exit 1
    fi
  ' _ "$rpm_pkg"

printf 'linux package install/reinstall/uninstall tests passed for %s\n' "$DIST_DIR"
