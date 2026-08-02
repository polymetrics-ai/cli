from __future__ import annotations

import argparse
import importlib.util
import io
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any
from unittest import mock


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
            "+919876543210",
            "+911123456789",
            "+91 (0) 80 2345 6789",
            "00911123456789",
            "911123456789",
            "98765-43210",
            "+91 (80) 2345 6789",
            "080 2345 6789",
            "08023456789",
            "+91–98765–43210",
            "+91/98765/43210",
            "+٩١–٩٨٧٦٥–٤٣٢١٠",
            "care.team@example.com",
            "Please contact care.team@example.com.",
            "care.team@fm-bahmni.invalid.example.",
        ]
        for sample in samples:
            with self.subTest(sample=sample):
                self.assertTrue(labctl.contact_failures({"note": sample}))

    def test_allows_reserved_invalid_email_and_invalid_phone_placeholder(self) -> None:
        value = {"email": "Write to patient@fm-bahmni.invalid.", "phone": "000-000-0000"}
        self.assertEqual([], labctl.contact_failures(value))

    def test_committed_fixture_passes_recursive_contact_guard(self) -> None:
        self.assertEqual([], labctl.contact_failures(labctl.load_json(labctl.SEED_PATH)))

    def test_outgoing_api_body_is_guarded_before_network_access(self) -> None:
        credentials = {
            "urls": {"openmrs_rest": "https://127.0.0.1:18443/openmrs/ws/rest/v1"},
            "openmrs": {"username": "admin", "password": "redacted"},
        }
        with mock.patch.dict(os.environ, {"BAHMNI_LAB_HTTPS_PORT": "18443"}), mock.patch.object(
            labctl, "read_credentials", return_value=credentials
        ):
            api = labctl.BahmniApi()
            with self.assertRaises(labctl.ApiError):
                api.rest("/obs", "POST", {"value": "Call +91 98765 43210"})


class ApiOriginGuardTests(unittest.TestCase):
    def credentials(self, rest: str = "https://127.0.0.1:18443/openmrs/ws/rest/v1") -> dict[str, Any]:
        return {
            "urls": {
                "openmrs_rest": rest,
                "openmrs_fhir": "https://example.com/openmrs/ws/fhir2/R4",
            },
            "openmrs": {"username": "admin", "password": "redacted"},
        }

    def test_accepts_exact_configured_ipv4_ipv6_and_localhost_rest_origins(self) -> None:
        bases = [
            "https://127.0.0.1:18443/openmrs/ws/rest/v1",
            "https://localhost:18443/openmrs/ws/rest/v1",
            "https://[::1]:18443/openmrs/ws/rest/v1",
        ]
        with mock.patch.dict(os.environ, {"BAHMNI_LAB_HTTPS_PORT": "18443"}), mock.patch.object(
            labctl, "read_credentials", return_value=self.credentials()
        ):
            for base in bases:
                with self.subTest(base=base):
                    api = labctl.BahmniApi(base)
                    self.assertEqual(base, api.rest_base)
                    self.assertEqual(labctl.derive_fhir_base(base), api.fhir_base)

    def test_rejects_remote_unsafe_ambiguous_and_unexpected_rest_origins_before_credentials(self) -> None:
        bases = [
            "http://127.0.0.1:18443/openmrs/ws/rest/v1",
            "https://example.com:18443/openmrs/ws/rest/v1",
            "https://127.0.0.2:18443/openmrs/ws/rest/v1",
            "https://admin:secret@127.0.0.1:18443/openmrs/ws/rest/v1",
            "https://127.0.0.1:18444/openmrs/ws/rest/v1",
            "https://127.0.0.1:18443/openmrs/ws/rest/v1?target=other",
            "https://127.0.0.1:18443/openmrs/ws/rest/v1#other",
            "https://127.0.0.1:18443/openmrs/ws/fhir2/R4",
            "//127.0.0.1:18443/openmrs/ws/rest/v1",
        ]
        with mock.patch.dict(os.environ, {"BAHMNI_LAB_HTTPS_PORT": "18443"}):
            for base in bases:
                with self.subTest(base=base), mock.patch.object(labctl, "read_credentials") as read:
                    with self.assertRaises(labctl.ApiError):
                        labctl.BahmniApi(base)
                    read.assert_not_called()

    def test_rejects_remote_rest_origin_loaded_from_credentials(self) -> None:
        with mock.patch.dict(os.environ, {"BAHMNI_LAB_HTTPS_PORT": "18443"}), mock.patch.object(
            labctl, "read_credentials", return_value=self.credentials("https://example.com:18443/openmrs/ws/rest/v1")
        ):
            with self.assertRaises(labctl.ApiError):
                labctl.BahmniApi()

    def test_authenticated_redirect_is_refused(self) -> None:
        credentials = self.credentials()
        with mock.patch.dict(os.environ, {"BAHMNI_LAB_HTTPS_PORT": "18443"}), mock.patch.object(
            labctl, "read_credentials", return_value=credentials
        ):
            api = labctl.BahmniApi()
        handler = next(item for item in api.opener.handlers if isinstance(item, labctl.RejectAuthenticatedRedirects))

        class RedirectingOpener:
            def open(self, request: urllib.request.Request, timeout: int) -> Any:
                raise urllib.error.HTTPError(
                    request.full_url,
                    302,
                    "Found",
                    {"Location": "https://example.com/steal"},
                    io.BytesIO(b""),
                )

        api.opener = RedirectingOpener()
        with self.assertRaisesRegex(labctl.ApiError, "authenticated redirect"):
            api.rest("/session")

        followup = handler.redirect_request(
            urllib.request.Request(api.rest_base + "/session", headers={"Authorization": api.auth_header}),
            None,
            302,
            "Found",
            {},
            "https://example.com/steal",
        )
        self.assertIsNone(followup)

    def test_authenticated_client_disables_environment_and_system_proxies(self) -> None:
        credentials = self.credentials()
        with mock.patch.dict(os.environ, {
            "BAHMNI_LAB_HTTPS_PORT": "18443",
            "HTTPS_PROXY": "https://proxy.example:8443",
            "ALL_PROXY": "socks5://proxy.example:1080",
        }), mock.patch.object(
            labctl, "read_credentials", return_value=credentials
        ), mock.patch.object(
            urllib.request, "getproxies", side_effect=AssertionError("system proxy discovery must not run")
        ):
            api = labctl.BahmniApi()

        self.assertEqual({}, api.proxy_handler.proxies)


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
                "patient": {"uuid": "existing-patient", "identifier": "SYN-HEN-0001"},
                "service": {"uuid": "existing-service"},
                "serviceType": {"uuid": "existing-service-type"},
                "location": {"uuid": "existing-location"},
                "appointmentKind": "Virtual",
                "status": "CheckedIn",
                "comments": old_comment,
                "startDateTime": labctl.appointment_ms(start),
                "endDateTime": labctl.appointment_ms("2026-01-20T09:45:00.000+0000"),
                "providers": [],
                }]
            return 200, {}

        api.rest = appointment_rest
        counts = summary()

        labctl.ensure_appointment(
            api, "SYN-HEN-0001", "patient-1", "provider-1", "service-1", "location-1",
            start, "2026-01-20T09:15:00.000+0000", current_comment, stable, counts,
            legacy_comments=[old_comment],
        )

        updates = [call for call in api.calls if call[0:2] == ("POST", "/appointment")]
        self.assertEqual(1, len(updates))
        update = updates[0][2]
        self.assertEqual("appointment-1", update["uuid"])
        self.assertEqual("CheckedIn", update["status"])
        self.assertEqual("Virtual", update["appointmentKind"])
        self.assertEqual("existing-service", update["serviceUuid"])
        self.assertEqual("existing-service-type", update["serviceTypeUuid"])
        self.assertEqual("existing-location", update["locationUuid"])
        self.assertEqual([], update["providers"])
        self.assertIn(stable, update["comments"])
        self.assertFalse(any(call[0:2] == ("POST", "/appointments/appointment-1") for call in api.calls))
        self.assertFalse(any(call[0:2] == ("POST", "/appointments") for call in api.calls))
        self.assertEqual(1, counts["appointments_reconciled"])

    def test_appointment_with_unrepresentable_provider_state_is_not_updated(self) -> None:
        appointment = {
            "uuid": "appointment-1",
            "patient": {"uuid": "patient-1"},
            "service": {"uuid": "service-1"},
            "location": {"uuid": "location-1"},
            "appointmentKind": "Scheduled",
            "status": "Scheduled",
            "startDateTime": 1,
            "endDateTime": 2,
            "providers": [{"uuid": "provider-1", "response": "ACCEPTED", "voided": True}],
        }

        with self.assertRaisesRegex(labctl.ApiError, "provider associations"):
            labctl.appointment_update_body(appointment, "new comment")

    def test_legacy_allergy_comment_is_reconciled_through_patient_subresource(self) -> None:
        old_comment = "FM-BAHMNI-LAB-R1 synthetic low-severity medication allergy; not real patient data SYN-HEN-0001"
        new_comment = "Low-severity medication allergy noted in local connector lab record SYN-HEN-0001"
        api = FakeApi()

        def allergy_rest(path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None) -> tuple[int, Any]:
            api.calls.append((method, path, body, params))
            if method == "GET":
                return 200, {"results": [{
                    "uuid": "allergy-1",
                    "allergen": {"codedAllergen": {"uuid": "penicillin"}},
                    "comment": old_comment,
                }]}
            return 200, {}

        api.rest = allergy_rest
        counts = summary()

        labctl.ensure_allergy(
            api,
            "patient-1",
            "penicillin",
            new_comment,
            counts,
            legacy_comments=[old_comment],
        )

        self.assertIn(("POST", "/patient/patient-1/allergy/allergy-1", {"comment": new_comment}, None), api.calls)
        self.assertEqual(1, counts["allergies_reconciled"])
        self.assertFalse(any(call[0:2] == ("POST", "/patient/patient-1/allergy") for call in api.calls))

    def test_unrecognized_allergy_comment_is_preserved_without_update(self) -> None:
        api = FakeApi()

        def allergy_rest(path: str, method: str = "GET", body: Any | None = None, params: dict[str, Any] | None = None) -> tuple[int, Any]:
            api.calls.append((method, path, body, params))
            return 200, {"results": [{
                "uuid": "allergy-1",
                "allergen": {"codedAllergen": {"uuid": "penicillin"}},
                "comment": "Clinician-authored comment",
            }]}

        api.rest = allergy_rest
        counts = summary()
        labctl.ensure_allergy(
            api,
            "patient-1",
            "penicillin",
            "Expected task comment",
            counts,
            legacy_comments=["Exact old task comment"],
        )

        self.assertFalse(any(call[0] == "POST" for call in api.calls))
        self.assertEqual(1, counts["allergies_existing"])


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

    def test_ownership_marker_survives_runtime_cleanup_and_same_resource_restart(self) -> None:
        with tempfile.TemporaryDirectory(prefix=".ownership-", dir=LAB_ROOT) as parent:
            home = Path(parent) / labctl.LAB_ID
            environment = {
                "BAHMNI_LAB_HOME": str(home),
                "BAHMNI_LAB_MACHINE": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_CONNECTION": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_PROJECT": "fm_bahmni_lab_r1",
                "XDG_STATE_HOME": str(Path(parent) / "state"),
            }
            with mock.patch.dict(os.environ, environment), mock.patch.object(labctl.shutil, "which", return_value=None):
                claimed = labctl.claim_ownership()
                marker_path = labctl.ownership_marker_path()
                self.assertTrue(marker_path.is_file())
                self.assertEqual(Path(parent) / "state" / labctl.LAB_ID / "ownership.json", marker_path)
                self.assertEqual(0o700, marker_path.parent.stat().st_mode & 0o777)
                self.assertEqual(0o600, marker_path.stat().st_mode & 0o777)
                shutil.rmtree(home)
                self.assertEqual(claimed, labctl.load_ownership_marker())
                self.assertEqual(claimed, labctl.claim_ownership())

    def test_legacy_runtime_marker_migrates_before_lifecycle_verification(self) -> None:
        with tempfile.TemporaryDirectory(prefix=".ownership-upgrade-", dir=LAB_ROOT) as parent:
            home = Path(parent) / labctl.LAB_ID
            home.mkdir()
            environment = {
                "BAHMNI_LAB_HOME": str(home),
                "BAHMNI_LAB_MACHINE": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_CONNECTION": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_PROJECT": "fm_bahmni_lab_r1",
            }
            environment["XDG_STATE_HOME"] = str(Path(parent) / "state")
            with mock.patch.dict(os.environ, environment):
                expected = labctl.expected_ownership()
                legacy_path = labctl.legacy_ownership_marker_paths()[0]
                legacy_marker = dict(expected, schema=1)
                legacy_marker.pop("resources")
                legacy_path.write_text(json.dumps(legacy_marker), encoding="utf-8")
                legacy_path.chmod(0o600)
                self.assertEqual(expected, labctl.load_ownership_marker())
                self.assertTrue(labctl.ownership_marker_path().is_file())

    def test_existing_machine_without_durable_marker_requires_explicit_recovery(self) -> None:
        with tempfile.TemporaryDirectory(prefix=".ownership-refuse-", dir=LAB_ROOT) as parent:
            environment = {
                "BAHMNI_LAB_HOME": str(Path(parent) / labctl.LAB_ID),
                "BAHMNI_LAB_MACHINE": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_CONNECTION": "fm-bahmni-lab-r1-machine",
                "BAHMNI_LAB_PROJECT": "fm_bahmni_lab_r1",
                "XDG_STATE_HOME": str(Path(parent) / "state"),
            }
            with mock.patch.dict(os.environ, environment), mock.patch.object(
                labctl.shutil, "which", return_value="/usr/bin/podman"
            ), mock.patch.object(
                labctl, "podman_names", side_effect=[{"fm-bahmni-lab-r1-machine"}, {"fm-bahmni-lab-r1-machine"}]
            ):
                with self.assertRaisesRegex(ValueError, "explicit ownership recovery"):
                    labctl.claim_ownership()

    def test_task_prefixed_resource_without_project_label_is_rejected(self) -> None:
        with self.assertRaises(ValueError):
            labctl.validate_owned_resources([{
                "Names": ["fm-bahmni-lab-r1-openmrs"],
                "Labels": {},
            }], "fm_bahmni_lab_r1", "container")

    def test_project_resource_without_explicit_owner_label_is_rejected(self) -> None:
        resource = {
            "Names": ["fm_bahmni_lab_r1_openmrs_1"],
            "Id": "container-id",
            "Labels": {"com.docker.compose.project": "fm_bahmni_lab_r1"},
        }
        with self.assertRaisesRegex(ValueError, "ownership label"):
            labctl.validate_owned_resources([resource], "fm_bahmni_lab_r1", "container")

    def test_network_or_volume_requires_label_or_exact_durable_identity(self) -> None:
        resource = {
            "Name": "fm_bahmni_lab_r1_data",
            "Id": "resource-id",
            "Labels": {"com.docker.compose.project": "fm_bahmni_lab_r1"},
        }
        with self.assertRaisesRegex(ValueError, "ownership proof"):
            labctl.validate_owned_resources([resource], "fm_bahmni_lab_r1", "volume")
        labctl.validate_owned_resources(
            [resource],
            "fm_bahmni_lab_r1",
            "volume",
            [{"kind": "volume", "name": "fm_bahmni_lab_r1_data", "id": "resource-id"}],
        )

    def test_generated_compose_labels_services_volumes_and_default_network(self) -> None:
        source = "services:\n  openmrs:\n    image: example/openmrs\nvolumes:\n  openmrs-data:\n"
        rendered = labctl.scope_compose_ownership(source)
        labctl.validate_compose_ownership(rendered)
        self.assertEqual(3, rendered.count(labctl.OWNER_LABEL))

    def test_forget_refuses_while_owned_machine_or_connection_exists(self) -> None:
        marker = labctl.validate_ownership_config(
            Path("/tmp/fm-bahmni-lab-r1"),
            "fm-bahmni-lab-r1-machine",
            "fm-bahmni-lab-r1-machine",
            "fm_bahmni_lab_r1",
        )
        inventories = [
            {"fm-bahmni-lab-r1-machine"},
            set(),
        ]
        with mock.patch.object(labctl, "load_ownership_marker", return_value=marker), mock.patch.object(
            labctl.shutil, "which", return_value="/usr/bin/podman"
        ), mock.patch.object(labctl, "podman_names", side_effect=inventories):
            with self.assertRaisesRegex(ValueError, "still exists"):
                labctl.forget_ownership()

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
