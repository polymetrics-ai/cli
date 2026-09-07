"""Local CP13 schema evidence; uses the already installed validator, no network."""

import copy
import json
from importlib.metadata import version
from pathlib import Path

from jsonschema import Draft202012Validator
from referencing import Registry, Resource


root = Path(__file__).resolve().parents[3]
schema_path = root / "docs/connector-canon/foundations/assessments.schema.json"
source_schema_path = root / "docs/connector-canon/source-lane-manifest.schema.json"
resources = []
for path in (schema_path, source_schema_path):
    schema = json.loads(path.read_bytes())
    Draft202012Validator.check_schema(schema)
    resources.append((path.as_uri(), Resource.from_contents(schema)))
# No retrieval callback is installed. Only the two exact local schemas above
# can resolve; remote URLs and arbitrary files are not fetched by this check.
registry = Registry().with_resources(resources)
validator = Draft202012Validator({"$ref": schema_path.as_uri()}, registry=registry)
document = json.loads(
    (root / "data/connector-canon/batch1-foundation-assessments.json").read_bytes()
)
cases = [("current canonical assessments", document, True)]
for field in (
    "id",
    "source_refs",
    "statement",
    "atlas_lookup",
    "assessment",
    "proof_ids",
    "affected_artifacts",
    "evidence_requirements",
    "decision_refs",
):
    for mutation in ("absent", "null"):
        candidate = copy.deepcopy(document)
        requirement = candidate["assessments"][0]["requirements"][0]
        if mutation == "absent":
            del requirement[field]
        else:
            requirement[field] = None
        cases.append((f"{field} {mutation}", candidate, False))
for location in ("document", "requirement", "lookup", "candidate"):
    candidate = copy.deepcopy(document)
    requirement = candidate["assessments"][0]["requirements"][0]
    targets = {
        "document": candidate,
        "requirement": requirement,
        "lookup": requirement["atlas_lookup"],
        "candidate": requirement["atlas_lookup"]["candidates"][0],
    }
    targets[location]["runtime_authority"] = True
    cases.append((f"unknown {location} field", candidate, False))
for field in ("fit_bindings", "provider_clause", "source_exclusion"):
    candidate = copy.deepcopy(document)
    candidate["assessments"][0]["requirements"][0][field] = None
    cases.append((f"optional {field} null matches Go pointer/slice", candidate, True))

# The pending receiver's required condition has a legitimate empty value.
# Exercise the real complete document rather than validating a copied grammar.
for mutation in ("explicit_empty", "absent", "null", "unknown_field"):
    candidate = copy.deepcopy(document)
    decisions = [
        decision
        for assessment in candidate["assessments"]
        for requirement in assessment["requirements"]
        for decision in requirement["decision_refs"]
        if decision["id"] == "cli-batch1-vercel-inbound-sync-decision-r1"
    ]
    assert decisions, "actual pending receiver decision control missing"
    for decision in decisions:
        assert decision["state"] == "pending" and decision["condition"] == ""
        if mutation == "absent":
            del decision["condition"]
        elif mutation == "null":
            decision["condition"] = None
        elif mutation == "unknown_field":
            decision["approved"] = True
    cases.append((f"pending decision condition {mutation}", candidate, mutation == "explicit_empty"))

failed = 0
for name, candidate, expected in cases:
    actual = validator.is_valid(candidate)
    passed = actual == expected
    failed += not passed
    print(json.dumps({"case": name, "expected_valid": expected, "actual_valid": actual, "pass": passed}))
print(json.dumps({"cases": len(cases), "failed": failed, "jsonschema": version("jsonschema"), "referencing": version("referencing")}))
raise SystemExit(bool(failed))
