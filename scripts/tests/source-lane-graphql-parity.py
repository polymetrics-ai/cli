#!/usr/bin/env python3
"""Check published GraphQL annotation schema against the actual Go reader.

Requires an installed Python jsonschema Draft202012Validator; does not install
packages or change the runtime/Go dependency set.
"""
import json
from pathlib import Path
import subprocess
import sys

from jsonschema import Draft202012Validator


def main():
    root = Path(__file__).resolve().parents[2]
    cases = json.loads((root / "cmd/connectorgen/testdata/source-lane-graphql-139.json").read_text())
    schema = json.loads((root / "docs/connector-canon/source-lane-manifest.schema.json").read_text())
    Draft202012Validator.check_schema(schema)
    validator = Draft202012Validator({
        "$schema": schema["$schema"], "$defs": schema["$defs"],
        "$ref": "#/$defs/source_annotation",
    })
    command = ["go", "test", "-json", "-count=1", "-timeout", "20m",
               "./cmd/connectorgen", "-run", "^TestSourceLane139GraphQLSchemaReader$"]
    result = subprocess.run(command, cwd=root, capture_output=True, text=True, check=False)
    sys.stdout.write(result.stdout)
    sys.stderr.write(result.stderr)
    events = [json.loads(line) for line in result.stdout.splitlines()]
    prefix = "TestSourceLane139GraphQLSchemaReader/"
    passed = {e.get("Test") for e in events if e.get("Action") == "pass" and e.get("Test", "").startswith(prefix)}
    expected = {prefix + case["name"].replace(" ", "_") for case in cases}
    failed = result.returncode != 0 or len(cases) != 11 or passed != expected
    for case in cases:
        errors = list(validator.iter_errors(case["annotation"]))
        valid = not errors
        failed |= valid != case["valid"]
        print(json.dumps({"case": case["name"], "schema_valid": valid,
                          "expected": case["valid"], "errors": [e.message for e in errors]}))
    print(json.dumps({"reader_exit": result.returncode, "reader_cases": sorted(passed),
                      "schema_cases": len(cases), "parity": not failed}))
    return int(failed)


if __name__ == "__main__":
    raise SystemExit(main())
