#!/usr/bin/env python3
"""Task-owned Bahmni Podman synthetic lab helper.

This module intentionally uses only the Python standard library. It prepares a
pinned official Bahmni Docker checkout in a local runtime directory, patches it
for rootless Podman loopback use, and seeds synthetic data through exposed
Bahmni/OpenMRS/FHIR APIs.
"""

from __future__ import annotations

import argparse
import base64
import datetime as dt
import hashlib
import json
import os
import re
import shutil
import ssl
import stat
import subprocess
import sys
import unicodedata
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Iterable

ROOT = Path(__file__).resolve().parents[1]
PIN_PATH = ROOT / "config" / "pin.json"
SEED_PATH = ROOT / "fixtures" / "synthetic-seed.json"
TAXONOMY_PATH = ROOT / "config" / "sparsh-hennur-taxonomy.json"
MODULE_SUPPORT_PATH = ROOT / "config" / "module-support.json"

LAB_ID = "fm-bahmni-lab-r1"
DEFAULT_HOME = Path("/tmp") / LAB_ID
OWNER_LABEL = "io.polymetrics.bahmni-lab.owner"
OWNER_MARKER_NAME = f".{LAB_ID}-ownership.json"
STABLE_RECORD_PREFIX = f"urn:polymetrics:bahmni-lab:{LAB_ID}"
MACHINE_NAME_RE = re.compile(rf"^{re.escape(LAB_ID)}-[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$")
PROJECT_NAME_RE = re.compile(r"^fm_bahmni_lab_r1(?:_[a-z0-9](?:[a-z0-9_]*[a-z0-9])?)?$")
PHONE_CANDIDATE_RE = re.compile(r"(?:\+|00)?\d(?:[^\w@]*\d){7,13}")
EMAIL_CANDIDATE_RE = re.compile(r"(?<![\w.!#$%&'*+/=?^`{|}~-])([\w.!#$%&'*+/=?^`{|}~-]+@[^\s<>()\[\]{},;:\"']+)")


def eprint(*parts: object) -> None:
    print(*parts, file=sys.stderr)


def load_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def dump_json(data: Any) -> None:
    print(json.dumps(data, indent=2, sort_keys=True, ensure_ascii=False))


def env_value(name: str, default: str) -> str:
    value = os.environ.get(name)
    return value if value not in (None, "") else default


def runtime_home() -> Path:
    return Path(env_value("BAHMNI_LAB_HOME", str(DEFAULT_HOME))).expanduser().resolve()


def project_name() -> str:
    pin = load_json(PIN_PATH)
    return env_value("BAHMNI_LAB_PROJECT", pin["compose_project"])


def machine_name() -> str:
    return env_value("BAHMNI_LAB_MACHINE", "fm-bahmni-lab-r1-machine")


def connection_name() -> str:
    return env_value("BAHMNI_LAB_CONNECTION", machine_name())


def port(name: str, default: int) -> str:
    env_name = f"BAHMNI_LAB_{name}_PORT"
    return env_value(env_name, str(default))


def repo_dir() -> Path:
    return runtime_home() / "bahmni-docker"


def standard_dir() -> Path:
    return repo_dir() / "bahmni-standard"


def generated_env_path() -> Path:
    return standard_dir() / ".env.fm"


def generated_compose_path() -> Path:
    return standard_dir() / "docker-compose.podman.yml"


def credentials_path() -> Path:
    return runtime_home() / ".local-credentials.json"


def podman_wrapper_path() -> Path:
    return runtime_home() / "bin" / "podman-fm"


def ownership_marker_path() -> Path:
    configured = os.environ.get("XDG_STATE_HOME")
    state_home = Path(configured).expanduser() if configured else Path.home() / ".local" / "state"
    if not state_home.is_absolute():
        raise ValueError("XDG_STATE_HOME must be an absolute path")
    return state_home.resolve() / LAB_ID / "ownership.json"


def legacy_ownership_marker_path() -> Path:
    return runtime_home() / OWNER_MARKER_NAME


def legacy_ownership_marker_paths() -> tuple[Path, ...]:
    return (runtime_home().parent / OWNER_MARKER_NAME, legacy_ownership_marker_path())


def validate_ownership_config(home: Path, machine: str, connection: str, project: str) -> dict[str, Any]:
    resolved = home.expanduser().resolve()
    failures: list[str] = []
    if resolved.name != LAB_ID or resolved == Path(resolved.anchor):
        failures.append(f"lab home basename must be exactly {LAB_ID}")
    if not MACHINE_NAME_RE.fullmatch(machine):
        failures.append(f"machine name must use the {LAB_ID}- prefix")
    if not MACHINE_NAME_RE.fullmatch(connection):
        failures.append(f"connection name must use the {LAB_ID}- prefix")
    if connection != machine:
        failures.append("connection name must match the task-owned machine name")
    if not PROJECT_NAME_RE.fullmatch(project):
        failures.append("compose project must use the fm_bahmni_lab_r1 prefix")
    if failures:
        raise ValueError("; ".join(failures))
    return {
        "schema": 2,
        "owner": LAB_ID,
        "lab_home": str(resolved),
        "machine": machine,
        "connection": connection,
        "project": project,
        "resources": [],
    }


def expected_ownership() -> dict[str, Any]:
    return validate_ownership_config(runtime_home(), machine_name(), connection_name(), project_name())


def validate_ownership_marker(actual: dict[str, Any], expected: dict[str, Any]) -> dict[str, Any]:
    identity_keys = ("owner", "lab_home", "machine", "connection", "project")
    if actual.get("schema") not in (1, 2) or any(actual.get(key) != expected.get(key) for key in identity_keys):
        raise ValueError("ownership marker does not match the requested lab resources")
    raw_resources = actual.get("resources", [])
    if not isinstance(raw_resources, list):
        raise ValueError("ownership marker contains invalid resource proof")
    resources: list[dict[str, str]] = []
    seen: set[tuple[str, str, str]] = set()
    for resource in raw_resources:
        if not isinstance(resource, dict):
            raise ValueError("ownership marker contains invalid resource proof")
        normalized = {key: str(resource.get(key) or "") for key in ("kind", "name", "id")}
        if normalized["kind"] not in ("network", "volume") or not normalized["name"] or not normalized["id"]:
            raise ValueError("ownership marker contains invalid resource proof")
        key = (normalized["kind"], normalized["name"], normalized["id"])
        if key in seen:
            raise ValueError("ownership marker contains duplicate resource proof")
        seen.add(key)
        resources.append(normalized)
    return expected | {"resources": sorted(resources, key=lambda item: (item["kind"], item["name"], item["id"]))}


def read_ownership_marker(path: Path, expected: dict[str, Any]) -> dict[str, Any]:
    if not path.is_file() or path.is_symlink():
        raise ValueError("durable lab ownership marker is missing")
    marker_stat = path.stat()
    if marker_stat.st_uid != os.getuid():
        raise ValueError("durable lab ownership marker belongs to another user")
    if stat.S_IMODE(marker_stat.st_mode) & 0o077:
        raise ValueError("durable lab ownership marker is not private")
    marker = load_json(path)
    return validate_ownership_marker(marker, expected)


def ensure_ownership_state_dir(path: Path) -> None:
    path.parent.mkdir(parents=True, mode=0o700, exist_ok=True)
    if path.parent.is_symlink() or path.parent.stat().st_uid != os.getuid():
        raise ValueError("durable lab ownership state directory is unsafe")
    path.parent.chmod(0o700)


def write_ownership_marker(path: Path, marker: dict[str, Any]) -> dict[str, Any]:
    expected = expected_ownership()
    normalized = validate_ownership_marker(marker, expected)
    ensure_ownership_state_dir(path)
    if path.exists() or path.is_symlink():
        read_ownership_marker(path, expected)
    temporary = path.with_name(f".{path.name}.{os.getpid()}.{os.urandom(8).hex()}.tmp")
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL
    if hasattr(os, "O_NOFOLLOW"):
        flags |= os.O_NOFOLLOW
    try:
        descriptor = os.open(temporary, flags, 0o600)
        with os.fdopen(descriptor, "w", encoding="utf-8") as fh:
            json.dump(normalized, fh, indent=2, sort_keys=True)
            fh.write("\n")
            fh.flush()
            os.fsync(fh.fileno())
        os.replace(temporary, path)
        path.chmod(0o600)
        directory_fd = os.open(path.parent, os.O_RDONLY)
        try:
            os.fsync(directory_fd)
        finally:
            os.close(directory_fd)
    finally:
        if temporary.exists() or temporary.is_symlink():
            temporary.unlink()
    return normalized


def load_ownership_marker() -> dict[str, Any]:
    expected = expected_ownership()
    path = ownership_marker_path()
    if path.exists() or path.is_symlink():
        return read_ownership_marker(path, expected)
    for legacy_path in legacy_ownership_marker_paths():
        if legacy_path.exists() or legacy_path.is_symlink():
            marker = read_ownership_marker(legacy_path, expected)
            return write_ownership_marker(path, marker)
    raise ValueError("durable lab ownership marker is missing")


def podman_names(command: list[str], fields: tuple[str, ...]) -> set[str]:
    proc = run(command, check=False)
    if proc.returncode != 0:
        raise ValueError("unable to inspect existing Podman ownership inventory")
    if not proc.stdout.strip():
        return set()
    try:
        values = json.loads(proc.stdout)
    except json.JSONDecodeError as exc:
        raise ValueError("Podman returned invalid ownership inventory JSON") from exc
    names: set[str] = set()
    for value in values if isinstance(values, list) else []:
        for field in fields:
            raw = value.get(field)
            if isinstance(raw, list):
                names.update(str(item) for item in raw if item)
            elif raw:
                names.add(str(raw))
    return names


def claim_ownership() -> dict[str, Any]:
    expected = expected_ownership()
    path = ownership_marker_path()
    if path.exists() or path.is_symlink() or any(item.exists() or item.is_symlink() for item in legacy_ownership_marker_paths()):
        return validate_ownership_marker(load_ownership_marker(), expected)
    podman = shutil.which("podman")
    existing_requested_resources = False
    if podman:
        machine_names = podman_names([podman, "machine", "list", "--format", "json"], ("Name", "name"))
        connection_names = podman_names([podman, "system", "connection", "list", "--format", "json"], ("Name", "name"))
        existing_requested_resources = expected["machine"] in machine_names or expected["connection"] in connection_names
    if existing_requested_resources:
        raise ValueError("refusing to adopt existing Podman resources without durable task ownership evidence; use explicit ownership recovery")
    ensure_dirs()
    return write_ownership_marker(path, expected)


def iter_string_values(value: Any, path: str = "$") -> Iterable[tuple[str, str]]:
    if isinstance(value, str):
        yield path, value
    elif isinstance(value, dict):
        for key, child in value.items():
            yield from iter_string_values(child, f"{path}.{key}")
    elif isinstance(value, list):
        for index, child in enumerate(value):
            yield from iter_string_values(child, f"{path}[{index}]")


def normalize_phone_candidate(value: str) -> str | None:
    raw = value.strip()
    digits = "".join(str(unicodedata.decimal(character)) for character in raw if character.isdecimal())
    if raw.startswith("+91") and digits.startswith("91"):
        digits = digits[2:]
    elif raw.startswith("0091") and digits.startswith("0091"):
        digits = digits[4:]
    elif len(digits) in (12, 13) and digits.startswith("91"):
        digits = digits[2:]
    if len(digits) == 11 and digits.startswith("0"):
        digits = digits[1:]
    if len(digits) == 10 and digits[0] in "123456789":
        return digits
    return None


def normalize_email_candidate(value: str) -> str:
    return value.rstrip(".,!?;:").lower()


def contact_failures(value: Any) -> list[str]:
    failures: list[str] = []
    for path, text in iter_string_values(value):
        if any(normalize_phone_candidate(match.group(0)) for match in PHONE_CANDIDATE_RE.finditer(text)):
            failures.append(f"{path} contains a real-looking Indian phone number")
        for match in EMAIL_CANDIDATE_RE.finditer(text):
            domain = normalize_email_candidate(match.group(1)).rsplit("@", 1)[1]
            if not domain.endswith(".invalid"):
                failures.append(f"{path} contains an email outside the approved .invalid domain")
    return failures


def run(cmd: list[str], cwd: Path | None = None, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, cwd=cwd, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=check)


def ensure_dirs() -> None:
    home = runtime_home()
    if home.exists() and home.stat().st_uid != os.getuid():
        raise SystemExit(f"refusing to use lab home owned by another user: {home}")
    for directory in (home, home / "bin"):
        directory.mkdir(parents=True, exist_ok=True)
        directory.chmod(0o700)


def parse_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key] = value.strip().strip("'").strip('"')
    return values


def write_env(upstream_env: Path, output_env: Path) -> dict[str, str]:
    text = upstream_env.read_text(encoding="utf-8")
    replacements = {
        "COMPOSE_PROFILES=emr": "COMPOSE_PROFILES=bahmni-standard",
        "TZ=UTC": f"TZ={env_value('BAHMNI_LAB_TZ', 'UTC')}",
    }
    for old, new in replacements.items():
        text = text.replace(old, new)

    memory = env_value("BAHMNI_LAB_OPENMRS_JAVA_MEMORY_OPTS", "-Xms512m -Xmx1536m")
    if memory:
        text = re.sub(r"^OMRS_JAVA_MEMORY_OPTS=.*$", f"OMRS_JAVA_MEMORY_OPTS='{memory}'", text, flags=re.MULTILINE)

    # Keep official app credentials inside this local-only runtime file; never commit it.
    text += "\n# fm-bahmni-lab-r1 local Podman adaptation\n"
    text += f"BAHMNI_LAB_HTTP_PORT={port('HTTP', 18080)}\n"
    text += f"BAHMNI_LAB_HTTPS_PORT={port('HTTPS', 18443)}\n"
    text += f"BAHMNI_LAB_ODOO_PORT={port('ODOO', 18069)}\n"
    text += f"BAHMNI_LAB_DCM4CHEE_HTTP_PORT={port('DCM4CHEE_HTTP', 18055)}\n"
    text += f"BAHMNI_LAB_DCM4CHEE_DICOM_PORT={port('DCM4CHEE_DICOM', 11112)}\n"
    output_env.write_text(text, encoding="utf-8")
    output_env.chmod(0o600)
    return parse_env_file(output_env)


def add_platform_after_service(text: str, service: str, platform: str) -> str:
    marker = f"  {service}:\n"
    # Only a top-level service key counts; the same key nested under depends_on
    # or another mapping is more deeply indented and must not be matched.
    if text.startswith(marker):
        start = 0
    else:
        anchored = text.find("\n" + marker)
        if anchored < 0:
            return text
        start = anchored + 1
    next_service = re.search(r"\n  [A-Za-z0-9_-]+:\n", text[start + len(marker):])
    end = start + len(marker) + next_service.start() if next_service else len(text)
    block = text[start:end]
    if "platform:" in block:
        return text
    return text[:start] + marker + f"    platform: {platform}\n" + text[start + len(marker):]


def validate_loopback_ports(compose_text: str) -> None:
    in_ports = False
    ports_indent = 0
    bad_ports: list[str] = []
    for raw in compose_text.splitlines():
        stripped = raw.strip()
        indent = len(raw) - len(raw.lstrip(" "))
        if stripped == "ports:":
            in_ports = True
            ports_indent = indent
            continue
        if in_ports and stripped and indent <= ports_indent and not stripped.startswith("-"):
            in_ports = False
        if not in_ports or not stripped.startswith("-"):
            continue
        value = stripped[1:].strip().strip("'").strip('"')
        if value and not value.startswith("127.0.0.1:"):
            bad_ports.append(value)
    if bad_ports:
        raise SystemExit("generated compose contains non-loopback published ports: " + ", ".join(bad_ports))


CONTAINER_NAME_RE = re.compile(r"(?m)^([ \t]*)container_name:[ \t]*(.+?)[ \t]*$")


def scope_container_names(text: str) -> str:
    prefix = f"{LAB_ID}-"

    def rewrite(match: re.Match[str]) -> str:
        indent, raw = match.group(1), match.group(2)
        name = raw.strip().strip("'").strip('"')
        if not name or name.startswith(prefix):
            return match.group(0)
        return f"{indent}container_name: {prefix}{name}"

    return CONTAINER_NAME_RE.sub(rewrite, text)


def validate_container_names(compose_text: str) -> None:
    prefix = f"{LAB_ID}-"
    unscoped = [
        match.group(2).strip().strip("'").strip('"')
        for match in CONTAINER_NAME_RE.finditer(compose_text)
        if not match.group(2).strip().strip("'").strip('"').startswith(prefix)
    ]
    if unscoped:
        raise SystemExit("generated compose contains container names outside this task: " + ", ".join(unscoped))


def service_blocks(compose_text: str) -> list[tuple[int, int]]:
    return section_blocks(compose_text, "services")


def section_blocks(compose_text: str, section: str) -> list[tuple[int, int]]:
    lines = compose_text.splitlines(keepends=True)
    try:
        section_start = next(index for index, line in enumerate(lines) if line.rstrip("\r\n") == f"{section}:")
    except StopIteration:
        return []
    section_end = len(lines)
    for index in range(section_start + 1, len(lines)):
        line = lines[index]
        if line.strip() and not line.startswith((" ", "\t", "#")):
            section_end = index
            break
    starts = [
        index
        for index in range(section_start + 1, section_end)
        if re.match(r"^  [A-Za-z0-9_.-]+:[^\r\n]*(?:\r?\n)?$", lines[index])
    ]
    return [(start, starts[offset + 1] if offset + 1 < len(starts) else section_end) for offset, start in enumerate(starts)]


def scope_service_labels(compose_text: str) -> str:
    return scope_section_labels(compose_text, "services")


def scope_section_labels(compose_text: str, section: str) -> str:
    lines = compose_text.splitlines(keepends=True)
    for start, end in reversed(section_blocks(compose_text, section)):
        block = "".join(lines[start:end])
        if OWNER_LABEL in block:
            continue
        entry = re.match(r"^  ([A-Za-z0-9_.-]+):([^\r\n]*)(?:\r?\n)?$", lines[start])
        remainder = entry.group(2).strip() if entry else ""
        if remainder and not remainder.startswith(("#", "{}")):
            raise SystemExit(f"generated compose contains unsupported inline {section} ownership entry")
        inline_empty = re.match(r"^(  [A-Za-z0-9_.-]+:)\s*\{\}\s*(?:#.*)?(?:\r?\n)?$", lines[start])
        if inline_empty:
            lines[start] = inline_empty.group(1) + "\n"
        labels_index = next((index for index in range(start + 1, end) if lines[index].rstrip("\r\n") == "    labels:"), None)
        if labels_index is None:
            lines.insert(start + 1, f"    labels:\n      {OWNER_LABEL}: {LAB_ID}\n")
            continue
        list_style = any(re.match(r"^      -\s", lines[index]) for index in range(labels_index + 1, end))
        value = f"      - {OWNER_LABEL}={LAB_ID}\n" if list_style else f"      {OWNER_LABEL}: {LAB_ID}\n"
        lines.insert(labels_index + 1, value)
    return "".join(lines)


def validate_service_labels(compose_text: str) -> None:
    validate_section_labels(compose_text, "services", required=True)


def validate_section_labels(compose_text: str, section: str, required: bool = False) -> None:
    blocks = section_blocks(compose_text, section)
    if required and not blocks:
        raise SystemExit(f"generated compose has no {section} to ownership-label")
    unlabeled = [
        compose_text.splitlines()[start].strip().rstrip(":")
        for start, end in blocks
        if OWNER_LABEL not in "".join(compose_text.splitlines(keepends=True)[start:end])
    ]
    if unlabeled:
        raise SystemExit(f"generated compose contains {section} without task ownership labels: " + ", ".join(unlabeled))


def ensure_default_network(compose_text: str) -> str:
    lines = compose_text.splitlines(keepends=True)
    try:
        section_start = next(index for index, line in enumerate(lines) if line.rstrip("\r\n") == "networks:")
    except StopIteration:
        suffix = "" if not compose_text or compose_text.endswith("\n") else "\n"
        return compose_text + suffix + "networks:\n  default:\n"
    if any(lines[start].strip().split(":", 1)[0] == "default" for start, _ in section_blocks(compose_text, "networks")):
        return compose_text
    section_end = len(lines)
    for index in range(section_start + 1, len(lines)):
        if lines[index].strip() and not lines[index].startswith((" ", "\t", "#")):
            section_end = index
            break
    lines.insert(section_end, "  default:\n")
    return "".join(lines)


def scope_compose_ownership(compose_text: str) -> str:
    rendered = scope_section_labels(compose_text, "services")
    rendered = scope_section_labels(rendered, "volumes")
    rendered = ensure_default_network(rendered)
    return scope_section_labels(rendered, "networks")


def validate_compose_ownership(compose_text: str) -> None:
    validate_section_labels(compose_text, "services", required=True)
    validate_section_labels(compose_text, "volumes")
    validate_section_labels(compose_text, "networks", required=True)
    network_names = [compose_text.splitlines()[start].strip().split(":", 1)[0] for start, _ in section_blocks(compose_text, "networks")]
    if "default" not in network_names:
        raise SystemExit("generated compose has no task-owned default network")


def normalized_labels(value: Any) -> dict[str, str]:
    if isinstance(value, dict):
        return {str(key): str(item) for key, item in value.items()}
    if isinstance(value, list):
        labels: dict[str, str] = {}
        for item in value:
            key, separator, raw = str(item).partition("=")
            if separator:
                labels[key] = raw
        return labels
    return {}


def resource_names(value: dict[str, Any]) -> list[str]:
    raw = value.get("Names") or value.get("Name") or value.get("name") or []
    if isinstance(raw, list):
        return [str(item).lstrip("/") for item in raw]
    return [str(raw).lstrip("/")] if raw else []


def resource_identity(value: dict[str, Any]) -> str:
    raw = value.get("Id") or value.get("ID") or value.get("id")
    if raw:
        return str(raw)
    names = resource_names(value)
    attributes = {
        key: value.get(key)
        for key in ("CreatedAt", "Created", "created", "Mountpoint", "mountpoint")
        if value.get(key)
    }
    if not names or not attributes:
        return ""
    material = json.dumps({"names": names, "attributes": attributes}, sort_keys=True, separators=(",", ":"))
    return "sha256:" + hashlib.sha256(material.encode("utf-8")).hexdigest()


def resource_is_candidate(value: dict[str, Any], project: str) -> bool:
    labels = normalized_labels(value.get("Labels") or value.get("labels"))
    project_label = labels.get("com.docker.compose.project") or labels.get("io.podman.compose.project")
    return project_label == project or any(name.startswith((project + "_", LAB_ID + "-")) for name in resource_names(value))


def validate_owned_resources(
    values: Iterable[dict[str, Any]],
    project: str,
    kind: str,
    proofs: Iterable[dict[str, str]] = (),
) -> None:
    proof_keys = {(item.get("kind"), item.get("name"), item.get("id")) for item in proofs}
    for value in values:
        labels = normalized_labels(value.get("Labels") or value.get("labels"))
        project_label = labels.get("com.docker.compose.project") or labels.get("io.podman.compose.project")
        owner_label = labels.get(OWNER_LABEL)
        names = resource_names(value)
        if not resource_is_candidate(value, project):
            continue
        if project_label != project:
            raise ValueError("task-prefixed Podman resource lacks the matching compose project label")
        if owner_label and owner_label != LAB_ID:
            raise ValueError("compose project contains a mismatched task ownership label")
        if owner_label == LAB_ID:
            continue
        if kind == "container":
            raise ValueError("compose project container lacks the explicit task ownership label")
        identity = resource_identity(value)
        if not identity or not any((kind, name, identity) in proof_keys for name in names):
            raise ValueError(f"compose project {kind} lacks an explicit task ownership label or exact durable ownership proof")


def podman_resource_inventory(podman: str) -> dict[str, list[dict[str, Any]]]:
    base = [podman, "--connection", connection_name()]
    commands = {
        "container": base + ["ps", "--all", "--format", "json"],
        "network": base + ["network", "ls", "--format", "json"],
        "volume": base + ["volume", "ls", "--format", "json"],
    }
    inventory: dict[str, list[dict[str, Any]]] = {}
    for kind, command in commands.items():
        proc = run(command, check=False)
        if proc.returncode != 0:
            raise ValueError("unable to inspect Podman ownership labels")
        try:
            values = json.loads(proc.stdout or "[]")
        except json.JSONDecodeError as exc:
            raise ValueError("Podman returned invalid resource-label inventory JSON") from exc
        inventory[kind] = values if isinstance(values, list) else []
    return inventory


def verify_owned_podman_resources() -> None:
    marker = load_ownership_marker()
    podman = shutil.which("podman")
    if not podman:
        raise ValueError("podman not found")
    for kind, values in podman_resource_inventory(podman).items():
        validate_owned_resources(values, project_name(), kind, marker["resources"])


def recover_ownership() -> dict[str, Any]:
    expected = expected_ownership()
    podman = shutil.which("podman")
    if not podman:
        raise ValueError("podman not found")
    machine_names = podman_names([podman, "machine", "list", "--format", "json"], ("Name", "name"))
    connection_names = podman_names([podman, "system", "connection", "list", "--format", "json"], ("Name", "name"))
    if expected["machine"] not in machine_names or expected["connection"] not in connection_names:
        raise ValueError("exact task-owned machine and connection must exist before ownership recovery")
    inventory = podman_resource_inventory(podman)
    validate_owned_resources(inventory["container"], expected["project"], "container")
    owned_containers = [value for value in inventory["container"] if resource_is_candidate(value, expected["project"])]
    resources: list[dict[str, str]] = []
    for kind in ("network", "volume"):
        for value in inventory[kind]:
            if not resource_is_candidate(value, expected["project"]):
                continue
            labels = normalized_labels(value.get("Labels") or value.get("labels"))
            project_label = labels.get("com.docker.compose.project") or labels.get("io.podman.compose.project")
            owner_label = labels.get(OWNER_LABEL)
            if project_label != expected["project"] or (owner_label and owner_label != LAB_ID):
                raise ValueError("refusing ownership recovery for a mismatched compose resource")
            if owner_label != LAB_ID and not owned_containers:
                raise ValueError("refusing ownership recovery for unlabeled resources without an explicitly owned container")
            if owner_label == LAB_ID:
                continue
            identity = resource_identity(value)
            names = resource_names(value)
            if not identity or len(names) != 1:
                raise ValueError("refusing ownership recovery for a resource without exact identity")
            resources.append({"kind": kind, "name": names[0], "id": identity})
    marker = expected | {"resources": resources}
    for kind, values in inventory.items():
        validate_owned_resources(values, expected["project"], kind, marker["resources"])
    return write_ownership_marker(ownership_marker_path(), marker)


def forget_ownership() -> None:
    marker = load_ownership_marker()
    podman = shutil.which("podman")
    if not podman:
        raise ValueError("cannot prove task-owned Podman resources are absent")
    machine_names = podman_names([podman, "machine", "list", "--format", "json"], ("Name", "name"))
    connection_names = podman_names([podman, "system", "connection", "list", "--format", "json"], ("Name", "name"))
    if marker["machine"] in machine_names or marker["connection"] in connection_names:
        raise ValueError("task-owned Podman machine or connection still exists")
    expected = expected_ownership()
    for path in (ownership_marker_path(), *legacy_ownership_marker_paths()):
        if path.exists() or path.is_symlink():
            read_ownership_marker(path, expected)
            path.unlink()


def write_compose(upstream_compose: Path, output_compose: Path) -> None:
    pin = load_json(PIN_PATH)
    defaults = pin["default_ports"]
    text = upstream_compose.read_text(encoding="utf-8")
    port_replacements = {
        "      - '80:80'": f"      - '127.0.0.1:{port('HTTP', defaults['http'])}:80'",
        "      - '443:443'": f"      - '127.0.0.1:{port('HTTPS', defaults['https'])}:443'",
        "      - '8069:8069'": f"      - '127.0.0.1:{port('ODOO', defaults['odoo'])}:8069'",
        "      - '8070:8069'": "      - '127.0.0.1:18070:8069'",
        "      - '8055:8055'": f"      - '127.0.0.1:{port('DCM4CHEE_HTTP', defaults['dcm4chee_http'])}:8055'",
        "      - '11112:11112'": f"      - '127.0.0.1:{port('DCM4CHEE_DICOM', defaults['dcm4chee_dicom'])}:11112'",
    }
    for old, new in port_replacements.items():
        text = text.replace(old, new)
    # Any explicit container_name bypasses the compose project prefix, so scope
    # every one of them to this lab before the stack can adopt a foreign container.
    text = scope_container_names(text)
    text = scope_compose_ownership(text)
    # Official odoo-16 image for the pinned Standard tag is amd64-only as of the
    # research run. Platform pin lets rootless Podman on Apple Silicon use qemu.
    text = add_platform_after_service(text, "odoo", "linux/amd64")
    validate_loopback_ports(text)
    validate_container_names(text)
    validate_compose_ownership(text)
    output_compose.write_text(text, encoding="utf-8")


def write_podman_wrapper() -> None:
    ensure_dirs()
    wrapper = podman_wrapper_path()
    wrapper.write_text(
        "#!/usr/bin/env bash\n"
        "set -euo pipefail\n"
        f"exec podman --connection {connection_name()!r} \"$@\"\n",
        encoding="utf-8",
    )
    wrapper.chmod(0o755)


def write_credentials(values: dict[str, str]) -> None:
    creds = {
        "warning": "Local-only credentials generated/copied for fm-bahmni-lab-r1. Do not commit or paste into status messages.",
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "urls": local_urls(),
        "openmrs": {
            "username": values.get("OPENMRS_ATOMFEED_USER", "admin"),
            "password": values.get("OPENMRS_ATOMFEED_PASSWORD", ""),
        },
        "odoo": {
            "url": local_urls()["odoo"],
            "database": values.get("ODOO_DB_NAME", "odoo"),
            "username_hint": values.get("ODOO_ATOMFEED_USER", ""),
            "password_hint": values.get("ODOO_ATOMFEED_PASSWORD", ""),
        },
        "notes": [
            "Use bin/bahmni-lab credentials --show only in a private terminal.",
            "The seed scripts read this local runtime file or the generated .env.fm; neither is committed.",
        ],
    }
    path = credentials_path()
    path.write_text(json.dumps(creds, indent=2), encoding="utf-8")
    path.chmod(stat.S_IRUSR | stat.S_IWUSR)


def local_urls() -> dict[str, str]:
    https = port("HTTPS", 18443)
    http = port("HTTP", 18080)
    return {
        "bahmni": f"https://127.0.0.1:{https}/bahmni/home/index.html",
        "openmrs_rest": f"https://127.0.0.1:{https}/openmrs/ws/rest/v1",
        "openmrs_fhir": f"https://127.0.0.1:{https}/openmrs/ws/fhir2/R4",
        "openmrs_legacy_http_redirect": f"http://127.0.0.1:{http}/openmrs",
        "odoo": f"http://127.0.0.1:{port('ODOO', 18069)}",
        "dcm4chee": f"http://127.0.0.1:{port('DCM4CHEE_HTTP', 18055)}",
        "dcm4chee_dicom": f"127.0.0.1:{port('DCM4CHEE_DICOM', 11112)}",
    }


def prepare(args: argparse.Namespace) -> None:
    pin = load_json(PIN_PATH)
    home = runtime_home()
    if args.dry_run:
        dump_json({
            "action": "prepare",
            "lab_home": str(home),
            "repository": pin["official_repository"],
            "commit": pin["official_commit"],
            "compose": str(generated_compose_path()),
            "env": str(generated_env_path()),
            "credentials": str(credentials_path()),
            "loopback_urls": local_urls(),
        })
        return

    ensure_dirs()
    if not repo_dir().exists():
        run(["git", "clone", pin["official_repository"], str(repo_dir())], check=True)
    run(["git", "fetch", "--tags", "--force", "origin"], cwd=repo_dir(), check=True)
    run(["git", "checkout", "--detach", pin["official_commit"]], cwd=repo_dir(), check=True)
    head = run(["git", "rev-parse", "HEAD"], cwd=repo_dir(), check=True).stdout.strip()
    if head != pin["official_commit"]:
        raise SystemExit(f"unexpected Bahmni checkout HEAD {head}")

    for name in [
        "extra-odoo-addons",
        "extra-odoo-10-addons",
        "openelis-db-dump",
        "odoo-db-dump",
        "openmrs-uploads",
        "restore-artifacts",
    ]:
        (standard_dir() / name).mkdir(parents=True, exist_ok=True)

    values = write_env(standard_dir() / ".env", generated_env_path())
    write_compose(standard_dir() / "docker-compose.yml", generated_compose_path())
    write_podman_wrapper()
    write_credentials(values)
    dump_json({
        "prepared": True,
        "lab_home": str(home),
        "head": head,
        "compose": str(generated_compose_path()),
        "env": str(generated_env_path()),
        "credentials_path": str(credentials_path()),
        "urls": local_urls(),
    })


def read_credentials() -> dict[str, Any]:
    if credentials_path().exists():
        return load_json(credentials_path())
    values = parse_env_file(generated_env_path())
    return {
        "openmrs": {
            "username": values.get("OPENMRS_ATOMFEED_USER", "admin"),
            "password": values.get("OPENMRS_ATOMFEED_PASSWORD", ""),
        },
        "urls": local_urls(),
    }


def credentials(args: argparse.Namespace) -> None:
    creds = read_credentials()
    if args.path:
        print(credentials_path())
        return
    if args.show:
        dump_json(creds)
        return
    safe = dict(creds)
    if "openmrs" in safe:
        safe["openmrs"] = {"username": safe["openmrs"].get("username"), "password": "<redacted>"}
    if "odoo" in safe:
        safe["odoo"] = {k: ("<redacted>" if "password" in k else v) for k, v in safe["odoo"].items()}
    dump_json(safe)


def normalize_name(value: str) -> str:
    return re.sub(r"[^a-z]", "", value.lower())


def provider_records(seed: dict[str, Any]) -> list[dict[str, Any]]:
    """Return every synthetic staff/provider record represented in OpenMRS as a Provider."""
    return list(seed.get("providers", [])) + list(seed.get("staff", []))


def clinical_text_values(seed: dict[str, Any]) -> Iterable[tuple[str, str]]:
    """Return patient-record-facing strings that should read like fictional clinical records."""
    records = seed.get("records", {})
    for key in ("appointment_comment_prefix", "encounter_marker_prefix", "allergy_comment", "document_note"):
        value = records.get(key)
        if value:
            yield f"records.{key}", str(value)
    for index, value in enumerate(records.get("condition_texts", [])):
        yield f"records.condition_texts[{index}]", str(value)

    for patient in seed.get("patients", []):
        identifier = patient.get("identifier", "unknown")
        if patient.get("chief_complaint"):
            yield f"patients.{identifier}.chief_complaint", str(patient["chief_complaint"])
        for index, value in enumerate(patient.get("diagnosis_notes", [])):
            yield f"patients.{identifier}.diagnosis_notes[{index}]", str(value)
        for event_index, event in enumerate(patient.get("history_events", [])):
            event_id = event.get("event_id", event_index)
            if event.get("summary"):
                yield f"patients.{identifier}.history_events.{event_id}.summary", str(event["summary"])
            for field in ("diagnoses", "tests", "medications", "procedures", "notes"):
                for index, value in enumerate(event.get(field, [])):
                    yield f"patients.{identifier}.history_events.{event_id}.{field}[{index}]", str(value)
            billing = event.get("billing") or {}
            for field in ("invoice_id", "status", "payer"):
                if billing.get(field):
                    yield f"patients.{identifier}.history_events.{event_id}.billing.{field}", str(billing[field])


def check_synthetic(args: argparse.Namespace) -> dict[str, Any]:
    seed = load_json(SEED_PATH)
    failures: list[str] = []
    if not seed.get("synthetic"):
        failures.append("dataset synthetic flag is false")
    if seed.get("organization", {}).get("name") != "Chikitsalayaḥ":
        failures.append("synthetic hospital name must be exactly Chikitsalayaḥ")
    failures.extend(contact_failures(seed))
    all_providers = provider_records(seed)
    if "SPARSH" in json.dumps(all_providers, ensure_ascii=False):
        failures.append("staff/provider fixture contains SPARSH")
    for provider in all_providers:
        ident = provider.get("identifier", "")
        if not ident.startswith(("SYN-PROV-", "SYN-STAFF-")):
            failures.append(f"staff/provider identifier not synthetic: {ident}")
        full = f"{provider.get('given_name', '')} {provider.get('family_name', '')}"
        norm = normalize_name(full)
        if any(token in norm for token in ["doctor", "sparsh", "hospital", "praveen", "abhijit"]):
            failures.append(f"staff/provider {ident} contains a forbidden identity token")
    for patient in seed.get("patients", []):
        ident = patient.get("identifier", "")
        if not ident.startswith("SYN-HEN-"):
            failures.append(f"patient identifier not synthetic: {ident}")
        contact = patient.get("synthetic_contact", {})
        if contact:
            phone = contact.get("phone", "")
            email = contact.get("email", "")
            if phone != seed.get("contact_defaults", {}).get("patient_phone_placeholder", "000-000-0000"):
                failures.append(f"patient {ident} contact phone is not the invalid synthetic placeholder")
            if email and not email.endswith(".invalid"):
                failures.append(f"patient {ident} contact email must use the .invalid TLD")

    clinical_texts_checked = 0
    for path, value in clinical_text_values(seed):
        clinical_texts_checked += 1
        if "synthetic" in value.lower():
            failures.append(f"clinical-facing fixture text {path} contains 'synthetic'; keep safeguards in identifiers/metadata")

    online_checked = False
    if args.online_source:
        online_checked = True
        source = seed["source_taxonomy"]["source_url"]
        try:
            html = urllib.request.urlopen(source, timeout=20).read().decode("utf-8", "replace")
            # Compare normalized synthetic provider names against live page text without
            # printing or storing real names.
            page_norm = normalize_name(html)
            for provider in all_providers:
                full_norm = normalize_name(f"{provider['given_name']} {provider['family_name']}")
                if full_norm and full_norm in page_norm:
                    failures.append(f"staff/provider {provider['identifier']} collides with online source text")
        except Exception as exc:  # pragma: no cover - network optional
            failures.append(f"online source check failed: {type(exc).__name__}")

    result = {
        "ok": not failures,
        "dataset_id": seed.get("dataset_id"),
        "providers_checked": len(all_providers),
        "clinical_providers_checked": len(seed.get("providers", [])),
        "staff_checked": len(seed.get("staff", [])),
        "patients_checked": len(seed.get("patients", [])),
        "clinical_texts_checked": clinical_texts_checked,
        "online_source_checked": online_checked,
        "failures": failures,
    }
    if not getattr(args, "quiet", False):
        if args.json:
            dump_json(result)
        else:
            print("synthetic check:", "ok" if result["ok"] else "failed")
            for failure in failures:
                print("-", failure)
    if failures:
        raise SystemExit(1)
    return result


class ApiError(RuntimeError):
    pass


REST_SUFFIX = "/ws/rest/v1"
FHIR_SUFFIX = "/ws/fhir2/R4"
OPENMRS_REST_PATH = "/openmrs" + REST_SUFFIX
LOOPBACK_API_HOSTS = {"127.0.0.1", "localhost", "::1"}


def derive_fhir_base(rest_base: str) -> str:
    trimmed = rest_base.rstrip("/")
    if trimmed.endswith(REST_SUFFIX):
        trimmed = trimmed[: -len(REST_SUFFIX)]
    return trimmed + FHIR_SUFFIX


def configured_https_port() -> int:
    raw = port("HTTPS", 18443)
    try:
        value = int(raw)
    except ValueError as exc:
        raise ApiError("configured Bahmni HTTPS port is invalid") from exc
    if not 1 <= value <= 65535 or str(value) != raw:
        raise ApiError("configured Bahmni HTTPS port is invalid")
    return value


def validate_rest_base(value: str) -> str:
    try:
        parsed = urllib.parse.urlsplit(value)
        host = parsed.hostname
        target_port = parsed.port
    except (TypeError, ValueError) as exc:
        raise ApiError("OpenMRS REST base URL is invalid") from exc
    if parsed.scheme != "https":
        raise ApiError("OpenMRS REST base URL must use HTTPS")
    if parsed.username is not None or parsed.password is not None:
        raise ApiError("OpenMRS REST base URL must not contain user information")
    if host not in LOOPBACK_API_HOSTS:
        raise ApiError("OpenMRS REST base URL must use the task loopback host")
    if target_port != configured_https_port():
        raise ApiError("OpenMRS REST base URL must use the configured task HTTPS port")
    if parsed.path != OPENMRS_REST_PATH or parsed.query or parsed.fragment:
        raise ApiError("OpenMRS REST base URL must use the exact task REST path")
    return urllib.parse.urlunsplit(("https", parsed.netloc, OPENMRS_REST_PATH, "", ""))


class RejectAuthenticatedRedirects(urllib.request.HTTPRedirectHandler):
    def redirect_request(
        self,
        req: urllib.request.Request,
        fp: Any,
        code: int,
        msg: str,
        headers: Any,
        newurl: str,
    ) -> None:
        return None


class BahmniApi:
    def __init__(self, base_https: str | None = None):
        if base_https is not None:
            self.rest_base = validate_rest_base(base_https)
            creds = read_credentials()
        else:
            creds = read_credentials()
            urls = creds.get("urls") if isinstance(creds.get("urls"), dict) else local_urls()
            self.rest_base = validate_rest_base(urls.get("openmrs_rest", local_urls()["openmrs_rest"]))
        self.fhir_base = derive_fhir_base(self.rest_base)
        self.allowed_bases = {self.rest_base, self.fhir_base}
        user = creds.get("openmrs", {}).get("username") or "admin"
        password = creds.get("openmrs", {}).get("password") or parse_env_file(generated_env_path()).get("OPENMRS_ATOMFEED_PASSWORD", "")
        token = base64.b64encode(f"{user}:{password}".encode("utf-8")).decode("ascii")
        self.auth_header = "Basic " + token
        self.ctx = ssl._create_unverified_context()
        self.proxy_handler = urllib.request.ProxyHandler({})
        self.opener = urllib.request.build_opener(
            self.proxy_handler,
            urllib.request.HTTPSHandler(context=self.ctx),
            RejectAuthenticatedRedirects(),
        )

    def request(self, base: str, path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None, fhir: bool = False) -> tuple[int, Any]:
        normalized_base = base.rstrip("/")
        if normalized_base not in self.allowed_bases:
            raise ApiError("refusing authenticated request outside the task loopback API")
        url = normalized_base + path
        if params:
            clean = {k: v for k, v in params.items() if v is not None}
            url += "?" + urllib.parse.urlencode(clean)
        headers = {"Authorization": self.auth_header, "Accept": "application/fhir+json" if fhir else "application/json"}
        data = None
        if body is not None:
            failures = contact_failures(body)
            if failures:
                raise ApiError("refusing API write containing realistic contact data: " + "; ".join(failures))
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/fhir+json" if fhir else "application/json"
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with self.opener.open(req, timeout=45) as resp:
                raw = resp.read().decode("utf-8", "replace")
                if not raw:
                    return resp.status, {}
                return resp.status, json.loads(raw)
        except urllib.error.HTTPError as exc:
            if 300 <= exc.code < 400:
                raise ApiError(f"refusing authenticated redirect for {method} {path}") from exc
            raw = exc.read().decode("utf-8", "replace")
            raise ApiError(f"{method} {path} HTTP {exc.code}: {raw[:800]}") from exc

    def rest(self, path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None) -> tuple[int, Any]:
        return self.request(self.rest_base, path, method, body, params, fhir=False)

    def fhir(self, path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None) -> tuple[int, Any]:
        return self.request(self.fhir_base, path, method, body, params, fhir=True)

    def results(self, path: str, **params: Any) -> list[dict[str, Any]]:
        _, data = self.rest(path, params=params)
        if isinstance(data, list):
            return data
        return data.get("results", [])

    def first_by_display(self, path: str, display: str) -> dict[str, Any]:
        for item in self.results(path, v="default", limit=200):
            if (item.get("display") or item.get("name")) == display:
                return item
        raise ApiError(f"could not find {display!r} at {path}")


def exact_search(api: BahmniApi, path: str, query: str) -> dict[str, Any] | None:
    for item in api.results(path, q=query, v="default", limit=100):
        if (item.get("display") or item.get("name")) == query:
            return item
    return None


def record_identifiers(item: dict[str, Any]) -> set[str]:
    identifiers: set[str] = set()
    direct = item.get("identifier")
    if isinstance(direct, str):
        identifiers.add(direct)
    for value in item.get("identifiers") or []:
        if isinstance(value, dict) and isinstance(value.get("identifier"), str):
            identifiers.add(value["identifier"])
    display = item.get("display")
    if isinstance(display, str) and " - " in display:
        identifiers.add(display.split(" - ", 1)[0])
    return identifiers


def exact_owned_record(items: Iterable[dict[str, Any]], identifier: str) -> dict[str, Any] | None:
    matches = [item for item in items if identifier in record_identifiers(item)]
    if len(matches) > 1:
        raise ApiError(f"multiple records claim task identifier {identifier}")
    return matches[0] if matches else None


def expected_record_display(record: dict[str, Any]) -> str:
    return f"{record['identifier']} - {record['given_name']} {record['family_name']}"


def exact_display_matches(item: dict[str, Any], record: dict[str, Any]) -> bool:
    return item.get("display") == expected_record_display(record)


def reconcile_person_name(api: BahmniApi, person: dict[str, Any], given_name: str, family_name: str) -> bool:
    person_uuid = person.get("uuid")
    if not person_uuid:
        raise ApiError("owned record does not expose a Person UUID")
    details = person
    preferred = details.get("preferredName")
    if not isinstance(preferred, dict) or not preferred.get("uuid"):
        _, details = api.rest(f"/person/{person_uuid}", params={"v": "full"})
        preferred = details.get("preferredName")
        if not isinstance(preferred, dict) or not preferred.get("uuid"):
            preferred = next((name for name in details.get("names") or [] if name.get("preferred") and name.get("uuid")), None)
    if not isinstance(preferred, dict) or not preferred.get("uuid"):
        raise ApiError("owned Person record does not expose a preferred name UUID")
    if preferred.get("givenName") == given_name and preferred.get("familyName") == family_name:
        return False
    api.rest(
        f"/person/{person_uuid}/name/{preferred['uuid']}",
        "POST",
        {"givenName": given_name, "familyName": family_name, "preferred": True},
    )
    return True


def ensure_location(api: BahmniApi, name: str, parent: str | None, summary: dict[str, int]) -> str:
    existing = exact_search(api, "/location", name)
    if existing:
        summary["locations_existing"] += 1
        return existing["uuid"]
    body: dict[str, Any] = {
        "name": name,
        "description": "FM-BAHMNI-LAB-R1 local connector lab location for Chikitsalayaḥ.",
    }
    if parent:
        body["parentLocation"] = parent
    _, created = api.rest("/location", "POST", body)
    summary["locations_created"] += 1
    return created["uuid"]


def ensure_provider(api: BahmniApi, provider: dict[str, str], summary: dict[str, int]) -> str:
    ident = provider["identifier"]
    item = exact_owned_record(api.results("/provider", q=ident, v="full", limit=20), ident)
    if item:
        person = item.get("person") or {}
        if reconcile_person_name(api, person, provider["given_name"], provider["family_name"]):
            summary["providers_reconciled"] += 1
        summary["providers_existing"] += 1
        return item["uuid"]
    person_body = {
        "names": [{"givenName": provider["given_name"], "familyName": provider["family_name"]}],
        "gender": provider["gender"],
        "birthdate": provider["birthdate"],
    }
    _, person = api.rest("/person", "POST", person_body)
    _, created = api.rest("/provider", "POST", {"identifier": ident, "person": person["uuid"]})
    summary["providers_created"] += 1
    return created["uuid"]


def ensure_patient(api: BahmniApi, patient: dict[str, str], pid_type: str, location: str, summary: dict[str, int]) -> str:
    ident = patient["identifier"]
    item = exact_owned_record(api.results("/patient", q=ident, v="full", limit=20), ident)
    if item:
        person = item.get("person") or {"uuid": item["uuid"]}
        if reconcile_person_name(api, person, patient["given_name"], patient["family_name"]):
            summary["patients_reconciled"] += 1
        summary["patients_existing"] += 1
        return item["uuid"]
    person = {
        "names": [{"givenName": patient["given_name"], "familyName": patient["family_name"]}],
        "gender": patient["gender"],
        "birthdate": patient["birthdate"],
        "addresses": [{
            "address1": f"{ident} Loopback Lane",
            "cityVillage": "Demo City",
            "stateProvince": "Local Test State",
            "country": "ZZ",
            "postalCode": "00000",
        }],
    }
    body = {
        "person": person,
        "identifiers": [{"identifier": ident, "identifierType": pid_type, "location": location, "preferred": True}],
    }
    _, created = api.rest("/patient", "POST", body)
    summary["patients_created"] += 1
    return created["uuid"]


def appointment_ms(value: str) -> int:
    normalized = value.replace("Z", "+00:00")
    normalized = re.sub(r"([+-]\d{2})(\d{2})$", r"\1:\2", normalized)
    parsed = dt.datetime.fromisoformat(normalized)
    return int(parsed.timestamp() * 1000)


def stable_record_marker(kind: str, *parts: str) -> str:
    encoded = [urllib.parse.quote(str(part), safe="-._~") for part in (kind, *parts)]
    return STABLE_RECORD_PREFIX + "/" + "/".join(encoded)


def marker_in_obs(encounter: dict[str, Any], marker: str) -> bool:
    def walk(obs_list: list[dict[str, Any]]) -> bool:
        for obs in obs_list:
            value = obs.get("value")
            if isinstance(value, str) and marker in value:
                return True
            group = obs.get("groupMembers") or []
            if group and walk(group):
                return True
        return False
    return walk(encounter.get("obs") or [])


def marker_observations(encounter: dict[str, Any], markers: set[str]) -> list[dict[str, Any]]:
    found: list[dict[str, Any]] = []

    def walk(obs_list: list[dict[str, Any]]) -> None:
        for obs in obs_list:
            if obs.get("value") in markers:
                found.append(obs)
            walk(obs.get("groupMembers") or [])

    walk(encounter.get("obs") or [])
    return found


def ensure_visit(api: BahmniApi, patient_uuid: str, visit_type: str, start: str, location: str, summary: dict[str, int], stop: str | None = None) -> str:
    visits = api.results("/visit", patient=patient_uuid, v="full", limit=100)
    for visit in visits:
        if str(visit.get("startDatetime", "")) == start and visit.get("visitType", {}).get("uuid") == visit_type:
            if stop and not visit.get("stopDatetime"):
                api.rest(f"/visit/{visit['uuid']}", "POST", {"stopDatetime": stop})
            summary["visits_existing"] += 1
            return visit["uuid"]
    body = {
        "patient": patient_uuid,
        "visitType": visit_type,
        "startDatetime": start,
        "location": location,
    }
    if stop:
        body["stopDatetime"] = stop
    _, created = api.rest("/visit", "POST", body)
    summary["visits_created"] += 1
    return created["uuid"]


def ensure_encounter(
    api: BahmniApi,
    patient_uuid: str,
    encounter_type: str,
    when: str,
    location: str,
    visit_uuid: str,
    obs: list[dict[str, Any]],
    marker: str,
    document_concept: str,
    summary: dict[str, int],
    legacy_markers: Iterable[str] = (),
) -> str:
    encounters = api.results("/encounter", patient=patient_uuid, v="full", limit=100)
    markers = {marker, *legacy_markers}
    candidates = [
        (encounter, marker_observations(encounter, markers))
        for encounter in encounters
    ]
    candidates = [(encounter, observations) for encounter, observations in candidates if observations]
    if candidates:
        candidates.sort(key=lambda value: (not any(obs.get("value") == marker for obs in value[1]), str(value[0].get("uuid", ""))))
        canonical_encounter = candidates[0][0]
        changed = False
        for encounter_index, (encounter, observations) in enumerate(candidates):
            observations.sort(key=lambda obs: (obs.get("value") != marker, str(obs.get("uuid", ""))))
            for obs_index, observation in enumerate(observations):
                observation_uuid = observation.get("uuid")
                if not observation_uuid:
                    raise ApiError("owned encounter marker does not expose an observation UUID")
                if encounter_index == 0 and obs_index == 0:
                    desired = marker
                elif obs_index == 0:
                    desired = f"{marker}:preserved:{encounter['uuid']}"
                else:
                    desired = f"{marker}:preserved-observation:{observation_uuid}"
                if observation.get("value") != desired:
                    api.rest(f"/obs/{observation_uuid}", "POST", {"value": desired})
                    changed = True
        if changed:
            summary["encounters_reconciled"] += 1
        summary["encounter_duplicates_preserved"] += max(0, len(candidates) - 1)
        summary["encounters_existing"] += 1
        return canonical_encounter["uuid"]
    full_obs = list(obs) + [{"concept": document_concept, "value": marker, "obsDatetime": when}]
    _, created = api.rest("/encounter", "POST", {
        "patient": patient_uuid,
        "encounterType": encounter_type,
        "encounterDatetime": when,
        "location": location,
        "visit": visit_uuid,
        "obs": full_obs,
    })
    summary["encounters_created"] += 1
    return created["uuid"]


def appointment_comment(comment: str, marker: str) -> str:
    return f"{comment}\n[{marker}]"


def reference_uuid(value: Any) -> str | None:
    if isinstance(value, dict):
        uuid = value.get("uuid")
        return str(uuid) if uuid else None
    return str(value) if value else None


def appointment_update_body(appointment: dict[str, Any], comments: str) -> dict[str, Any]:
    required = {
        "uuid": appointment.get("uuid"),
        "patientUuid": reference_uuid(appointment.get("patient")),
        "serviceUuid": reference_uuid(appointment.get("service")),
        "locationUuid": reference_uuid(appointment.get("location")),
        "appointmentKind": appointment.get("appointmentKind"),
        "status": appointment.get("status"),
        "startDateTime": appointment.get("startDateTime"),
        "endDateTime": appointment.get("endDateTime"),
    }
    missing = [key for key, value in required.items() if value is None or value == ""]
    providers_value = appointment.get("providers")
    if not isinstance(providers_value, list):
        missing.append("providers")
    if missing:
        raise ApiError("owned appointment cannot be safely updated; missing " + ", ".join(sorted(set(missing))))
    if providers_value:
        raise ApiError("owned appointment cannot be safely updated; provider associations are not losslessly representable")
    body = required | {"providers": [], "comments": comments}
    service_type_uuid = reference_uuid(appointment.get("serviceType"))
    if service_type_uuid:
        body["serviceTypeUuid"] = service_type_uuid
    return body


def ensure_appointment(
    api: BahmniApi,
    patient_ident: str,
    patient_uuid: str,
    provider_uuid: str,
    service_uuid: str,
    location: str,
    start: str,
    end: str,
    comment: str,
    marker: str,
    summary: dict[str, int],
    legacy_comments: Iterable[str] = (),
) -> None:
    _, appointments = api.rest("/appointments")
    start_ms = appointment_ms(start.replace(".000+0000", "+00:00"))
    accepted_comments = {comment, *legacy_comments}
    candidates: list[dict[str, Any]] = []
    if isinstance(appointments, list):
        for appt in appointments:
            comments = str(appt.get("comments") or "")
            owned_comment = marker in comments or comments in accepted_comments
            if appt.get("patient", {}).get("identifier") == patient_ident and appt.get("startDateTime") == start_ms and owned_comment:
                candidates.append(appt)
    body: dict[str, Any] = {
        "patientUuid": patient_uuid,
        "serviceUuid": service_uuid,
        "appointmentKind": "Scheduled",
        "status": "Scheduled",
        "startDateTime": start,
        "endDateTime": end,
        "providers": [{"uuid": provider_uuid, "response": "ACCEPTED"}],
        "locationUuid": location,
    }
    if candidates:
        canonical_token = f"[{marker}]"
        candidates.sort(key=lambda appt: (canonical_token not in str(appt.get("comments") or ""), str(appt.get("uuid", ""))))
        updates: list[dict[str, Any]] = []
        for index, appt in enumerate(candidates):
            appointment_uuid = appt.get("uuid")
            if not appointment_uuid:
                raise ApiError("owned appointment does not expose a UUID")
            desired_marker = marker if index == 0 else f"{marker}:preserved:{appointment_uuid}"
            desired_comment = appointment_comment(comment, desired_marker)
            if appt.get("comments") != desired_comment:
                updates.append(appointment_update_body(appt, desired_comment))
        for update in updates:
            api.rest("/appointment", "POST", update)
        if updates:
            summary["appointments_reconciled"] += 1
        summary["appointment_duplicates_preserved"] += max(0, len(candidates) - 1)
        summary["appointments_existing"] += 1
        return
    body["comments"] = appointment_comment(comment, marker)
    api.rest("/appointments", "POST", body)
    summary["appointments_created"] += 1


def ensure_allergy(
    api: BahmniApi,
    patient_uuid: str,
    concept_uuid: str,
    comment: str,
    summary: dict[str, int],
    legacy_comments: Iterable[str] = (),
) -> None:
    _, existing = api.rest(f"/patient/{patient_uuid}/allergy", params={"v": "full"})
    allergies = existing.get("results", []) if isinstance(existing, dict) else []
    coded_matches = [
        allergy
        for allergy in allergies
        if allergy.get("allergen", {}).get("codedAllergen", {}).get("uuid") == concept_uuid
    ]
    if any(allergy.get("comment") == comment for allergy in coded_matches):
        summary["allergies_existing"] += 1
        return
    accepted_legacy = set(legacy_comments)
    legacy_matches = [allergy for allergy in coded_matches if allergy.get("comment") in accepted_legacy]
    updates: list[str] = []
    for allergy in legacy_matches:
        allergy_uuid = allergy.get("uuid")
        if not allergy_uuid:
            raise ApiError("owned legacy allergy does not expose a UUID")
        updates.append(str(allergy_uuid))
    for allergy_uuid in updates:
        api.rest(f"/patient/{patient_uuid}/allergy/{allergy_uuid}", "POST", {"comment": comment})
    if updates:
        summary["allergies_reconciled"] += len(updates)
        summary["allergies_existing"] += 1
        return
    if coded_matches:
        summary["allergies_existing"] += 1
        return
    body = {
        "allergen": {"allergenType": "DRUG", "codedAllergen": concept_uuid},
        "severity": "MILD",
        "comment": comment,
    }
    api.rest(f"/patient/{patient_uuid}/allergy", "POST", body)
    summary["allergies_created"] += 1


def fhir_patient_id(api: BahmniApi, patient_identifier: str) -> str | None:
    _, bundle = api.fhir("/Patient", params={"identifier": patient_identifier})
    entries = bundle.get("entry", [])
    if not entries:
        return None
    return entries[0]["resource"]["id"]


def ensure_condition(api: BahmniApi, patient_identifier: str, text: str, summary: dict[str, int]) -> None:
    patient_id = fhir_patient_id(api, patient_identifier)
    if not patient_id:
        summary["conditions_skipped"] += 1
        return
    _, bundle = api.fhir("/Condition", params={"patient": patient_id})
    # The installed OpenMRS FHIR2 condition resource accepts a code/text payload
    # but the pinned Bahmni distro does not echo that code back on read. Treat
    # any existing condition for the synthetic patient as the idempotency guard;
    # diagnosis detail is also recorded in encounter text observations.
    if bundle.get("total", 0) > 0:
        summary["conditions_existing"] += 1
        return
    body = {
        "resourceType": "Condition",
        "clinicalStatus": {"coding": [{"system": "http://terminology.hl7.org/CodeSystem/condition-clinical", "code": "active"}]},
        "verificationStatus": {"coding": [{"system": "http://terminology.hl7.org/CodeSystem/condition-ver-status", "code": "confirmed"}]},
        "code": {"text": text},
        "subject": {"reference": "Patient/" + patient_id},
        "onsetDateTime": "2026-01-15T00:00:00Z",
    }
    api.fhir("/Condition", "POST", body)
    summary["conditions_created"] += 1


def existing_orders(api: BahmniApi, patient_uuid: str) -> list[dict[str, Any]]:
    return api.results("/order", patient=patient_uuid, v="full", limit=100)


def order_exists(orders: list[dict[str, Any]], concept_uuid: str | None = None, drug_uuid: str | None = None, order_type_uuid: str | None = None) -> bool:
    for order in orders:
        if order_type_uuid and order.get("orderType", {}).get("uuid") != order_type_uuid:
            continue
        if concept_uuid and order.get("concept", {}).get("uuid") == concept_uuid:
            return True
        if drug_uuid and order.get("drug", {}).get("uuid") == drug_uuid:
            return True
    return False


def ensure_test_order(api: BahmniApi, orders: list[dict[str, Any]], patient_uuid: str, encounter_uuid: str, provider_uuid: str, concept_uuid: str, order_type_uuid: str, care_setting: str, summary_key: str, summary: dict[str, int]) -> None:
    if order_exists(orders, concept_uuid=concept_uuid, order_type_uuid=order_type_uuid):
        summary[summary_key + "_existing"] += 1
        return
    api.rest("/order", "POST", {
        "type": "testorder",
        "patient": patient_uuid,
        "concept": concept_uuid,
        "orderType": order_type_uuid,
        "careSetting": care_setting,
        "encounter": encounter_uuid,
        "orderer": provider_uuid,
        "action": "NEW",
    })
    summary[summary_key + "_created"] += 1


def ensure_drug_order(api: BahmniApi, orders: list[dict[str, Any]], patient_uuid: str, encounter_uuid: str, provider_uuid: str, seed: dict[str, Any], summary: dict[str, int]) -> None:
    drugs = api.results("/drug", q=seed["known_openmrs_uuids"]["drug_search"], v="default", limit=1)
    if not drugs:
        summary["medication_orders_skipped"] += 1
        return
    drug_uuid = drugs[0]["uuid"]
    if order_exists(orders, drug_uuid=drug_uuid, order_type_uuid=seed["known_openmrs_uuids"]["order_types"]["drug_order"]):
        summary["medication_orders_existing"] += 1
        return
    concepts = seed["concepts"]
    api.rest("/order", "POST", {
        "type": "drugorder",
        "patient": patient_uuid,
        "drug": drug_uuid,
        "orderType": seed["known_openmrs_uuids"]["order_types"]["drug_order"],
        "careSetting": seed["known_openmrs_uuids"]["care_setting_outpatient"],
        "encounter": encounter_uuid,
        "orderer": provider_uuid,
        "action": "NEW",
        "dosingType": "org.openmrs.SimpleDosingInstructions",
        "dose": 5,
        "doseUnits": concepts["dose_mg"],
        "route": concepts["route_oral"],
        "frequency": seed["known_openmrs_uuids"]["frequency_once_daily"],
        "quantity": 10,
        "quantityUnits": concepts["quantity_tablet"],
        "duration": 3,
        "durationUnits": concepts["duration_days"],
        "numRefills": 0,
    })
    summary["medication_orders_created"] += 1


def history_event_count(seed_data: dict[str, Any]) -> int:
    return sum(len(patient.get("history_events", [])) for patient in seed_data.get("patients", []))


def history_event_text_observations(event: dict[str, Any], concepts: dict[str, str], when: str) -> list[dict[str, Any]]:
    lines = [
        f"FM-BAHMNI-LAB-R1 history event {event['event_id']}: {event.get('summary', '')}",
    ]
    for field, label in [
        ("diagnoses", "diagnoses"),
        ("tests", "tests"),
        ("medications", "medications"),
        ("procedures", "procedures"),
        ("notes", "notes"),
    ]:
        values = event.get(field) or []
        if values:
            lines.append(f"FM-BAHMNI-LAB-R1 {label}: " + "; ".join(str(value) for value in values))
    billing = event.get("billing") or {}
    if billing:
        lines.append(
            "FM-BAHMNI-LAB-R1 billing note: "
            f"invoice={billing.get('invoice_id')}; status={billing.get('status')}; "
            f"amount_inr={billing.get('amount_inr')}; payer={billing.get('payer')}"
        )
    return [{"concept": concepts["document_text"], "value": line, "obsDatetime": when} for line in lines if line]


def history_event_numeric_observations(event: dict[str, Any], concepts: dict[str, str], when: str) -> list[dict[str, Any]]:
    concept_keys = {
        "weight_kg": "weight_kg",
        "height_cm": "height_cm",
        "temperature_c": "temperature_c",
        "pulse": "pulse",
        "respiratory_rate": "respiratory_rate",
        "systolic_bp": "systolic_bp",
        "diastolic_bp": "diastolic_bp",
    }
    lab_keys = {
        "serum_glucose": "serum_glucose",
        "white_blood_cells": "white_blood_cells",
        "serum_creatinine_mg_dl": "serum_creatinine_mg_dl",
    }
    obs: list[dict[str, Any]] = []
    for source, mapping in [(event.get("vitals") or {}, concept_keys), (event.get("lab_results") or {}, lab_keys)]:
        for key, value in source.items():
            concept_key = mapping.get(key)
            if concept_key and concept_key in concepts:
                obs.append({"concept": concepts[concept_key], "value": value, "obsDatetime": when})
    return obs


def seed_history_event(
    api: BahmniApi,
    patient: dict[str, Any],
    event: dict[str, Any],
    patient_uuid: str,
    provider_uuid_by_identifier: dict[str, str],
    visit_types: dict[str, str],
    required_encounter_types: dict[str, str],
    location_by_department: dict[str, str],
    parent_location: str,
    concepts: dict[str, str],
    summary: dict[str, int],
) -> None:
    event_id = event["event_id"]
    provider_identifier = event.get("provider") or patient["provider"]
    provider_uuid = provider_uuid_by_identifier[provider_identifier]
    location = location_by_department.get(event.get("department") or patient["department"], parent_location)
    visit_type_uuid = visit_types.get(event.get("visit_type") or patient["visit_type"]) or visit_types.get("OPD") or next(iter(visit_types.values()))
    visit_uuid = ensure_visit(api, patient_uuid, visit_type_uuid, event["start"], location, summary, stop=event.get("stop"))
    when = event.get("encounter_datetime") or event["start"]
    legacy_marker_prefix = f"FM-BAHMNI-LAB-R1 history {patient['identifier']} {event_id}"
    consultation_obs = history_event_numeric_observations(event, concepts, when) + history_event_text_observations(event, concepts, when)
    ensure_encounter(
        api,
        patient_uuid,
        required_encounter_types["Consultation"],
        when,
        location,
        visit_uuid,
        consultation_obs,
        stable_record_marker("encounter", patient["identifier"], "history", event_id, "consultation"),
        concepts["document_text"],
        summary,
        legacy_markers=[f"{legacy_marker_prefix} consultation"],
    )
    if event.get("lab_results"):
        ensure_encounter(
            api,
            patient_uuid,
            required_encounter_types["LAB_RESULT"],
            when,
            location,
            visit_uuid,
            history_event_numeric_observations({"lab_results": event.get("lab_results", {})}, concepts, when),
            stable_record_marker("encounter", patient["identifier"], "history", event_id, "lab-result"),
            concepts["document_text"],
            summary,
            legacy_markers=[f"{legacy_marker_prefix} lab-result"],
        )
    if event.get("tests") or event.get("procedures"):
        ensure_encounter(
            api,
            patient_uuid,
            required_encounter_types["INVESTIGATION"],
            when,
            location,
            visit_uuid,
            history_event_text_observations({"event_id": event_id, "summary": event.get("summary", ""), "tests": event.get("tests", []), "procedures": event.get("procedures", [])}, concepts, when),
            stable_record_marker("encounter", patient["identifier"], "history", event_id, "investigation"),
            concepts["document_text"],
            summary,
            legacy_markers=[f"{legacy_marker_prefix} investigation"],
        )
    if event.get("billing") or event.get("notes"):
        ensure_encounter(
            api,
            patient_uuid,
            required_encounter_types["Patient Document"],
            when,
            location,
            visit_uuid,
            history_event_text_observations({"event_id": event_id, "summary": event.get("summary", ""), "billing": event.get("billing", {}), "notes": event.get("notes", [])}, concepts, when),
            stable_record_marker("encounter", patient["identifier"], "history", event_id, "document-billing"),
            concepts["document_text"],
            summary,
            legacy_markers=[f"{legacy_marker_prefix} document-billing"],
        )


def new_summary() -> dict[str, int]:
    keys = [
        "locations_created", "locations_existing", "providers_created", "providers_existing", "patients_created", "patients_existing",
        "visits_created", "visits_existing", "encounters_created", "encounters_existing", "appointments_created", "appointments_existing",
        "providers_reconciled", "patients_reconciled", "encounters_reconciled", "appointments_reconciled",
        "encounter_duplicates_preserved", "appointment_duplicates_preserved",
        "allergies_created", "allergies_existing", "allergies_reconciled", "conditions_created", "conditions_existing", "conditions_skipped",
        "lab_orders_created", "lab_orders_existing", "radiology_orders_created", "radiology_orders_existing", "procedure_orders_created", "procedure_orders_existing",
        "test_orders_created", "test_orders_existing", "medication_orders_created", "medication_orders_existing", "medication_orders_skipped",
    ]
    return {key: 0 for key in keys}


def seed(args: argparse.Namespace) -> None:
    if args.dry_run:
        seed_data = load_json(SEED_PATH)
        dump_json({
            "dry_run": True,
            "dataset_id": seed_data["dataset_id"],
            "planned_counts": {
                "departments": len(seed_data["departments"]),
                "service_locations": len(seed_data["service_locations"]),
                "providers": len(provider_records(seed_data)),
                "clinical_providers": len(seed_data["providers"]),
                "staff": len(seed_data.get("staff", [])),
                "patients": len(seed_data["patients"]),
                "history_events": history_event_count(seed_data),
                "appointment_per_patient": 1,
                "visit_per_patient": 1,
                "encounter_families_per_patient": 4,
                "order_families_per_patient": 5,
            },
            "mechanisms": ["OpenMRS REST", "Bahmni appointments REST", "OpenMRS FHIR2 R4"],
        })
        return

    check_synthetic(argparse.Namespace(json=False, online_source=False, quiet=True))
    api = BahmniApi(args.base_url)
    seed_data = load_json(SEED_PATH)
    summary = new_summary()
    concepts = seed_data["concepts"]

    # Metadata resolution from the running distro.
    base_location = api.first_by_display("/location", seed_data["known_openmrs_uuids"]["base_location_name"])["uuid"]
    patient_identifier_type = api.first_by_display("/patientidentifiertype", seed_data["known_openmrs_uuids"]["patient_identifier_type_name"])["uuid"]
    visit_types = {item.get("display"): item["uuid"] for item in api.results("/visittype", v="default", limit=100)}
    encounter_types = {item.get("display"): item["uuid"] for item in api.results("/encountertype", v="default", limit=100)}
    appointment_services = api.results("/appointmentService/all/default")
    appointment_service_uuid = appointment_services[0]["uuid"] if appointment_services else None
    if not appointment_service_uuid:
        raise ApiError("no appointment services are available in this Bahmni deployment")
    # Resolve every required encounter type up front so a distro that renames or
    # omits one fails before any record is written.
    required_encounter_types = {}
    for name in ("Consultation", "LAB_RESULT", "INVESTIGATION", "Patient Document"):
        uuid = encounter_types.get(name)
        if not uuid:
            raise ApiError(f"encounter type {name!r} is not available in this Bahmni deployment")
        required_encounter_types[name] = uuid

    parent_location = ensure_location(api, seed_data["organization"]["name"], base_location, summary)
    location_by_department: dict[str, str] = {}
    for department in seed_data["departments"]:
        location_by_department[department] = ensure_location(api, f"Chikitsalaya Department - {department}", parent_location, summary)
    for service_location in seed_data["service_locations"]:
        ensure_location(api, service_location, parent_location, summary)

    provider_uuid_by_identifier = {provider["identifier"]: ensure_provider(api, provider, summary) for provider in provider_records(seed_data)}

    base_date = dt.datetime(2026, 1, 15, 9, 0, tzinfo=dt.timezone.utc)
    for index, patient in enumerate(seed_data["patients"]):
        patient_uuid = ensure_patient(api, patient, patient_identifier_type, base_location, summary)
        provider_uuid = provider_uuid_by_identifier[patient["provider"]]
        department_location = location_by_department.get(patient["department"], parent_location)
        visit_type_uuid = visit_types.get(patient["visit_type"]) or visit_types.get("OPD") or next(iter(visit_types.values()))
        visit_start_dt = base_date + dt.timedelta(days=index)
        visit_start = visit_start_dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000")
        visit_stop = None
        if patient.get("visit_status") == "completed":
            visit_stop = (visit_start_dt + dt.timedelta(hours=2)).strftime("%Y-%m-%dT%H:%M:%S.000+0000")
        visit_uuid = ensure_visit(api, patient_uuid, visit_type_uuid, visit_start, department_location, summary, stop=visit_stop)

        current_consultation_marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} consultation"
        legacy_consultation_marker = f"FM-BAHMNI-LAB-R1 synthetic encounter {patient['identifier']} consultation"
        marker = stable_record_marker("encounter", patient["identifier"], "baseline", "consultation")
        obs_time = (base_date + dt.timedelta(days=index, minutes=15)).strftime("%Y-%m-%dT%H:%M:%S.000+0000")
        is_cold_fever_visit = patient["identifier"] == "SYN-HEN-0009"
        vitals = [
            {"concept": concepts["weight_kg"], "value": round(58.0 + index * 2.7, 1), "obsDatetime": obs_time},
            {"concept": concepts["height_cm"], "value": 150 + index * 3, "obsDatetime": obs_time},
            {"concept": concepts["temperature_c"], "value": 38.6 if is_cold_fever_visit else round(36.4 + (index % 4) * 0.2, 1), "obsDatetime": obs_time},
            {"concept": concepts["pulse"], "value": 96 if is_cold_fever_visit else 72 + index, "obsDatetime": obs_time},
            {"concept": concepts["systolic_bp"], "value": 112 + index * 2, "obsDatetime": obs_time},
            {"concept": concepts["diastolic_bp"], "value": 72 + index, "obsDatetime": obs_time},
        ]
        if is_cold_fever_visit:
            vitals.extend([
                {"concept": concepts["respiratory_rate"], "value": 22, "obsDatetime": obs_time},
                {"concept": concepts["document_text"], "value": patient["chief_complaint"], "obsDatetime": obs_time},
            ])
            for diagnosis_note in patient.get("diagnosis_notes", []):
                vitals.append({"concept": concepts["document_text"], "value": diagnosis_note, "obsDatetime": obs_time})
        consultation_encounter = ensure_encounter(
            api,
            patient_uuid,
            required_encounter_types["Consultation"],
            obs_time,
            department_location,
            visit_uuid,
            vitals,
            marker,
            concepts["document_text"],
            summary,
            legacy_markers=[current_consultation_marker, legacy_consultation_marker],
        )

        current_lab_marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} lab-result"
        legacy_lab_marker = f"FM-BAHMNI-LAB-R1 synthetic encounter {patient['identifier']} lab-result"
        lab_marker = stable_record_marker("encounter", patient["identifier"], "baseline", "lab-result")
        lab_obs = [
            {"concept": concepts["serum_glucose"], "value": 90 + index * 3, "obsDatetime": obs_time},
            {"concept": concepts["white_blood_cells"], "value": round(5.0 + index * 0.4, 1), "obsDatetime": obs_time},
            {"concept": concepts["serum_creatinine_mg_dl"], "value": round(0.7 + index * 0.03, 2), "obsDatetime": obs_time},
        ]
        ensure_encounter(
            api, patient_uuid, required_encounter_types["LAB_RESULT"], obs_time, department_location, visit_uuid,
            lab_obs, lab_marker, concepts["document_text"], summary,
            legacy_markers=[current_lab_marker, legacy_lab_marker],
        )

        current_investigation_marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} investigation"
        legacy_investigation_marker = f"FM-BAHMNI-LAB-R1 synthetic encounter {patient['identifier']} investigation"
        investigation_marker = stable_record_marker("encounter", patient["identifier"], "baseline", "investigation")
        ensure_encounter(
            api, patient_uuid, required_encounter_types["INVESTIGATION"], obs_time, department_location, visit_uuid,
            [], investigation_marker, concepts["document_text"], summary,
            legacy_markers=[current_investigation_marker, legacy_investigation_marker],
        )

        current_document_marker = f"{seed_data['records']['document_note']} for {patient['identifier']}"
        legacy_document_marker = f"FM-BAHMNI-LAB-R1 synthetic document metadata only; no real attachment bytes are stored for {patient['identifier']}"
        document_marker = stable_record_marker("encounter", patient["identifier"], "baseline", "document")
        ensure_encounter(
            api, patient_uuid, required_encounter_types["Patient Document"], obs_time, department_location, visit_uuid,
            [], document_marker, concepts["document_text"], summary,
            legacy_markers=[current_document_marker, legacy_document_marker],
        )

        appointment_start_dt = dt.datetime(2026, 1, 20 + index, 9 + (index % 6), 0, tzinfo=dt.timezone.utc)
        appointment_end_dt = appointment_start_dt + dt.timedelta(minutes=15)
        appointment_start = appointment_start_dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000")
        appointment_comment_text = f"{seed_data['records']['appointment_comment_prefix']} {patient['identifier']}"
        legacy_appointment_comment = f"FM-BAHMNI-LAB-R1 synthetic appointment {patient['identifier']}"
        appointment_marker = stable_record_marker("appointment", patient["identifier"], str(appointment_ms(appointment_start)))
        ensure_appointment(
            api,
            patient["identifier"],
            patient_uuid,
            provider_uuid,
            appointment_service_uuid,
            department_location,
            appointment_start,
            appointment_end_dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000"),
            appointment_comment_text,
            appointment_marker,
            summary,
            legacy_comments=[legacy_appointment_comment],
        )

        condition_texts = seed_data["records"]["condition_texts"]
        condition_text = condition_texts[index] if index < len(condition_texts) else f"Clinical condition follow-up {patient['identifier']}"
        ensure_condition(api, patient["identifier"], condition_text, summary)
        ensure_allergy(
            api,
            patient_uuid,
            concepts["penicillin"],
            f"{seed_data['records']['allergy_comment']} {patient['identifier']}",
            summary,
            legacy_comments=[f"FM-BAHMNI-LAB-R1 synthetic low-severity medication allergy; not real patient data {patient['identifier']}"],
        )

        orders = existing_orders(api, patient_uuid)
        order_types = seed_data["known_openmrs_uuids"]["order_types"]
        care_setting = seed_data["known_openmrs_uuids"]["care_setting_outpatient"]
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["complete_blood_count"], order_types["test_order"], care_setting, "test_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["complete_blood_count"], order_types["lab_order"], care_setting, "lab_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["ultrasound"], order_types["radiology_order"], care_setting, "radiology_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["electrocardiogram_diagnosis"], order_types["procedure_order"], care_setting, "procedure_orders", summary)
        ensure_drug_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, seed_data, summary)

        for event in patient.get("history_events", []):
            seed_history_event(
                api,
                patient,
                event,
                patient_uuid,
                provider_uuid_by_identifier,
                visit_types,
                required_encounter_types,
                location_by_department,
                parent_location,
                concepts,
                summary,
            )

    result = {
        "ok": True,
        "dataset_id": seed_data["dataset_id"],
        "summary": summary,
        "urls": local_urls(),
        "note": "Counts distinguish created vs existing records so re-running the seed is safe and idempotent.",
    }
    dump_json(result)


def health(args: argparse.Namespace) -> dict[str, Any]:
    urls = local_urls()
    checks: list[dict[str, Any]] = []
    for name, url in [
        ("openmrs_session", urls["openmrs_rest"] + "/session"),
        ("fhir_metadata", urls["openmrs_fhir"] + "/metadata"),
        ("bahmni_web", urls["bahmni"]),
    ]:
        try:
            req = urllib.request.Request(url, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req, context=ssl._create_unverified_context(), timeout=20) as resp:
                checks.append({"name": name, "url": url, "status": resp.status, "ok": resp.status in (200, 302)})
        except Exception as exc:
            checks.append({"name": name, "url": url, "ok": False, "error": type(exc).__name__})
    result = {"ok": all(item["ok"] for item in checks), "checks": checks, "urls": urls}
    if args.json:
        dump_json(result)
    else:
        for check in checks:
            print(check)
    if not result["ok"]:
        raise SystemExit(1)
    return result


def verify(args: argparse.Namespace) -> None:
    seed_data = load_json(SEED_PATH)
    support = load_json(MODULE_SUPPORT_PATH)
    synthetic = check_synthetic(argparse.Namespace(json=False, online_source=False, quiet=True))
    offline = {
        "pin": load_json(PIN_PATH)["official_commit"],
        "dataset_id": seed_data["dataset_id"],
        "departments": len(seed_data["departments"]),
        "providers": len(provider_records(seed_data)),
        "clinical_providers": len(seed_data["providers"]),
        "staff": len(seed_data.get("staff", [])),
        "patients": len(seed_data["patients"]),
        "history_events": history_event_count(seed_data),
        "karthik_history_events": next((len(patient.get("history_events", [])) for patient in seed_data["patients"] if patient.get("identifier") == "SYN-HEN-0009"), 0),
        "rohit_history_events": next((len(patient.get("history_events", [])) for patient in seed_data["patients"] if patient.get("identifier") == "SYN-HEN-0010"), 0),
        "module_families_seeded": len(support["supported_and_seeded"]),
        "synthetic_check_ok": synthetic["ok"],
        "unicode_hospital_name": seed_data["organization"]["name"],
        "unicode_hospital_name_utf8_sha256": hashlib.sha256(seed_data["organization"]["name"].encode("utf-8")).hexdigest(),
        "karthik_patient_present": any(patient.get("given_name") == "Karthik" and patient.get("identifier") == "SYN-HEN-0009" for patient in seed_data["patients"]),
        "rohit_patient_present": any(patient.get("given_name") == "Rohit" and patient.get("identifier") == "SYN-HEN-0010" for patient in seed_data["patients"]),
    }
    if args.offline:
        dump_json({"ok": True, "offline": offline})
        return

    api = BahmniApi(args.base_url)
    _, session = api.rest("/session")
    if not session.get("authenticated"):
        raise SystemExit("OpenMRS authentication failed")
    counts = {
        "patients": 0,
        "providers": 0,
        "conditions": 0,
        "allergies": 0,
        "orders": 0,
        "encounters": 0,
    }
    identity_display_names_exact = True
    allergy_comments_exact = True
    for patient in seed_data["patients"]:
        patient_hits = api.results("/patient", q=patient["identifier"], v="default", limit=5)
        patient_hit = exact_owned_record(patient_hits, patient["identifier"])
        if patient_hit and exact_display_matches(patient_hit, patient):
            counts["patients"] += 1
            patient_uuid = patient_hit["uuid"]
            counts["orders"] += len(api.results("/order", patient=patient_uuid, v="default", limit=100))
            counts["encounters"] += len(api.results("/encounter", patient=patient_uuid, v="default", limit=100))
            try:
                _, allergies = api.rest(f"/patient/{patient_uuid}/allergy", params={"v": "full"})
                allergy_results = allergies.get("results", []) if isinstance(allergies, dict) else []
                counts["allergies"] += len(allergy_results)
                expected_comment = f"{seed_data['records']['allergy_comment']} {patient['identifier']}"
                allergy_comments_exact = allergy_comments_exact and any(
                    item.get("allergen", {}).get("codedAllergen", {}).get("uuid") == seed_data["concepts"]["penicillin"]
                    and item.get("comment") == expected_comment
                    for item in allergy_results
                )
            except ApiError:
                allergy_comments_exact = False
            fhir_id = fhir_patient_id(api, patient["identifier"])
            if fhir_id:
                _, condition_bundle = api.fhir("/Condition", params={"patient": fhir_id})
                counts["conditions"] += int(condition_bundle.get("total", 0) or 0)
        else:
            identity_display_names_exact = False
            allergy_comments_exact = False
    for provider in provider_records(seed_data):
        provider_hit = exact_owned_record(api.results("/provider", q=provider["identifier"], v="default", limit=5), provider["identifier"])
        if provider_hit and exact_display_matches(provider_hit, provider):
            counts["providers"] += 1
        else:
            identity_display_names_exact = False
    _, appointments = api.rest("/appointments")
    appointment_count = 0
    appointment_identity_ok = True
    if isinstance(appointments, list):
        for index, patient in enumerate(seed_data["patients"]):
            start = dt.datetime(2026, 1, 20 + index, 9 + (index % 6), 0, tzinfo=dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.000+0000")
            marker = stable_record_marker("appointment", patient["identifier"], str(appointment_ms(start)))
            matches = [
                item for item in appointments
                if item.get("patient", {}).get("identifier") == patient["identifier"] and f"[{marker}]" in str(item.get("comments") or "")
            ]
            appointment_count += len(matches)
            appointment_identity_ok = appointment_identity_ok and len(matches) == 1
    else:
        appointment_identity_ok = False
    fhir_patient_ids = {patient["identifier"]: fhir_patient_id(api, patient["identifier"]) for patient in seed_data["patients"]}
    fhir_patient_exact_count = sum(1 for value in fhir_patient_ids.values() if value)
    _, fhir_observations = api.fhir("/Observation", params={"code": seed_data["concepts"]["weight_kg"]})

    unicode_location_present = any(
        item.get("display") == seed_data["organization"]["name"]
        for item in api.results("/location", q=seed_data["organization"]["name"], v="default", limit=20)
    )
    karthik_live: dict[str, Any] = {"patient_found": False, "completed_visit_found": False, "cold_fever_conditions_found": False, "fever_observation_found": False}
    karthik_record = next(patient for patient in seed_data["patients"] if patient["identifier"] == "SYN-HEN-0009")
    karthik_hit = exact_owned_record(api.results("/patient", q="SYN-HEN-0009", v="default", limit=5), "SYN-HEN-0009")
    if karthik_hit and exact_display_matches(karthik_hit, karthik_record):
        karthik_live["patient_found"] = True
        karthik_uuid = karthik_hit["uuid"]
        visits = api.results("/visit", patient=karthik_uuid, v="full", limit=50)
        karthik_live["completed_visit_found"] = any(bool(visit.get("stopDatetime")) for visit in visits)
        encounters = api.results("/encounter", patient=karthik_uuid, v="full", limit=100)
        karthik_live["fever_observation_found"] = any(
            any(obs.get("concept", {}).get("uuid") == seed_data["concepts"]["temperature_c"] and float(obs.get("value", 0) or 0) >= 38.0 for obs in encounter.get("obs", []))
            for encounter in encounters
        )
        cold_fever_note_found = any(
            any(isinstance(obs.get("value"), str) and "cold and fever" in obs.get("value", "").lower() for obs in encounter.get("obs", []))
            for encounter in encounters
        )
        patient_id = fhir_patient_id(api, "SYN-HEN-0009")
        if patient_id:
            _, condition_bundle = api.fhir("/Condition", params={"patient": patient_id})
            karthik_live["cold_fever_conditions_found"] = condition_bundle.get("total", 0) > 0 and cold_fever_note_found

    history_live: dict[str, dict[str, Any]] = {}
    for patient in seed_data["patients"]:
        expected_events = patient.get("history_events", [])
        if not expected_events:
            continue
        patient_hit = exact_owned_record(api.results("/patient", q=patient["identifier"], v="default", limit=5), patient["identifier"])
        found_events = 0
        if patient_hit:
            encounters = api.results("/encounter", patient=patient_hit["uuid"], v="full", limit=100)
            for event in expected_events:
                marker_prefix = stable_record_marker("encounter", patient["identifier"], "history", event["event_id"])
                if any(marker_in_obs(encounter, marker_prefix) for encounter in encounters):
                    found_events += 1
        history_live[patient["identifier"]] = {"expected": len(expected_events), "found": found_events, "ok": found_events == len(expected_events)}

    live_ok = counts["patients"] == len(seed_data["patients"]) and counts["providers"] == len(provider_records(seed_data))
    live_ok = live_ok and identity_display_names_exact and appointment_identity_ok and allergy_comments_exact
    live_ok = live_ok and fhir_patient_exact_count == len(seed_data["patients"])
    live_ok = live_ok and int(fhir_observations.get("total", 0) or 0) > 0
    live_ok = live_ok and unicode_location_present and all(karthik_live.values()) and all(item["ok"] for item in history_live.values())
    result = {
        "ok": live_ok,
        "offline": offline,
        "live_counts": counts | {"appointments": appointment_count},
        "identity_display_names_exact": identity_display_names_exact,
        "appointment_identity_ok": appointment_identity_ok,
        "allergy_comments_exact": allergy_comments_exact,
        "unicode_location_present": unicode_location_present,
        "karthik_live": karthik_live,
        "history_live": history_live,
        "fhir": {
            "patient_exact_match_count": fhir_patient_exact_count,
            "sample_patient_id_present": bool(fhir_patient_ids.get("SYN-HEN-0009")),
            "weight_observation_total": fhir_observations.get("total"),
        },
        "urls": local_urls(),
        "known_gaps": support["enabled_but_not_seeded_or_limited"],
    }
    dump_json(result)
    if not result["ok"]:
        raise SystemExit(1)


def inventory(args: argparse.Namespace) -> None:
    podman = shutil.which("podman")
    if not podman:
        raise SystemExit("podman not found")
    result: dict[str, Any] = {"podman": run([podman, "--version"], check=False).stdout.strip(), "machines": [], "connections": []}
    machines = run([podman, "machine", "list", "--format", "json"], check=False)
    if machines.returncode == 0 and machines.stdout.strip():
        result["machines"] = json.loads(machines.stdout)
    conns = run([podman, "system", "connection", "list", "--format", "json"], check=False)
    if conns.returncode == 0 and conns.stdout.strip():
        result["connections"] = json.loads(conns.stdout)
    workloads = []
    for conn in result["connections"]:
        name = conn.get("Name")
        if not name or str(name).endswith("-root"):
            continue
        entry: dict[str, Any] = {"connection": name}
        for kind, cmd in [
            ("containers", [podman, "--connection", name, "ps", "--all", "--format", "json"]),
            ("networks", [podman, "--connection", name, "network", "ls", "--format", "json"]),
            ("volumes", [podman, "--connection", name, "volume", "ls", "--format", "json"]),
        ]:
            proc = run(cmd, check=False)
            if proc.returncode == 0 and proc.stdout.strip():
                try:
                    values = json.loads(proc.stdout)
                    entry[kind] = {"count": len(values), "names": [(v.get("Names") or v.get("Name") or v.get("name")) for v in values[:50]]}
                except json.JSONDecodeError:
                    entry[kind] = {"error": "invalid json"}
            else:
                entry[kind] = {"unavailable": True}
        workloads.append(entry)
    result["workloads"] = workloads
    dump_json(result)


def ownership(args: argparse.Namespace) -> None:
    try:
        if args.action in ("recover", "forget") and not args.yes:
            raise ValueError(f"ownership {args.action} requires --yes")
        if args.action == "claim":
            marker = claim_ownership()
        elif args.action == "recover":
            marker = recover_ownership()
        elif args.action == "forget":
            forget_ownership()
            marker = {"owner": LAB_ID}
        else:
            marker = load_ownership_marker()
        if args.resources and args.action != "forget":
            verify_owned_podman_resources()
    except ValueError as exc:
        raise SystemExit(f"Podman ownership check failed: {exc}") from exc
    if args.json:
        dump_json({"ok": True, "owner": marker["owner"], "action": args.action, "resources_checked": args.resources})


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(description="Bahmni Podman synthetic lab helper")
    sub = parser.add_subparsers(dest="command", required=True)

    p = sub.add_parser("prepare")
    p.add_argument("--dry-run", action="store_true")
    p.set_defaults(func=prepare)

    p = sub.add_parser("credentials")
    p.add_argument("--show", action="store_true", help="print local-only credentials; use only in a private terminal")
    p.add_argument("--path", action="store_true")
    p.set_defaults(func=credentials)

    p = sub.add_parser("check-synthetic")
    p.add_argument("--json", action="store_true")
    p.add_argument("--online-source", action="store_true", help="fetch the public taxonomy page and verify synthetic provider names do not collide")
    p.set_defaults(func=check_synthetic)

    p = sub.add_parser("seed")
    p.add_argument("--dry-run", action="store_true")
    p.add_argument("--json", action="store_true")
    p.add_argument("--base-url")
    p.set_defaults(func=seed)

    p = sub.add_parser("health")
    p.add_argument("--json", action="store_true")
    p.set_defaults(func=health)

    p = sub.add_parser("verify")
    p.add_argument("--offline", action="store_true")
    p.add_argument("--base-url")
    p.set_defaults(func=verify)

    p = sub.add_parser("inventory")
    p.add_argument("--json", action="store_true")
    p.set_defaults(func=inventory)

    p = sub.add_parser("ownership")
    p.add_argument("action", choices=("claim", "verify", "recover", "forget"))
    p.add_argument("--resources", action="store_true")
    p.add_argument("--yes", action="store_true")
    p.add_argument("--json", action="store_true")
    p.set_defaults(func=ownership)

    args = parser.parse_args(argv)
    args.func(args)


if __name__ == "__main__":
    main()
