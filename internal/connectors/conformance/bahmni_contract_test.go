package conformance

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestBahmniFrozenScopeTextContracts(t *testing.T) {
	b := loadTestBundle(t, "../defs", "bahmni")

	for label, text := range map[string]string{
		"metadata.description": b.Metadata.Description,
		"api_surface.scope":    b.Surface.Scope,
		"spec":                 string(b.RawSpec),
		"operations":           string(b.RawOperations),
		"cli_surface":          string(b.RawCLISurface),
		"docs":                 b.Docs,
	} {
		if strings.Contains(text, "STANDARD/LITE") || strings.Contains(text, "LITE") {
			t.Fatalf("%s still claims LITE support", label)
		}
	}

	appointments, ok := bahmniCommand(b, "appointments list")
	if !ok {
		t.Fatalf("appointments list command missing")
	}
	if strings.Contains(appointments.Summary, "patient_uuid") {
		t.Fatalf("appointments summary still claims patient_uuid scoping: %q", appointments.Summary)
	}

	patientSearch, ok := bahmniCommand(b, "bahmnicore patient-search")
	if !ok {
		t.Fatalf("bahmnicore patient-search command missing")
	}
	for label, text := range map[string]string{
		"approval": patientSearch.Approval,
		"risk":     patientSearch.Risk,
	} {
		if strings.Contains(text, "POST") || strings.Contains(text, "body") {
			t.Fatalf("patient-search %s still describes POST/body semantics: %q", label, text)
		}
	}
}

func TestBahmniVersionPinnedReadContracts(t *testing.T) {
	b := loadTestBundle(t, "../defs", "bahmni")

	patients := requireBahmniStream(t, b, "patients")
	patientQuery := patients.Query["q"]
	if patientQuery.OmitWhenAbsent {
		t.Fatalf("patients stream q query omits when absent; pinned OpenMRS REST patient does not support unqualified get-all")
	}

	labResults := requireBahmniStream(t, b, "lab_results")
	if _, ok := labResults.Query["concept"]; !ok {
		t.Fatalf("lab_results stream missing required concept query for BahmniObservationsController 1.2.1")
	}
	if labResults.Query["patientUuid"].OmitWhenAbsent {
		t.Fatalf("lab_results patientUuid query omits when absent; pinned controller requires patientUuid")
	}

	appointments := requireBahmniStream(t, b, "appointments")
	if _, ok := appointments.Query["patientUuid"]; ok {
		t.Fatalf("appointments stream exposes patientUuid query even though appointments 2.0.2 ignores it")
	}
	forDate := appointments.Query["forDate"]
	if forDate.Template == "" || forDate.OmitWhenAbsent {
		t.Fatalf("appointments stream must require exact forDate scoping instead of unbounded appointment book reads: %+v", forDate)
	}

	diagnoses := requireBahmniStream(t, b, "diagnoses")
	if diagnoses.Path != "/ws/rest/v1/bahmnicore/diagnosis/search" {
		t.Fatalf("diagnoses path = %q, want pinned BahmniDiagnosisController search route", diagnoses.Path)
	}
	if diagnoses.Query["patientUuid"].OmitWhenAbsent {
		t.Fatalf("diagnoses patientUuid query omits when absent; pinned diagnosis search requires patientUuid")
	}
}

func TestBahmniVersionPinnedDirectOperationContracts(t *testing.T) {
	b := loadTestBundle(t, "../defs", "bahmni")

	if cmd, ok := bahmniCommand(b, "bahmnicore patient-detail"); ok && cmd.Availability == "implemented" {
		t.Fatalf("bahmnicore patient-detail is implemented but pinned BahmniPatientProfileResource exposes POST create/update, not GET by UUID")
	}

	cmd, ok := bahmniCommand(b, "bahmnicore patient-search")
	if !ok {
		t.Fatalf("bahmnicore patient-search command missing")
	}
	if cmd.Intent != "direct_read" || cmd.Availability != "implemented" || cmd.Operation != "bahmni.patient_search" {
		t.Fatalf("patient-search command not implemented as typed operation: %+v", cmd)
	}
	for _, flag := range cmd.Flags {
		if hasPrefix(flag.MapsTo, "body.") {
			t.Fatalf("patient-search flag %s maps to %q; pinned bahmni-commons 1.1.0 route is GET query params", flag.Name, flag.MapsTo)
		}
	}

	op, ok := bahmniOperation(b, "bahmni.patient_search")
	if !ok {
		t.Fatalf("bahmni.patient_search operation missing")
	}
	if op.REST == nil || op.REST.Method != "GET" || op.REST.Path != "/ws/rest/v1/bahmni/search/patient" {
		t.Fatalf("patient_search REST = %+v, want GET /ws/rest/v1/bahmni/search/patient", op.REST)
	}
	if op.SensitivePolicy == nil || !containsStringSlice(op.SensitivePolicy.RedactFields, "identifier") || !containsStringSlice(op.SensitivePolicy.RedactFields, "addressFieldValue") {
		t.Fatalf("patient_search sensitive_policy must declare enforced redaction fields, got %+v", op.SensitivePolicy)
	}
}

func TestBahmniVersionPinnedWriteContracts(t *testing.T) {
	b := loadTestBundle(t, "../defs", "bahmni")

	for _, unsupported := range []string{"create_diagnosis", "create_observations_bulk", "upload_patient_document", "upload_visit_document", "create_drug_order", "reschedule_appointment"} {
		if action, ok := bahmniWrite(b, unsupported); ok {
			t.Fatalf("unsupported write %q still declared at %s %s", unsupported, action.Method, action.Path)
		}
	}

	want := map[string]string{
		"create_patient":                       "POST /ws/rest/v1/patient",
		"update_patient":                       "POST /ws/rest/v1/patient/{{ record.uuid }}",
		"create_encounter":                     "POST /ws/rest/v1/encounter",
		"create_observation":                   "POST /ws/rest/v1/obs",
		"create_visit":                         "POST /ws/rest/v1/visit",
		"create_lab_order":                     "POST /ws/rest/v1/order",
		"create_patient_diagnosis":             "POST /ws/rest/v1/patientdiagnoses",
		"create_appointment":                   "POST /ws/rest/v1/appointments",
		"update_appointment_status":            "POST /ws/rest/v1/appointments/{{ record.appointmentUuid }}/status-change",
		"update_appointment_provider_response": "POST /ws/rest/v1/appointments/{{ record.appointmentUuid }}/providerResponse",
		"create_note":                          "POST /ws/rest/v1/notes",
	}
	for name, requestLine := range want {
		action, ok := bahmniWrite(b, name)
		if !ok {
			t.Fatalf("supported pinned write %q missing", name)
		}
		got := action.Method + " " + action.Path
		if got != requestLine {
			t.Fatalf("write %q request line = %q, want %q", name, got, requestLine)
		}
	}

	for name, field := range map[string]string{
		"update_patient":                       "uuid",
		"update_appointment_status":            "appointmentUuid",
		"update_appointment_provider_response": "appointmentUuid",
	} {
		action, ok := bahmniWrite(b, name)
		if !ok {
			t.Fatalf("write %q missing", name)
		}
		if !containsStringSlice(action.RedactFields, field) {
			t.Fatalf("write %q redact_fields = %v, want %q", name, action.RedactFields, field)
		}
	}
}

func requireBahmniStream(t *testing.T, b engine.Bundle, name string) engine.StreamSpec {
	t.Helper()
	for _, stream := range b.Streams {
		if stream.Name == name {
			return stream
		}
	}
	t.Fatalf("stream %q missing", name)
	return engine.StreamSpec{}
}

func bahmniCommand(b engine.Bundle, path string) (engine.CLICommand, bool) {
	if b.CLISurface == nil {
		return engine.CLICommand{}, false
	}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Path == path {
			return cmd, true
		}
	}
	return engine.CLICommand{}, false
}

func bahmniOperation(b engine.Bundle, id string) (engine.OperationSpec, bool) {
	for _, op := range b.Operations {
		if op.ID == id {
			return op, true
		}
	}
	return engine.OperationSpec{}, false
}

func bahmniWrite(b engine.Bundle, name string) (engine.WriteAction, bool) {
	for _, action := range b.Writes {
		if action.Name == name {
			return action, true
		}
	}
	return engine.WriteAction{}, false
}

func hasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}

func containsStringSlice(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
