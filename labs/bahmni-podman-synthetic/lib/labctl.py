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
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
PIN_PATH = ROOT / "config" / "pin.json"
SEED_PATH = ROOT / "fixtures" / "synthetic-seed.json"
TAXONOMY_PATH = ROOT / "config" / "sparsh-hennur-taxonomy.json"
MODULE_SUPPORT_PATH = ROOT / "config" / "module-support.json"

LAB_ID = "fm-bahmni-lab-r1"
DEFAULT_HOME = Path("/tmp") / LAB_ID


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
    # Official odoo-16 image for the pinned Standard tag is amd64-only as of the
    # research run. Platform pin lets rootless Podman on Apple Silicon use qemu.
    text = add_platform_after_service(text, "odoo", "linux/amd64")
    validate_loopback_ports(text)
    validate_container_names(text)
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


def check_synthetic(args: argparse.Namespace) -> dict[str, Any]:
    seed = load_json(SEED_PATH)
    failures: list[str] = []
    if not seed.get("synthetic"):
        failures.append("dataset synthetic flag is false")
    if seed.get("organization", {}).get("name") != "Chikitsalayaḥ":
        failures.append("synthetic hospital name must be exactly Chikitsalayaḥ")
    if "SPARSH" in json.dumps(seed.get("providers", []), ensure_ascii=False):
        failures.append("provider fixture contains SPARSH")
    for provider in seed.get("providers", []):
        ident = provider.get("identifier", "")
        if not ident.startswith("SYN-PROV-"):
            failures.append(f"provider identifier not synthetic: {ident}")
        full = f"{provider.get('given_name', '')} {provider.get('family_name', '')}"
        norm = normalize_name(full)
        if any(token in norm for token in ["doctor", "sparsh", "hospital", "praveen", "abhijit"]):
            failures.append(f"provider {ident} contains a forbidden identity token")
        if not any(marker in provider.get("family_name", "").lower() for marker in ["demo", "test", "sim", "mock", "fict", "synth"]):
            failures.append(f"provider {ident} family name is not obviously synthetic")
    for patient in seed.get("patients", []):
        ident = patient.get("identifier", "")
        if not ident.startswith("SYN-HEN-"):
            failures.append(f"patient identifier not synthetic: {ident}")
        if not any(marker in patient.get("family_name", "").lower() for marker in ["demo", "test", "sim", "mock", "fict", "trial", "connector", "synth"]):
            failures.append(f"patient {ident} family name is not obviously synthetic")
        contact = patient.get("synthetic_contact", {})
        if contact:
            phone = contact.get("phone", "")
            email = contact.get("email", "")
            if phone != seed.get("contact_defaults", {}).get("patient_phone_placeholder", "000-000-0000"):
                failures.append(f"patient {ident} contact phone is not the invalid synthetic placeholder")
            if email and not email.endswith(".invalid"):
                failures.append(f"patient {ident} contact email must use the .invalid TLD")

    online_checked = False
    if args.online_source:
        online_checked = True
        source = seed["source_taxonomy"]["source_url"]
        try:
            html = urllib.request.urlopen(source, timeout=20).read().decode("utf-8", "replace")
            # Compare normalized synthetic provider names against live page text without
            # printing or storing real names.
            page_norm = normalize_name(html)
            for provider in seed.get("providers", []):
                full_norm = normalize_name(f"{provider['given_name']} {provider['family_name']}")
                if full_norm and full_norm in page_norm:
                    failures.append(f"provider {provider['identifier']} collides with online source text")
        except Exception as exc:  # pragma: no cover - network optional
            failures.append(f"online source check failed: {type(exc).__name__}")

    result = {
        "ok": not failures,
        "dataset_id": seed.get("dataset_id"),
        "providers_checked": len(seed.get("providers", [])),
        "patients_checked": len(seed.get("patients", [])),
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


def derive_fhir_base(rest_base: str) -> str:
    trimmed = rest_base.rstrip("/")
    if trimmed.endswith(REST_SUFFIX):
        trimmed = trimmed[: -len(REST_SUFFIX)]
    return trimmed + FHIR_SUFFIX


class BahmniApi:
    def __init__(self, base_https: str | None = None):
        creds = read_credentials()
        urls = creds.get("urls", local_urls())
        self.rest_base = (base_https or urls["openmrs_rest"]).rstrip("/")
        # A --base-url override must move REST and FHIR together, otherwise one
        # run would write to two different deployments.
        if base_https:
            self.fhir_base = derive_fhir_base(self.rest_base)
        else:
            self.fhir_base = urls.get("openmrs_fhir", derive_fhir_base(self.rest_base)).rstrip("/")
        user = creds.get("openmrs", {}).get("username") or "admin"
        password = creds.get("openmrs", {}).get("password") or parse_env_file(generated_env_path()).get("OPENMRS_ATOMFEED_PASSWORD", "")
        token = base64.b64encode(f"{user}:{password}".encode("utf-8")).decode("ascii")
        self.auth_header = "Basic " + token
        self.ctx = ssl._create_unverified_context()

    def request(self, base: str, path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None, fhir: bool = False) -> tuple[int, Any]:
        url = base.rstrip("/") + path
        if params:
            clean = {k: v for k, v in params.items() if v is not None}
            url += "?" + urllib.parse.urlencode(clean)
        headers = {"Authorization": self.auth_header, "Accept": "application/fhir+json" if fhir else "application/json"}
        data = None
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/fhir+json" if fhir else "application/json"
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, context=self.ctx, timeout=45) as resp:
                raw = resp.read().decode("utf-8", "replace")
                if not raw:
                    return resp.status, {}
                return resp.status, json.loads(raw)
        except urllib.error.HTTPError as exc:
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


def ensure_location(api: BahmniApi, name: str, parent: str | None, summary: dict[str, int]) -> str:
    existing = exact_search(api, "/location", name)
    if existing:
        summary["locations_existing"] += 1
        return existing["uuid"]
    body: dict[str, Any] = {
        "name": name,
        "description": "FM-BAHMNI-LAB-R1 synthetic local connector lab location; not a real facility.",
    }
    if parent:
        body["parentLocation"] = parent
    _, created = api.rest("/location", "POST", body)
    summary["locations_created"] += 1
    return created["uuid"]


def ensure_provider(api: BahmniApi, provider: dict[str, str], summary: dict[str, int]) -> str:
    ident = provider["identifier"]
    for item in api.results("/provider", q=ident, v="default", limit=20):
        if ident in (item.get("display") or ""):
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
    for item in api.results("/patient", q=ident, v="default", limit=20):
        if ident in (item.get("display") or ""):
            summary["patients_existing"] += 1
            return item["uuid"]
    person = {
        "names": [{"givenName": patient["given_name"], "familyName": patient["family_name"]}],
        "gender": patient["gender"],
        "birthdate": patient["birthdate"],
        "addresses": [{
            "address1": f"Synthetic {ident} Loopback Lane",
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
    # value is UTC ISO without timezone suffix in fixture generation below.
    parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    return int(parsed.timestamp() * 1000)


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
) -> str:
    encounters = api.results("/encounter", patient=patient_uuid, v="full", limit=100)
    for encounter in encounters:
        if marker_in_obs(encounter, marker):
            summary["encounters_existing"] += 1
            return encounter["uuid"]
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


def ensure_appointment(api: BahmniApi, patient_ident: str, patient_uuid: str, provider_uuid: str, service_uuid: str, location: str, start: str, end: str, comment: str, summary: dict[str, int]) -> None:
    _, appointments = api.rest("/appointments")
    if isinstance(appointments, list):
        start_ms = appointment_ms(start.replace(".000+0000", "+00:00"))
        for appt in appointments:
            if appt.get("patient", {}).get("identifier") == patient_ident and appt.get("comments") == comment and appt.get("startDateTime") == start_ms:
                summary["appointments_existing"] += 1
                return
    body = {
        "patientUuid": patient_uuid,
        "serviceUuid": service_uuid,
        "appointmentKind": "Scheduled",
        "status": "Scheduled",
        "startDateTime": start,
        "endDateTime": end,
        "providers": [{"uuid": provider_uuid, "response": "ACCEPTED"}],
        "locationUuid": location,
        "comments": comment,
    }
    api.rest("/appointments", "POST", body)
    summary["appointments_created"] += 1


def ensure_allergy(api: BahmniApi, patient_uuid: str, concept_uuid: str, comment: str, summary: dict[str, int]) -> None:
    _, existing = api.rest(f"/patient/{patient_uuid}/allergy", params={"v": "full"})
    allergies = existing.get("results", []) if isinstance(existing, dict) else []
    for allergy in allergies:
        coded = allergy.get("allergen", {}).get("codedAllergen", {})
        if comment in (allergy.get("comment") or "") or coded.get("uuid") == concept_uuid:
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


def new_summary() -> dict[str, int]:
    keys = [
        "locations_created", "locations_existing", "providers_created", "providers_existing", "patients_created", "patients_existing",
        "visits_created", "visits_existing", "encounters_created", "encounters_existing", "appointments_created", "appointments_existing",
        "allergies_created", "allergies_existing", "conditions_created", "conditions_existing", "conditions_skipped",
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
                "providers": len(seed_data["providers"]),
                "patients": len(seed_data["patients"]),
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
        location_by_department[department] = ensure_location(api, f"Synthetic Department - {department}", parent_location, summary)
    for service_location in seed_data["service_locations"]:
        ensure_location(api, service_location, parent_location, summary)

    provider_uuid_by_identifier = {provider["identifier"]: ensure_provider(api, provider, summary) for provider in seed_data["providers"]}

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

        marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} consultation"
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
        )

        lab_marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} lab-result"
        lab_obs = [
            {"concept": concepts["serum_glucose"], "value": 90 + index * 3, "obsDatetime": obs_time},
            {"concept": concepts["white_blood_cells"], "value": round(5.0 + index * 0.4, 1), "obsDatetime": obs_time},
            {"concept": concepts["serum_creatinine_mg_dl"], "value": round(0.7 + index * 0.03, 2), "obsDatetime": obs_time},
        ]
        ensure_encounter(api, patient_uuid, required_encounter_types["LAB_RESULT"], obs_time, department_location, visit_uuid, lab_obs, lab_marker, concepts["document_text"], summary)

        investigation_marker = f"{seed_data['records']['encounter_marker_prefix']} {patient['identifier']} investigation"
        ensure_encounter(api, patient_uuid, required_encounter_types["INVESTIGATION"], obs_time, department_location, visit_uuid, [], investigation_marker, concepts["document_text"], summary)

        document_marker = f"{seed_data['records']['document_note']} for {patient['identifier']}"
        ensure_encounter(api, patient_uuid, required_encounter_types["Patient Document"], obs_time, department_location, visit_uuid, [], document_marker, concepts["document_text"], summary)

        appointment_start_dt = dt.datetime(2026, 1, 20 + index, 9 + (index % 6), 0, tzinfo=dt.timezone.utc)
        appointment_end_dt = appointment_start_dt + dt.timedelta(minutes=15)
        ensure_appointment(
            api,
            patient["identifier"],
            patient_uuid,
            provider_uuid,
            appointment_service_uuid,
            department_location,
            appointment_start_dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000"),
            appointment_end_dt.strftime("%Y-%m-%dT%H:%M:%S.000+0000"),
            f"{seed_data['records']['appointment_comment_prefix']} {patient['identifier']}",
            summary,
        )

        condition_texts = seed_data["records"]["condition_texts"]
        condition_text = condition_texts[index] if index < len(condition_texts) else f"FM-BAHMNI-LAB-R1 synthetic condition {patient['identifier']}"
        ensure_condition(api, patient["identifier"], condition_text, summary)
        ensure_allergy(api, patient_uuid, concepts["penicillin"], f"{seed_data['records']['allergy_comment']} {patient['identifier']}", summary)

        orders = existing_orders(api, patient_uuid)
        order_types = seed_data["known_openmrs_uuids"]["order_types"]
        care_setting = seed_data["known_openmrs_uuids"]["care_setting_outpatient"]
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["complete_blood_count"], order_types["test_order"], care_setting, "test_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["complete_blood_count"], order_types["lab_order"], care_setting, "lab_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["ultrasound"], order_types["radiology_order"], care_setting, "radiology_orders", summary)
        ensure_test_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, concepts["electrocardiogram_diagnosis"], order_types["procedure_order"], care_setting, "procedure_orders", summary)
        ensure_drug_order(api, orders, patient_uuid, consultation_encounter, provider_uuid, seed_data, summary)

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
        "providers": len(seed_data["providers"]),
        "patients": len(seed_data["patients"]),
        "module_families_seeded": len(support["supported_and_seeded"]),
        "synthetic_check_ok": synthetic["ok"],
        "unicode_hospital_name": seed_data["organization"]["name"],
        "unicode_hospital_name_utf8_sha256": hashlib.sha256(seed_data["organization"]["name"].encode("utf-8")).hexdigest(),
        "karthik_patient_present": any(patient.get("given_name") == "Karthik" and patient.get("identifier") == "SYN-HEN-0009" for patient in seed_data["patients"]),
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
    for patient in seed_data["patients"]:
        patient_hits = api.results("/patient", q=patient["identifier"], v="default", limit=5)
        if patient_hits:
            counts["patients"] += 1
            patient_uuid = patient_hits[0]["uuid"]
            counts["orders"] += len(api.results("/order", patient=patient_uuid, v="default", limit=100))
            counts["encounters"] += len(api.results("/encounter", patient=patient_uuid, v="default", limit=100))
            try:
                _, allergies = api.rest(f"/patient/{patient_uuid}/allergy", params={"v": "full"})
                counts["allergies"] += len(allergies.get("results", [])) if isinstance(allergies, dict) else 0
            except ApiError:
                # Some Bahmni/OpenMRS builds return 500 for selected allergy
                # projections. The seed path uses the supported create/list
                # shape; verification keeps going and reports other endpoints.
                pass
            fhir_id = fhir_patient_id(api, patient["identifier"])
            if fhir_id:
                _, condition_bundle = api.fhir("/Condition", params={"patient": fhir_id})
                counts["conditions"] += int(condition_bundle.get("total", 0) or 0)
    for provider in seed_data["providers"]:
        if api.results("/provider", q=provider["identifier"], v="default", limit=5):
            counts["providers"] += 1
    _, appointments = api.rest("/appointments")
    appointment_count = 0
    if isinstance(appointments, list):
        appointment_count = sum(1 for item in appointments if str(item.get("comments", "")).startswith(seed_data["records"]["appointment_comment_prefix"]))
    _, fhir_patients = api.fhir("/Patient", params={"identifier": "SYN-HEN"})
    _, fhir_observations = api.fhir("/Observation", params={"code": seed_data["concepts"]["weight_kg"]})

    unicode_location_present = any(
        item.get("display") == seed_data["organization"]["name"]
        for item in api.results("/location", q=seed_data["organization"]["name"], v="default", limit=20)
    )
    karthik_live: dict[str, Any] = {"patient_found": False, "completed_visit_found": False, "cold_fever_conditions_found": False, "fever_observation_found": False}
    karthik_hits = api.results("/patient", q="SYN-HEN-0009", v="default", limit=5)
    if karthik_hits:
        karthik_live["patient_found"] = True
        karthik_uuid = karthik_hits[0]["uuid"]
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

    live_ok = counts["patients"] == len(seed_data["patients"]) and counts["providers"] == len(seed_data["providers"])
    live_ok = live_ok and unicode_location_present and all(karthik_live.values())
    result = {
        "ok": live_ok,
        "offline": offline,
        "live_counts": counts | {"appointments": appointment_count},
        "unicode_location_present": unicode_location_present,
        "karthik_live": karthik_live,
        "fhir": {
            "patient_search_total": fhir_patients.get("total"),
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

    args = parser.parse_args(argv)
    args.func(args)


if __name__ == "__main__":
    main()
