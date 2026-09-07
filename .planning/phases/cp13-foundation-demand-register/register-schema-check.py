"""Closed CP13 register wire checks using pinned local schemas; no retrieval."""

import argparse
import copy
import hashlib
import json
from importlib.metadata import version
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource

parser = argparse.ArgumentParser()
parser.add_argument("--document", type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
schema_path = root / "docs/connector-canon/foundations/demand-register.schema.json"
paths = [schema_path, root / "docs/connector-canon/foundations/assessments.schema.json",
         root / "docs/connector-canon/foundations/proofs.schema.json",
         root / "docs/connector-canon/source-lane-manifest.schema.json"]
resources, pins = [], []
for path in paths:
    raw = path.read_bytes()
    schema = json.loads(raw)
    Draft202012Validator.check_schema(schema)
    resources.append((path.as_uri(), Resource.from_contents(schema)))
    pins.append({"path": str(path.relative_to(root)), "sha256": hashlib.sha256(raw).hexdigest(), "bytes": len(raw)})
registry = Registry().with_resources(resources)
validator = Draft202012Validator({"$ref": schema_path.as_uri()}, registry=registry)
raw = args.document.read_bytes()
document = json.loads(raw)
print(json.dumps({"document": str(args.document), "sha256": hashlib.sha256(raw).hexdigest(), "bytes": len(raw), "schemas": pins}))

cases, failed = 0, 0

def check(name, candidate, expected):
    global cases, failed
    errors = list(validator.iter_errors(candidate))
    actual = not errors
    passed = actual == expected
    cases += 1
    failed += not passed
    result = {"case": name, "expected_valid": expected, "actual_valid": actual, "pass": passed}
    if not passed:
        result["errors"] = [{"path": list(e.absolute_path), "message": e.message} for e in errors[:3]]
    print(json.dumps(result), flush=True)

check("original generated register", document, True)
for field in document:
    for mutation in ("absent", "null"):
        candidate = copy.deepcopy(document)
        if mutation == "absent":
            del candidate[field]
        else:
            candidate[field] = None
        check(f"root {field} {mutation}", candidate, False)
for container, field in (("checks", "selected_reuse_requirements"), ("known_obligations", "sentry_registration")):
    for mutation in ("absent", "null"):
        candidate = copy.deepcopy(document)
        target = candidate[container] if container == "checks" else candidate[container][0]
        if mutation == "absent":
            del target[field]
        else:
            target[field] = None
        check(f"required zero {field} {mutation}", candidate, False)
for location in ("root", "requirement", "lookup", "facet", "known", "proof", "input", "example", "checks"):
    candidate = copy.deepcopy(document)
    targets = {"root": candidate, "requirement": candidate["requirements"][0],
               "lookup": candidate["requirements"][0]["atlas_lookup"], "facet": candidate["source_fit"][0],
               "known": candidate["known_obligations"][0], "proof": candidate["foundation_proof_observations"][0]["record"],
               "input": candidate["inputs"][0], "example": candidate["atlas_examples"][0], "checks": candidate["checks"]}
    targets[location]["runtime_authority"] = True
    check(f"closed {location}", candidate, False)
if args.document.read_bytes() != raw:
    raise SystemExit("input document changed during schema observation")
for pin, path in zip(pins, paths):
    if hashlib.sha256(path.read_bytes()).hexdigest() != pin["sha256"]:
        raise SystemExit("schema changed during observation")
print(json.dumps({"cases": cases, "failed": failed, "jsonschema": version("jsonschema"), "referencing": version("referencing")}))
raise SystemExit(bool(failed))
