package ashby

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestHiringTeamWritesUseTypedRecords(t *testing.T) {
	validator := New().(connectors.WriteValidator)
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.ashbyhq.com"}}
	validRecord := connectors.Record{
		"applicationId": "application_fixture",
		"teamMemberId":  "user_fixture",
		"roleId":        "role_fixture",
	}
	for _, action := range []string{"add_hiring_team_member", "remove_hiring_team_member"} {
		action := action
		t.Run(action+" rejects arbitrary record", func(t *testing.T) {
			req := connectors.WriteRequest{Action: action, Config: cfg}
			if err := validator.ValidateWrite(context.Background(), req, []connectors.Record{{"unexpected": "field"}}); err == nil {
				t.Fatalf("ValidateWrite(%s) accepted an arbitrary record", action)
			}
		})
		t.Run(action+" accepts documented fields", func(t *testing.T) {
			req := connectors.WriteRequest{Action: action, Config: cfg}
			if err := validator.ValidateWrite(context.Background(), req, []connectors.Record{validRecord}); err != nil {
				t.Fatalf("ValidateWrite(%s) with documented fields: %v", action, err)
			}
		})
	}
}
