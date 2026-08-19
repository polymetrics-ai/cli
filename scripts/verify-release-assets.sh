#!/usr/bin/env bash
set -euo pipefail

DIST_DIR=dist
if [[ $# -gt 0 && "$1" != --* ]]; then
  DIST_DIR=$1
  shift
fi

PRINT_SUBJECTS=0
PRINT_EXPECTED_RELEASE_ASSETS=0
REQUIRE_EXACT_RELEASE_ASSETS=0
RELEASE_VERSION=${RELEASE_VERSION:-}
REQUIRE_TRUST_EVIDENCE=${REQUIRE_TRUST_EVIDENCE:-0}
VERIFY_GITHUB_ATTESTATIONS=${VERIFY_GITHUB_ATTESTATIONS:-0}
ALLOW_UNSIGNED_TRUST_FIXTURES=${ALLOW_UNSIGNED_TRUST_FIXTURES:-0}
COSIGN_CERT_OIDC_ISSUER=${COSIGN_CERT_OIDC_ISSUER:-https://token.actions.githubusercontent.com}
GITHUB_ATTESTATION_REPO=${GITHUB_ATTESTATION_REPO:-${GITHUB_REPOSITORY:-polymetrics-ai/cli}}
GITHUB_ATTESTATION_CERT_OIDC_ISSUER=${GITHUB_ATTESTATION_CERT_OIDC_ISSUER:-$COSIGN_CERT_OIDC_ISSUER}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --print-subjects)
      PRINT_SUBJECTS=1
      ;;
    --print-expected-release-assets)
      PRINT_EXPECTED_RELEASE_ASSETS=1
      ;;
    --require-exact-release-assets)
      REQUIRE_EXACT_RELEASE_ASSETS=1
      ;;
    --release-version)
      if [[ $# -lt 2 ]]; then
        printf '%s\n' '--release-version requires a value' >&2
        exit 2
      fi
      RELEASE_VERSION=$2
      shift
      ;;
    --require-trust-evidence)
      REQUIRE_TRUST_EVIDENCE=1
      ;;
    --verify-github-attestations)
      VERIFY_GITHUB_ATTESTATIONS=1
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      exit 2
      ;;
  esac
  shift
done

# Windows is deliberately absent, both architectures. windows/arm64 was never
# buildable — pm embeds DuckDB and go-duckdb ships no library for it — and
# windows/amd64 was dropped because pm has no Windows customer. It returns on a
# customer ask, restored from git history.
archive_targets=(
  "darwin amd64 tar.gz pm"
  "darwin arm64 tar.gz pm"
  "linux amd64 tar.gz pm"
  "linux arm64 tar.gz pm"
)

package_targets=(
  "deb amd64 amd64"
  "deb arm64 arm64"
  "rpm amd64 x86_64"
  "rpm arm64 aarch64"
)

validate_release_version() {
  local version=$1
  if [[ ! "$version" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]*$ ]]; then
    printf 'invalid release version for asset names: %s\n' "$version" >&2
    exit 2
  fi
}

require_release_version() {
  if [[ -z "$RELEASE_VERSION" ]]; then
    printf '%s\n' '--release-version or RELEASE_VERSION is required for exact release asset checks' >&2
    exit 2
  fi
  validate_release_version "$RELEASE_VERSION"
  printf '%s\n' "$RELEASE_VERSION"
}

expected_subject_names_for_version() {
  local version=$1
  local target goos goarch extension format package_arch
  validate_release_version "$version"

  for target in "${archive_targets[@]}"; do
    read -r goos goarch extension _ <<<"$target"
    printf 'pm_%s_%s_%s.%s\n' "$version" "$goos" "$goarch" "$extension"
  done

  for target in "${package_targets[@]}"; do
    read -r format _ package_arch <<<"$target"
    printf 'pm_%s_linux_%s.%s\n' "$version" "$package_arch" "$format"
  done

  printf 'checksums.txt\n'
}

expected_release_asset_names_for_version() {
  local version=$1 subject
  while IFS= read -r subject; do
    printf '%s\n' "$subject"
    printf '%s.sigstore.json\n' "$subject"
  done < <(expected_subject_names_for_version "$version")
}

if [[ "$PRINT_EXPECTED_RELEASE_ASSETS" == "1" ]]; then
  expected_release_asset_names_for_version "$(require_release_version)"
  exit 0
fi

if [[ ! -d "$DIST_DIR" ]]; then
  printf 'release asset directory not found: %s\n' "$DIST_DIR" >&2
  exit 1
fi
if [[ ! -f "$DIST_DIR/checksums.txt" ]]; then
  printf 'release checksum manifest not found: %s/checksums.txt\n' "$DIST_DIR" >&2
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf '%s is required to verify release assets\n' "$1" >&2
    exit 1
  fi
}

sha256_file() {
  local file=$1
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{ print $1 }'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{ print $1 }'
  else
    printf 'sha256sum or shasum is required to verify release assets\n' >&2
    exit 1
  fi
}

one_match() {
  local description=$1
  local pattern=$2
  shift 2
  local matches=("$@")
  if [[ ${#matches[@]} -ne 1 ]]; then
    printf 'expected one %s release asset, found %d\n' "$description" "${#matches[@]}" >&2
    printf 'pattern: %s\n' "$pattern" >&2
    exit 1
  fi
  printf '%s\n' "${matches[0]}"
}

compare_exact() {
  local label=$1
  local expected=$2
  local actual=$3
  if [[ "$actual" != "$expected" ]]; then
    printf 'unexpected %s\n' "$label" >&2
    diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$actual") || true
    exit 1
  fi
}

compare_exact_release_assets() {
  local version=$1
  local directory=$2
  local expected actual
  expected=$(expected_release_asset_names_for_version "$version" | LC_ALL=C sort)
  actual=$(find "$directory" -maxdepth 1 -type f -exec basename {} \; | LC_ALL=C sort)
  compare_exact "release asset set: $directory" "$expected" "$actual"
}

control_field() {
  local fields=$1
  local key=$2
  awk -F': ' -v key="$key" '$1 == key { print substr($0, index($0, $2)); exit }' <<<"$fields"
}

require_deb_field() {
  local package=$1
  local fields=$2
  local key=$3
  local want=$4
  local got
  got=$(control_field "$fields" "$key")
  if [[ "$got" != "$want" ]]; then
    printf 'unexpected %s field in %s: got %q want %q\n' "$key" "$package" "$got" "$want" >&2
    exit 1
  fi
}

verify_deb_package() {
  local package=$1
  local arch=$2
  require_cmd dpkg-deb

  local fields
  fields=$(dpkg-deb --field "$package")
  require_deb_field "$package" "$fields" Package pm
  require_deb_field "$package" "$fields" Architecture "$arch"
  require_deb_field "$package" "$fields" Maintainer "Polymetrics AI"
  require_deb_field "$package" "$fields" Homepage "https://cli.polymetrics.ai"

  local description
  description=$(control_field "$fields" Description)
  if [[ "$description" != *"Local-first data CLI"* ]]; then
    printf 'unexpected Description field in %s: %q\n' "$package" "$description" >&2
    exit 1
  fi

  local contents expected
  contents=$(dpkg-deb --contents "$package" | awk '$1 !~ /^d/ { print $6 }' | sed 's#^\./#/#' | LC_ALL=C sort)
  expected=$(printf '%s\n' /usr/bin/pm /usr/share/doc/pm/LICENSE /usr/share/doc/pm/NOTICE | LC_ALL=C sort)
  compare_exact "deb contents: $package" "$expected" "$contents"
}

rpm_query() {
  local package=$1
  local format=$2
  rpm -qp --queryformat "$format" "$package" 2>/dev/null
}

verify_rpm_package() {
  local package=$1
  local arch=$2
  require_cmd rpm

  local metadata
  metadata=$(rpm_query "$package" '%{NAME}\n%{ARCH}\n%{LICENSE}\n%{URL}\n%{VENDOR}\n%{PACKAGER}\n%{SUMMARY}\n')
  local name rpm_arch license url vendor packager summary
  name=$(sed -n '1p' <<<"$metadata")
  rpm_arch=$(sed -n '2p' <<<"$metadata")
  license=$(sed -n '3p' <<<"$metadata")
  url=$(sed -n '4p' <<<"$metadata")
  vendor=$(sed -n '5p' <<<"$metadata")
  packager=$(sed -n '6p' <<<"$metadata")
  summary=$(sed -n '7p' <<<"$metadata")
  if [[ "$name" != pm || "$rpm_arch" != "$arch" || "$license" != AGPL-3.0-only || "$url" != https://cli.polymetrics.ai || "$vendor" != "Polymetrics AI" ]]; then
    printf 'unexpected rpm metadata in %s\n%s\n' "$package" "$metadata" >&2
    exit 1
  fi
  if [[ "$packager" != "Polymetrics AI" && "$packager" != "(none)" ]]; then
    printf 'unexpected rpm packager in %s: %q\n' "$package" "$packager" >&2
    exit 1
  fi
  if [[ "$summary" != *"Local-first data CLI"* ]]; then
    printf 'unexpected rpm summary in %s: %q\n' "$package" "$summary" >&2
    exit 1
  fi

  local contents expected
  contents=$(rpm -qlp "$package" | grep -Ev '^/usr/share/doc/pm/?$' | LC_ALL=C sort)
  expected=$(printf '%s\n' /usr/bin/pm /usr/share/doc/pm/LICENSE /usr/share/doc/pm/NOTICE | LC_ALL=C sort)
  compare_exact "rpm contents: $package" "$expected" "$contents"
}

verify_unsigned_fixture() {
  local subject=$1
  local bundle=$2
  local digest
  digest=$(sha256_file "$subject")
  python3 - "$subject" "$bundle" "$digest" <<'PY'
import json
import os
import sys

subject, bundle, digest = sys.argv[1:4]
with open(bundle, "r", encoding="utf-8") as fh:
    data = json.load(fh)
if data.get("_pm_unsigned_fixture") is not True:
    raise SystemExit(f"{bundle}: missing _pm_unsigned_fixture=true")
subject_data = data.get("subject") or {}
if subject_data.get("name") != os.path.basename(subject):
    raise SystemExit(f"{bundle}: subject name mismatch")
if subject_data.get("sha256") != digest:
    raise SystemExit(f"{bundle}: sha256 mismatch")
if "UNSIGNED" not in data.get("warning", ""):
    raise SystemExit(f"{bundle}: fixture warning missing")
PY
}

verify_cosign_bundle() {
  local subject=$1
  local bundle=$2
  if [[ "$ALLOW_UNSIGNED_TRUST_FIXTURES" == "1" ]]; then
    verify_unsigned_fixture "$subject" "$bundle"
    return
  fi

  require_cmd cosign
  local identity_args=()
  if [[ -n "${COSIGN_CERT_IDENTITY_REGEXP:-}" ]]; then
    identity_args=(--certificate-identity-regexp "$COSIGN_CERT_IDENTITY_REGEXP")
  elif [[ -n "${COSIGN_CERT_IDENTITY:-}" ]]; then
    identity_args=(--certificate-identity "$COSIGN_CERT_IDENTITY")
  else
    printf 'COSIGN_CERT_IDENTITY or COSIGN_CERT_IDENTITY_REGEXP is required when verifying real Cosign bundles\n' >&2
    exit 1
  fi
  cosign verify-blob "$subject" \
    --bundle "$bundle" \
    "${identity_args[@]}" \
    --certificate-oidc-issuer "$COSIGN_CERT_OIDC_ISSUER"
}

verify_github_attestation() {
  local subject=$1
  if [[ "$VERIFY_GITHUB_ATTESTATIONS" != "1" ]]; then
    return
  fi
  require_cmd gh
  local identity_args=()
  if [[ -n "${GITHUB_ATTESTATION_CERT_IDENTITY_REGEX:-}" ]]; then
    identity_args=(--cert-identity-regex "$GITHUB_ATTESTATION_CERT_IDENTITY_REGEX")
  elif [[ -n "${GITHUB_ATTESTATION_CERT_IDENTITY:-}" ]]; then
    identity_args=(--cert-identity "$GITHUB_ATTESTATION_CERT_IDENTITY")
  else
    printf 'GITHUB_ATTESTATION_CERT_IDENTITY or GITHUB_ATTESTATION_CERT_IDENTITY_REGEX is required when verifying GitHub attestations\n' >&2
    exit 1
  fi
  gh attestation verify "$subject" \
    --repo "$GITHUB_ATTESTATION_REPO" \
    "${identity_args[@]}" \
    --cert-oidc-issuer "$GITHUB_ATTESTATION_CERT_OIDC_ISSUER" \
    --predicate-type https://slsa.dev/provenance/v1 \
    --deny-self-hosted-runners >/dev/null
}

shopt -s nullglob
assets=()
asset_paths=()
for target in "${archive_targets[@]}"; do
  read -r goos goarch extension binary_name <<<"$target"
  pattern="$DIST_DIR/pm_*_${goos}_${goarch}.${extension}"
  matches=("$DIST_DIR"/pm_*_"${goos}"_"${goarch}"."${extension}")
  asset=$(one_match "$goos/$goarch archive" "$pattern" "${matches[@]}")
  assets+=("$(basename "$asset")")
  asset_paths+=("$asset")

  if [[ "$extension" == "zip" ]]; then
    contents=$(unzip -Z1 "$asset" | LC_ALL=C sort)
  else
    contents=$(tar -tzf "$asset" | LC_ALL=C sort)
  fi
  expected=$(printf '%s\n' LICENSE NOTICE README.md "$binary_name" | LC_ALL=C sort)
  compare_exact "archive contents: $asset" "$expected" "$contents"
done

for target in "${package_targets[@]}"; do
  read -r format goarch package_arch <<<"$target"
  pattern="$DIST_DIR/pm_*_linux_${package_arch}.${format}"
  matches=("$DIST_DIR"/pm_*_linux_"${package_arch}"."${format}")
  asset=$(one_match "linux/$goarch $format package" "$pattern" "${matches[@]}")
  assets+=("$(basename "$asset")")
  asset_paths+=("$asset")

  case "$format" in
    deb) verify_deb_package "$asset" "$package_arch" ;;
    rpm) verify_rpm_package "$asset" "$package_arch" ;;
  esac
done

expected_names=$(printf '%s\n' "${assets[@]}" | LC_ALL=C sort)
manifest_names=$(
  awk '
    NF != 2 || $1 !~ /^[0-9a-fA-F]{64}$/ || $2 ~ /^\// || $2 ~ /\.\./ || $2 ~ /\\/ { exit 2 }
    { print $2 }
  ' "$DIST_DIR/checksums.txt" | LC_ALL=C sort
) || {
  printf 'invalid checksum manifest format: %s/checksums.txt\n' "$DIST_DIR" >&2
  exit 1
}
if [[ "$manifest_names" != "$expected_names" ]]; then
  printf 'checksum manifest does not cover exactly the expected release assets\n' >&2
  diff -u <(printf '%s\n' "$expected_names") <(printf '%s\n' "$manifest_names") || true
  exit 1
fi

checksum_output=''
if command -v sha256sum >/dev/null 2>&1; then
  checksum_output=$(cd "$DIST_DIR" && sha256sum --check checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
  checksum_output=$(cd "$DIST_DIR" && shasum -a 256 --check checksums.txt)
else
  printf 'sha256sum or shasum is required to verify release assets\n' >&2
  exit 1
fi
if [[ "$PRINT_SUBJECTS" != "1" ]]; then
  printf '%s\n' "$checksum_output"
fi

subjects=("${asset_paths[@]}" "$DIST_DIR/checksums.txt")

if [[ "$REQUIRE_EXACT_RELEASE_ASSETS" == "1" ]]; then
  compare_exact_release_assets "$(require_release_version)" "$DIST_DIR"
fi

if [[ "$REQUIRE_TRUST_EVIDENCE" == "1" ]]; then
  for subject in "${subjects[@]}"; do
    bundle="$subject.sigstore.json"
    if [[ ! -f "$bundle" ]]; then
      printf 'missing Cosign bundle for release subject: %s\n' "$subject" >&2
      exit 1
    fi
    verify_cosign_bundle "$subject" "$bundle"
    verify_github_attestation "$subject"
  done
fi

if [[ "$PRINT_SUBJECTS" == "1" ]]; then
  printf '%s\n' "${subjects[@]}"
  exit 0
fi

if [[ "$REQUIRE_TRUST_EVIDENCE" == "1" ]]; then
  printf 'verified %d release assets plus trust evidence for %d subjects in %s\n' "${#assets[@]}" "${#subjects[@]}" "$DIST_DIR"
else
  printf 'verified %d release assets in %s\n' "${#assets[@]}" "$DIST_DIR"
fi
