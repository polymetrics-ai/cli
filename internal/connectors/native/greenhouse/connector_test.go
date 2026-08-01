package greenhouse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestValidateWriteRejectsHiringTeamEmptyLists(t *testing.T) {
	t.Parallel()

	connector := New()
	validator := connector.(connectors.WriteValidator)
	dryRunner := connector.(connectors.DryRunWriter)

	for _, action := range []string{"replace_hiring_team", "add_hiring_team_members", "remove_hiring_team_member"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()
			req := connectors.WriteRequest{Action: action}
			empty := []connectors.Record{{"job_id": "job_id_fixture", "hiring_managers": []any{}}}
			if err := validator.ValidateWrite(context.Background(), req, empty); err == nil {
				t.Fatalf("ValidateWrite accepted empty hiring_managers")
			}
			if _, err := dryRunner.DryRunWrite(context.Background(), req, empty); err == nil {
				t.Fatalf("DryRunWrite accepted empty hiring_managers")
			}

			mixed := []connectors.Record{{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}, "recruiters": []any{}}}
			if err := validator.ValidateWrite(context.Background(), req, mixed); err == nil {
				t.Fatalf("ValidateWrite accepted an empty supplied recruiter list")
			}

			valid := []connectors.Record{{"job_id": "job_id_fixture", "hiring_managers": []any{float64(1234)}}}
			if err := validator.ValidateWrite(context.Background(), req, valid); err != nil {
				t.Fatalf("ValidateWrite rejected valid hiring-team record: %v", err)
			}
			if _, err := dryRunner.DryRunWrite(context.Background(), req, valid); err != nil {
				t.Fatalf("DryRunWrite rejected valid hiring-team record: %v", err)
			}
		})
	}
}

func TestValidateWriteRejectsAnonymizeEmptyFields(t *testing.T) {
	t.Parallel()

	connector := New()
	validator := connector.(connectors.WriteValidator)
	dryRunner := connector.(connectors.DryRunWriter)
	req := connectors.WriteRequest{Action: "anonymize_candidate"}

	empty := []connectors.Record{{"candidate_id": "candidate_id_fixture", "field_names": []any{}}}
	if err := validator.ValidateWrite(context.Background(), req, empty); err == nil {
		t.Fatalf("ValidateWrite accepted empty field_names")
	}
	if _, err := dryRunner.DryRunWrite(context.Background(), req, empty); err == nil {
		t.Fatalf("DryRunWrite accepted empty field_names")
	}

	unsupported := []connectors.Record{{"candidate_id": "candidate_id_fixture", "field_names": []any{"not_a_documented_field"}}}
	if err := validator.ValidateWrite(context.Background(), req, unsupported); err == nil {
		t.Fatalf("ValidateWrite accepted unsupported field_names")
	}

	valid := []connectors.Record{{"candidate_id": "candidate_id_fixture", "field_names": []any{"full_name", "email_addresses"}}}
	if err := validator.ValidateWrite(context.Background(), req, valid); err != nil {
		t.Fatalf("ValidateWrite rejected valid field_names: %v", err)
	}
	preview, err := dryRunner.DryRunWrite(context.Background(), req, valid)
	if err != nil {
		t.Fatalf("DryRunWrite rejected valid field_names: %v", err)
	}
	if len(preview.Warnings) < 2 || !strings.Contains(preview.Warnings[1], "/v1/candidates/candidate_id_fixture/anonymize?fields=full_name,email_addresses") {
		t.Fatalf("preview warnings = %#v, want joined anonymize fields", preview.Warnings)
	}
}

func TestWriteAnonymizeCandidateSendsJoinedFields(t *testing.T) {
	t.Parallel()

	captured := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured <- r.Method + " " + r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	connector := New()
	result, err := connector.Write(context.Background(), connectors.WriteRequest{
		Action: "anonymize_candidate",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": srv.URL},
			Secrets: map[string]string{"api_key": "fixture_api_key"},
		},
	}, []connectors.Record{{"candidate_id": "candidate_id_fixture", "field_names": []any{"full_name", "email_addresses"}}})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("Write result = %+v, want RecordsWritten=1 RecordsFailed=0", result)
	}
	got := <-captured
	want := "PUT /v1/candidates/candidate_id_fixture/anonymize?fields=full_name,email_addresses"
	if got != want {
		t.Fatalf("request = %q, want %q", got, want)
	}
}

func TestValidateWriteRejectsDestroyOpeningsEmptyIDs(t *testing.T) {
	t.Parallel()

	connector := New()
	validator := connector.(connectors.WriteValidator)
	dryRunner := connector.(connectors.DryRunWriter)
	req := connectors.WriteRequest{Action: "destroy_openings"}
	records := []connectors.Record{{"job_id": "job_id_fixture", "ids": []any{}}}
	if err := validator.ValidateWrite(context.Background(), req, records); err == nil {
		t.Fatalf("ValidateWrite accepted empty ids")
	}
	if _, err := dryRunner.DryRunWrite(context.Background(), req, records); err == nil {
		t.Fatalf("DryRunWrite accepted empty ids")
	}
}
