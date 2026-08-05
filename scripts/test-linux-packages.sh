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

shopt -s nullglob

one_package() {
  local arch=$1
  local format=$2
  shift 2
  local matches=("$@")
  if [[ ${#matches[@]} -ne 1 ]]; then
    printf 'expected one %s %s package, found %d\n' "$arch" "$format" "${#matches[@]}" >&2
    exit 1
  fi
  printf '%s\n' "${matches[0]}"
}

check_platform() {
  local docker_platform=$1
  local expected_uname=$2
  local actual_uname
  actual_uname=$(docker run --rm --platform "$docker_platform" ubuntu:24.04 uname -m)
  if [[ "$actual_uname" != "$expected_uname" ]]; then
    printf 'docker platform %s reported uname %s; want %s\n' "$docker_platform" "$actual_uname" "$expected_uname" >&2
    exit 1
  fi
}

test_deb_package() {
  local docker_platform=$1
  local expected_uname=$2
  local deb_arch=$3
  local deb_pkg=$4

  docker run --rm \
    --platform "$docker_platform" \
    --volume "$abs_dist:/dist:ro" \
    --network bridge \
    ubuntu:24.04 \
    bash -euo pipefail -c '
      if [ "$(uname -m)" != "$1" ]; then
        echo "unexpected deb test architecture: $(uname -m), want $1" >&2
        exit 1
      fi
      if [ "$(dpkg --print-architecture)" != "$2" ]; then
        echo "unexpected dpkg architecture: $(dpkg --print-architecture), want $2" >&2
        exit 1
      fi
      export DEBIAN_FRONTEND=noninteractive
      apt_update_attempt=1
      until apt-get update >/dev/null; do
        if [ "$apt_update_attempt" -ge 4 ]; then
          echo "apt-get update failed after $apt_update_attempt attempts" >&2
          exit 1
        fi
        echo "apt-get update attempt $apt_update_attempt failed; retrying after mirror sync" >&2
        sleep $((apt_update_attempt * 5))
        apt_update_attempt=$((apt_update_attempt + 1))
      done
      apt-get install -y --no-install-recommends ca-certificates >/dev/null
      apt-get install -y "$3" >/dev/null
      pm version --json >/tmp/pm-version.json
      grep -q "\"version\"" /tmp/pm-version.json
      pm connectors inspect sample --json >/tmp/pm-sample.json
      grep -q "\"name\"" /tmp/pm-sample.json
      apt-get install -y --reinstall "$3" >/dev/null
      pm version --json >/dev/null
      apt-get remove -y pm >/dev/null
      hash -r || true
      if dpkg -s pm >/dev/null 2>&1 || [ -e /usr/bin/pm ]; then
        echo "pm still present after deb uninstall" >&2
        exit 1
      fi
    ' _ "$expected_uname" "$deb_arch" "$deb_pkg"
}

test_rpm_package() {
  local docker_platform=$1
  local expected_uname=$2
  local rpm_arch=$3
  local rpm_pkg=$4

  docker run --rm \
    --platform "$docker_platform" \
    --volume "$abs_dist:/dist:ro" \
    --network bridge \
    fedora:44 \
    bash -euo pipefail -c '
      if [ "$(uname -m)" != "$1" ]; then
        echo "unexpected rpm test architecture: $(uname -m), want $1" >&2
        exit 1
      fi
      if [ "$(rpm --eval "%{_arch}")" != "$2" ]; then
        echo "unexpected rpm architecture: $(rpm --eval "%{_arch}"), want $2" >&2
        exit 1
      fi
      dnf install -y ca-certificates >/dev/null
      dnf install -y "$3" >/dev/null
      pm version --json >/tmp/pm-version.json
      grep -q "\"version\"" /tmp/pm-version.json
      pm connectors inspect sample --json >/tmp/pm-sample.json
      grep -q "\"name\"" /tmp/pm-sample.json
      rpm -Uvh --replacepkgs "$3" >/dev/null
      pm version --json >/dev/null
      dnf remove -y pm >/dev/null
      hash -r || true
      if rpm -q pm >/dev/null 2>&1 || [ -e /usr/bin/pm ]; then
        echo "pm still present after rpm uninstall" >&2
        exit 1
      fi
    ' _ "$expected_uname" "$rpm_arch" "$rpm_pkg"
}

test_architecture() {
  local goarch=$1
  local docker_platform=$2
  local expected_uname=$3
  local deb_arch=$4
  local rpm_arch=$5
  local debs rpms deb_pkg rpm_pkg

  check_platform "$docker_platform" "$expected_uname"

  debs=("$abs_dist"/pm_*_linux_"$deb_arch".deb)
  rpms=("$abs_dist"/pm_*_linux_"$rpm_arch".rpm)
  deb_pkg=/dist/$(basename "$(one_package "$goarch" deb "${debs[@]}")")
  rpm_pkg=/dist/$(basename "$(one_package "$goarch" rpm "${rpms[@]}")")

  test_deb_package "$docker_platform" "$expected_uname" "$deb_arch" "$deb_pkg"
  test_rpm_package "$docker_platform" "$expected_uname" "$rpm_arch" "$rpm_pkg"
}

test_architecture amd64 linux/amd64 x86_64 amd64 x86_64
test_architecture arm64 linux/arm64 aarch64 arm64 aarch64

printf 'linux package install/reinstall/uninstall tests passed for amd64 and arm64 in %s\n' "$DIST_DIR"
