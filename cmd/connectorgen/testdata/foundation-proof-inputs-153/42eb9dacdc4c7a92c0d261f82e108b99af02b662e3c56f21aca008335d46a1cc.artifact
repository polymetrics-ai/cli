package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

// A preview is the operator's description of the approval the write is about
// to demand, and `--preview --json` publishes its approval target verbatim.
// PreviewPreparedWrite used to stamp every prepared write `destructive`,
// including the safe ones GateDestructiveExecution waves straight through, so
// an approval-only create previewed a typed confirmation that RunReverseETL
// never asks for and `--confirm destructive` would be rejected as invalid.
//
// The declaration is now keyed on DestructiveTarget.RequiresApproval — the
// same predicate the gate itself uses — so the preview and the gate cannot
// describe different contracts.
func TestPreparedWritePreviewDeclaresConfirmationOnlyWhereTheGateDemandsIt(t *testing.T) {
	cases := []struct {
		name   string
		target DestructiveTarget
		want   connectors.ConfirmationKind
	}{
		{
			name:   "approval-only create",
			target: DestructiveTarget{Connector: "acme", Operation: "create_widget", Method: "POST", MutationClass: "create"},
		},
		{
			name:   "update",
			target: DestructiveTarget{Connector: "acme", Operation: "update_widget", Method: "PATCH", MutationClass: "update"},
		},
		{
			name:   "delete",
			target: DestructiveTarget{Connector: "acme", Operation: "delete_widget", Method: "DELETE", MutationClass: "delete"},
			want:   connectors.ConfirmationKindDestructive,
		},
		{
			name:   "declared destructive confirmation on a non-delete method",
			target: DestructiveTarget{Connector: "acme", Operation: "archive_widget", Method: "PATCH", MutationClass: "update", Confirmation: connectors.ConfirmationKindDestructive},
			want:   connectors.ConfirmationKindDestructive,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prepared := PreparedWrite{
				Target:              tc.target,
				CredentialRevision:  "cred-rev-1",
				ConfigurationDigest: "config-digest-1",
				RecordsStaged:       1,
				Action:              tc.target.Operation,
				Requests: []PreparedRequest{{
					Method: tc.target.Method,
					URL:    "https://api.example.com/widgets/1",
				}},
			}
			preview, err := PreviewPreparedWrite(prepared)
			if err != nil {
				t.Fatalf("PreviewPreparedWrite: %v", err)
			}
			if got := preview.ApprovalTarget.Confirmation.Kind; got != tc.want {
				t.Fatalf("preview approval target confirmation = %q, want %q; the gate %s demand one",
					got, tc.want, map[bool]string{true: "does", false: "does not"}[tc.target.RequiresApproval()])
			}
			if tc.target.RequiresApproval() != (preview.ApprovalTarget.Confirmation.Kind != "") {
				t.Fatalf("preview confirmation (%q) and gate requirement (%t) disagree",
					preview.ApprovalTarget.Confirmation.Kind, tc.target.RequiresApproval())
			}
		})
	}
}
