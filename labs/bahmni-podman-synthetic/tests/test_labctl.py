from __future__ import annotations

import argparse
import importlib.util
import os
import subprocess
import sys
import unittest
from pathlib import Path
from typing import Any


LAB_ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location("bahmni_labctl", LAB_ROOT / "lib" / "labctl.py")
assert SPEC and SPEC.loader
labctl = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = labctl
SPEC.loader.exec_module(labctl)


class FakeApi:
    def __init__(self, results: dict[str, list[dict[str, Any]]] | None = None):
        self.result_values = results or {}
        self.calls: list[tuple[str, str, Any, Any]] = []

    def results(self, path: str, **params: Any) -> list[dict[str, Any]]:
        self.calls.append(("RESULTS", path, None, params))
        return self.result_values.get(path, [])

    def rest(
        self,
        path: str,
        method: str = "GET",
        body: Any | None = None,
        params: dict[str, Any] | None = None,
    ) -> tuple[int, Any]:
        self.calls.append((method, path, body, params))
        return 200, {"uuid": "created"}


def summary() -> dict[str, int]:
    return labctl.new_summary()


class ContactGuardTests(unittest.TestCase):
    def test_rejects_formatted_indian_contacts_and_external_email(self) -> None:
        samples = [
            "+91 98765 43210",
            "98765-43210",
            "+91 (80) 2345 6789",
            "080 2345 6789",
            "08023456789",
            "care.team@example.com",
        ]
        for sample in samples:
            with self.subTest(sample=sample):
                self.assertTrue(labctl.contact_failures({"note": sample}))

    def test_allows_reserved_invalid_email_and_invalid_phone_placeholder(self) -> None:
        value = {"email": "patient@fm-bahmni.invalid", "phone": "000-000-0000"}
        self.assertEqual([], labctl.contact_failures(value))

    def test_committed_fixture_passes_recursive_contact_guard(self) -> None:
        self.assertEqual([], labctl.contact_failures(labctl.load_json(labctl.SEED_PATH)))

    def test_outgoing_api_body_is_guarded_before_network_access(self) -> None:
        api = object.__new__(labctl.BahmniApi)
        api.auth_header = "Basic redacted"
        api.ctx = None
        with self.assertRaises(labctl.ApiError):
            api.request("https://127.0.0.1", "/obs", "POST", {"value": "Call +91 98765 43210"})


class NameReconciliationTests(unittest.TestCase):
    def test_provider_name_updates_only_after_exact_identifier_match(self) -> None:
        api = FakeApi({
            "/provider": [{
                "uuid": "provider-1",
                "identifier": "SYN-PROV-CARD-001",
                "display": "SYN-PROV-CARD-001 - Asha Demoheart",
                "person": {
                    "uuid": "person-1",
                    "preferredName": {
                        "uuid": "name-1",
                        "givenName": "Asha",
                        "familyName": "Demoheart",
                    },
                },
            }],
        })
        counts = summary()

        result = labctl.ensure_provider(api, {
            "identifier": "SYN-PROV-CARD-001",
            "given_name": "Aarav",
            "family_name": "Raman",
            "gender": "F",
            "birthdate": "1979-04-12",
        }, counts)

        self.assertEqual("provider-1", result)
        self.assertIn(("POST", "/person/person-1/name/name-1", {
            "givenName": "Aarav",
            "familyName": "Raman",
            "preferred": True,
        }, None), api.calls)
        self.assertEqual(1, counts["providers_reconciled"])
        self.assertFalse(any(call[0:2] == ("POST", "/person") for call in api.calls))

    def test_identifier_substring_does_not_prove_patient_ownership(self) -> None:
        api = FakeApi({
            "/patient": [{
                "uuid": "other",
                "identifier": "SYN-HEN-00010",
                "display": "SYN-HEN-00010 - Unrelated Record",
            }],
        })
        counts = summary()

        labctl.ensure_patient(api, {
            "identifier": "SYN-HEN-0001",
            "given_name": "Meera",
            "family_name": "Rao",
            "gender": "F",
            "birthdate": "1988-03-14",
        }, "pid-type", "location", counts)

        self.assertTrue(any(call[0:2] == ("POST", "/patient") for call in api.calls))
        self.assertEqual(1, counts["patients_created"])

    def test_patient_preferred_name_is_reconciled_in_place(self) -> None:
        api = FakeApi({
            "/patient": [{
                "uuid": "patient-1",
                "identifiers": [{"identifier": "SYN-HEN-0009"}],
                "display": "SYN-HEN-0009 - Karthik Syntheticcase",
                "person": {
                    "uuid": "patient-1",
                    "preferredName": {
                        "uuid": "name-9",
                        "givenName": "Karthik",
                        "familyName": "Syntheticcase",
                    },
                },
            }],
        })
        counts = summary()

        result = labctl.ensure_patient(api, {
            "identifier": "SYN-HEN-0009",
            "given_name": "Karthik",
            "family_name": "Iyer",
            "gender": "M",
            "birthdate": "1991-05-17",
        }, "pid-type", "location", counts)

        self.assertEqual("patient-1", result)
        self.assertIn(("POST", "/person/patient-1/name/name-9", {
            "givenName": "Karthik",
            "familyName": "Iyer",
            "preferred": True,
        }, None), api.calls)
        self.assertEqual(1, counts["patients_reconciled"])


class StableRecordUpgradeTests(unittest.TestCase):
    def test_legacy_encounter_is_canonicalized_without_replacement(self) -> None:
        legacy = "FM-BAHMNI-LAB-R1 synthetic encounter SYN-HEN-0001 consultation"
        stable = labctl.stable_record_marker("encounter", "SYN-HEN-0001", "consultation")
        api = FakeApi({
            "/encounter": [{
                "uuid": "encounter-1",
                "obs": [{"uuid": "obs-1", "value": legacy}],
            }],
        })
        counts = summary()

        result = labctl.ensure_encounter(
            api,
            "patient-1",
            "encounter-type",
            "2026-01-15T09:15:00.000+0000",
            "location",
            "visit",
            [],
            stable,
            "document-concept",
            counts,
            legacy_markers=[legacy],
        )

        self.assertEqual("encounter-1", result)
        self.assertIn(("POST", "/obs/obs-1", {"value": stable}, None), api.calls)
        self.assertFalse(any(call[0:2] == ("POST", "/encounter") for call in api.calls))
        self.assertEqual(1, counts["encounters_reconciled"])

    def test_ambiguous_legacy_encounters_are_preserved_but_not_canonical(self) -> None:
        legacy = "FM-BAHMNI-LAB-R1 encounter SYN-HEN-0001 consultation"
        stable = labctl.stable_record_marker("encounter", "SYN-HEN-0001", "consultation")
        api = FakeApi({
            "/encounter": [
                {"uuid": "encounter-b", "obs": [{"uuid": "obs-b", "value": legacy}]},
                {"uuid": "encounter-a", "obs": [{"uuid": "obs-a", "value": legacy}]},
            ],
        })
        counts = summary()

        labctl.ensure_encounter(
            api, "patient-1", "encounter-type", "when", "location", "visit", [],
            stable, "document-concept", counts, legacy_markers=[legacy],
        )

        self.assertIn(("POST", "/obs/obs-a", {"value": stable}, None), api.calls)
        self.assertIn(("POST", "/obs/obs-b", {
            "value": f"{stable}:preserved:encounter-b",
        }, None), api.calls)
        self.assertFalse(any(call[0] == "DELETE" for call in api.calls))
        self.assertFalse(any(call[0:2] == ("POST", "/encounter") for call in api.calls))
        self.assertEqual(1, counts["encounter_duplicates_preserved"])

    def test_legacy_appointment_comment_is_updated_in_place(self) -> None:
        start = "2026-01-20T09:00:00.000+0000"
        old_comment = "FM-BAHMNI-LAB-R1 synthetic appointment SYN-HEN-0001"
        current_comment = "Cardiology follow-up"
        stable = labctl.stable_record_marker("appointment", "SYN-HEN-0001", str(labctl.appointment_ms(start)))
        api = FakeApi()

        def appointment_rest(path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None) -> tuple[int, Any]:
            api.calls.append((method, path, body, params))
            if path == "/appointments" and method == "GET":
                return 200, [{
                "uuid": "appointment-1",
                "patient": {"identifier": "SYN-HEN-0001"},
                "comments": old_comment,
                "startDateTime": labctl.appointment_ms(start),
                }]
            return 200, {}

        api.rest = appointment_rest
        counts = summary()

        labctl.ensure_appointment(
            api, "SYN-HEN-0001", "patient-1", "provider-1", "service-1", "location-1",
            start, "2026-01-20T09:15:00.000+0000", current_comment, stable, counts,
            legacy_comments=[old_comment],
        )

        updates = [call for call in api.calls if call[0:2] == ("POST", "/appointments/appointment-1")]
        self.assertEqual(1, len(updates))
        self.assertIn(stable, updates[0][2]["comments"])
        self.assertFalse(any(call[0:2] == ("POST", "/appointments") for call in api.calls))
        self.assertEqual(1, counts["appointments_reconciled"])


class OwnershipGuardTests(unittest.TestCase):
    def test_rejects_unrelated_resource_names(self) -> None:
        with self.assertRaises(ValueError):
            labctl.validate_ownership_config(
                Path("/tmp/fm-bahmni-lab-r1"),
                "unrelated-machine",
                "unrelated-connection",
                "unrelated_project",
            )

    def test_rejects_mismatched_durable_marker(self) -> None:
        expected = labctl.validate_ownership_config(
            Path("/tmp/fm-bahmni-lab-r1"),
            "fm-bahmni-lab-r1-machine",
            "fm-bahmni-lab-r1-machine",
            "fm_bahmni_lab_r1",
        )
        actual = dict(expected, project="fm_bahmni_lab_r1_other")
        with self.assertRaises(ValueError):
            labctl.validate_ownership_marker(actual, expected)

    def test_generated_compose_services_receive_owner_label(self) -> None:
        source = "services:\n  openmrs:\n    image: example/openmrs\n  proxy:\n    labels:\n      existing: value\n    image: example/proxy\n"
        rendered = labctl.scope_service_labels(source)
        labctl.validate_service_labels(rendered)
        self.assertEqual(2, rendered.count(labctl.OWNER_LABEL))

    def test_task_prefixed_resource_without_project_label_is_rejected(self) -> None:
        with self.assertRaises(ValueError):
            labctl.validate_owned_resources([{
                "Names": ["fm-bahmni-lab-r1-openmrs"],
                "Labels": {},
            }], "fm_bahmni_lab_r1")

    def test_unrelated_shell_override_fails_before_podman_action(self) -> None:
        env = os.environ.copy()
        env.update({
            "BAHMNI_LAB_HOME": str(LAB_ROOT / ".test-runtime" / labctl.LAB_ID),
            "BAHMNI_LAB_MACHINE": "unrelated-machine",
            "BAHMNI_LAB_CONNECTION": "unrelated-connection",
            "BAHMNI_LAB_PROJECT": "unrelated_project",
        })
        proc = subprocess.run(
            [str(LAB_ROOT / "bin" / "bahmni-lab"), "start"],
            env=env,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )
        self.assertNotEqual(0, proc.returncode)
        self.assertIn("ownership check failed", proc.stderr.lower())


class FocusedStaticChecks(unittest.TestCase):
    def test_shell_wrapper_parses(self) -> None:
        subprocess.run(["bash", "-n", str(LAB_ROOT / "bin" / "bahmni-lab")], check=True)

    def test_offline_synthetic_check_passes(self) -> None:
        result = labctl.check_synthetic(argparse.Namespace(json=False, online_source=False, quiet=True))
        self.assertTrue(result["ok"])


if __name__ == "__main__":
    unittest.main()
