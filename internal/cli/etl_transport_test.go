package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
)

func TestETLTransportBareAndLeafHelpAreContextual(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"--help"},
		{"github-issue-label"},
		{"github-issue-label", "--help"},
		{"github-issue-label", "cleanup"},
		{"github-issue-label", "cleanup", "--help"},
	} {
		var stdout bytes.Buffer
		if err := runETLTransport(context.Background(), nil, args, &stdout, false); err != nil {
			t.Fatalf("runETLTransport(%v) = %v", args, err)
		}
		for _, want := range []string{"pm etl transport github-issue-label plan", "--approval-token-stdin", "cleanup"} {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("runETLTransport(%v) help missing %q", args, want)
			}
		}
	}
	for _, args := range [][]string{
		{"transport"},
		{"transport", "github-issue-label"},
		{"transport", "github-issue-label", "cleanup"},
		{"transport", "postgres-managed-target"},
		{"transport", "declarative-typed-destination"},
	} {
		if command, ok := etlTransportManualCommand(args); !ok || command == "" {
			t.Fatalf("etlTransportManualCommand(%v) = %q, %t; want a project-free transport manual", args, command, ok)
		}
	}
	var declaredStdout bytes.Buffer
	if err := runETLTransport(context.Background(), nil, []string{"declarative-typed-destination"}, &declaredStdout, false); err != nil {
		t.Fatalf("runDeclarativeTypedDestinationTransport help = %v", err)
	}
	for _, want := range []string{"destination_action", "writes.json", "cannot pass a connector, action, URL, method, body, mapping, or evidence", "--approval-token-stdin"} {
		if !strings.Contains(declaredStdout.String(), want) {
			t.Fatalf("declarative typed destination help missing %q", want)
		}
	}
	var postgresStdout bytes.Buffer
	if err := runETLTransport(context.Background(), nil, []string{"postgres-managed-target"}, &postgresStdout, false); err != nil {
		t.Fatalf("runPostgresManagedTargetTransport help = %v", err)
	}
	for _, want := range []string{"warehouse worksets", "incremental_upsert", "write=false", "--approval-token-stdin", "--authorization-lifetime", "24h through 48h"} {
		if !strings.Contains(postgresStdout.String(), want) {
			t.Fatalf("PostgreSQL transport help missing %q", want)
		}
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"etl", "transport", "github-issue-label", "cleanup", "--root", filepath.Join(t.TempDir(), "uninitialized")}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "pm etl transport github-issue-label plan") || stderr.Len() != 0 {
		t.Fatalf("bare cleanup namespace = code %d stdout=%q stderr=%q, want contextual manual without opening a project", code, stdout.String(), stderr.String())
	}
}

func TestDeclarativeTypedDestinationTransportRejectsCallerActionBeforeProjectIO(t *testing.T) {
	var stdout bytes.Buffer
	err := runETLTransport(context.Background(), nil, []string{
		"declarative-typed-destination", "plan", "--connection", "typed_destination", "--stream", "widgets", "--action", "replace_widget",
	}, &stdout, false)
	if err == nil {
		t.Fatal("declarative typed destination transport accepted caller action selection")
	}
	if !strings.Contains(err.Error(), "unknown transport flag --action") {
		t.Fatalf("caller action refusal = %v, want unknown closed-selector flag", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("caller action refusal wrote output before project I/O: %q", stdout.String())
	}
}

func TestPostgresManagedTargetPlanRejectsMalformedAuthorizationLifetimeBeforeProjectIO(t *testing.T) {
	var stdout bytes.Buffer
	err := runETLTransport(context.Background(), nil, []string{
		"postgres-managed-target", "plan", "--connection", "transport-demo", "--stream", "commits", "--authorization-lifetime", "tomorrow",
	}, &stdout, false)
	var refusal *cliError
	if !errors.As(err, &refusal) || refusal.category != categoryValidation {
		t.Fatalf("malformed authorization lifetime error = %T %v, want typed validation refusal", err, err)
	}
	if !strings.Contains(refusal.Error(), "invalid --authorization-lifetime") {
		t.Fatalf("malformed authorization lifetime refusal = %q, want exact flag reason", refusal.Error())
	}
}

func TestETLRunTransportApprovalRejectsUnsafeOrIncompleteCarriage(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		stdin string
		code  int
	}{
		{
			name: "raw approval token flag",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-token", "not-allowed"},
			code: 2,
		},
		{
			name: "caller selected provider method",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--method", "POST"},
			code: 2,
		},
		{
			name: "runtime is not part of closed carrier",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-plan", "rplan_one", "--approval-token-stdin", "--confirm", "destructive", "--runtime"},
			code: 2,
		},
		{
			name: "repeated plan selector",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-plan", "rplan_one", "--approval-plan", "rplan_two", "--approval-token-stdin", "--confirm", "destructive"},
			code: 2,
		},
		{
			name: "value supplied to stdin marker",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-plan", "rplan_one", "--approval-token-stdin", "token-on-argv", "--confirm", "destructive"},
			code: 2,
		},
		{
			name: "partial tuple",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-plan", "rplan_one"},
			code: 3,
		},
		{
			name: "wrong stream",
			args: []string{"--connection", "transport-demo", "--stream", "pull_requests", "--batch-size", "1", "--approval-plan", "rplan_one", "--approval-token-stdin", "--confirm", "destructive"},
			code: 3,
		},
		{
			name: "wrong batch size",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "2", "--approval-plan", "rplan_one", "--approval-token-stdin", "--confirm", "destructive"},
			code: 3,
		},
		{
			name: "invalid confirmation",
			args: []string{"--connection", "transport-demo", "--stream", "issues", "--batch-size", "1", "--approval-plan", "rplan_one", "--approval-token-stdin", "--confirm", "yes"},
			code: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, present, _, err := parseETLRunTransportApproval(tt.args, strings.NewReader(tt.stdin))
			if !present || err == nil {
				t.Fatalf("parseETLRunTransportApproval(%v) = present=%t err=%v, want rejected transport carriage", tt.args, present, err)
			}
			if got := exitCodeFor(classifyError(err)); got != tt.code {
				t.Fatalf("transport carriage error exit code = %d, want %d", got, tt.code)
			}
			if strings.Contains(err.Error(), "token-on-argv") {
				t.Fatal("transport carriage error echoed a raw token-like value")
			}
		})
	}
}

func TestETLRunTransportApprovalReadsExactlyOneEphemeralStdinLine(t *testing.T) {
	validArgs := []string{
		"--connection", "transport-demo",
		"--stream", "issues",
		"--batch-size", "1",
		"--approval-plan", "rplan_one",
		"--approval-token-stdin",
		"--confirm", "destructive",
	}
	approval, present, _, err := parseETLRunTransportApproval(validArgs, strings.NewReader("ephemeral-token\r\n"))
	if err != nil || !present || approval.PlanID != "rplan_one" || approval.ApprovalToken != "ephemeral-token" || approval.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("valid stdin approval = %#v present=%t err=%v", approval, present, err)
	}

	for _, input := range []string{
		strings.Repeat("x", maxApprovalTokenStdinBytes) + "\n",
		strings.Repeat("x", maxApprovalTokenStdinBytes) + "\r\n",
	} {
		approval, present, _, err := parseETLRunTransportApproval(validArgs, strings.NewReader(input))
		if err != nil || !present || len(approval.ApprovalToken) != maxApprovalTokenStdinBytes {
			t.Fatalf("full bounded approval stdin = present=%t token_len=%d err=%v, want accepted", present, len(approval.ApprovalToken), err)
		}
	}

	for _, input := range []string{
		"",
		"one\ntwo\n",
		"one\n\n",
		"one\n ",
		strings.Repeat("x", maxApprovalTokenStdinBytes+1) + "\n",
	} {
		_, _, _, err := parseETLRunTransportApproval(validArgs, strings.NewReader(input))
		if err == nil {
			t.Fatal("invalid approval stdin was accepted")
		}
		if input != "" && strings.Contains(err.Error(), input) {
			t.Fatal("approval stdin error echoed input bytes")
		}
	}
}

func TestETLRunTransportApprovalAllowsDurablePlanReferenceWithoutTokenCarrier(t *testing.T) {
	args := []string{
		"--connection", "transport-demo",
		"--stream", "issues",
		"--batch-size", "1",
		"--approval-plan", "rplan_one",
		"--confirm", "destructive",
	}
	approval, present, flags, err := parseETLRunTransportApproval(args, strings.NewReader(""))
	if err != nil {
		t.Fatalf("parseETLRunTransportApproval() = %v, want durable authorization carriage", err)
	}
	if !present || approval.PlanID != "rplan_one" || approval.ApprovalToken != "" || approval.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("parseETLRunTransportApproval() = present=%t approval=%+v, want plan-scoped durable authorization", present, approval)
	}
	if flags.value("approval-token-stdin") != "" {
		t.Fatal("durable authorization carriage unexpectedly requested token stdin")
	}
}

func TestETLRunTransportApprovalLeavesStreamAndBatchPolicyToTheResolvedRoute(t *testing.T) {
	approval, present, _, err := parseETLRunTransportApproval([]string{
		"--connection", "postgres-transport",
		"--stream", "issues",
		"--batch-size", "1000",
		"--approval-plan", "rplan_postgres",
		"--approval-token-stdin",
		"--confirm", "destructive",
	}, strings.NewReader("ephemeral-token\n"))
	if err != nil || !present || approval.ApprovalToken != "ephemeral-token" {
		t.Fatalf("generic transport carriage = %#v present=%t err=%v, want route-owned stream and batch policy", approval, present, err)
	}
}

func TestETLRunTransportApprovalLeavesLegacyRuntimeAlone(t *testing.T) {
	approval, present, _, err := parseETLRunTransportApproval([]string{
		"--connection", "ordinary-connection",
		"--stream", "issues",
		"--batch-size", "10",
		"--runtime",
	}, strings.NewReader(""))
	if err != nil || present || approval.PlanID != "" || approval.ApprovalToken != "" || approval.Evidence != nil || approval.AuthorizeNextUnit != nil {
		t.Fatalf("legacy --runtime carrier = %#v present=%t err=%v, want untouched legacy ETL arguments", approval, present, err)
	}
}

func TestETLRunRejectsRuntimeWriteShapingFlagsBeforeLegacyDispatch(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value string
	}{
		{name: "destination-action", value: "arbitrary"},
		{name: "destination_action", value: "arbitrary"},
		{name: "connector", value: "foreign"},
		{name: "action", value: "replace"},
		{name: "route", value: "/foreign"},
		{name: "verb", value: "PATCH"},
		{name: "method", value: "POST"},
		{name: "path", value: "/foreign"},
		{name: "url", value: "https://foreign.example.test"},
		{name: "body", value: `{"override":true}`},
		{name: "payload", value: "override"},
		{name: "query", value: "override=true"},
		{name: "headers", value: "X-Override=true"},
		{name: "mapping", value: "source:target"},
		{name: "map", value: "source:target"},
		{name: "input-fields", value: "source=target"},
		{name: "evidence", value: "caller-evidence"},
		{name: "destination-config", value: "target=override"},
		{name: "destination", value: "foreign:credential"},
		{name: "credential", value: "foreign"},
		{name: "source-config", value: "source=override"},
		{name: "sql", value: "UPDATE records"},
		{name: "shell", value: "touch ignored"},
		{name: "http", value: "request"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := runETL(context.Background(), nil, []string{
				"run",
				"--connection", "ordinary-connection",
				"--stream", "widgets",
				"--batch-size", "1",
				"--" + tt.name, tt.value,
			}, io.Discard, false, config.Config{})
			if err == nil {
				t.Fatalf("runtime --%s selection was accepted", tt.name)
			}
			var refusal *cliError
			if !errors.As(err, &refusal) || refusal.category != categoryUsage {
				t.Fatalf("runtime --%s refusal = %T %v, want usage error", tt.name, err, err)
			}
			if strings.Contains(err.Error(), tt.value) {
				t.Fatalf("runtime --%s refusal leaked caller value %q", tt.name, tt.value)
			}
		})
	}
}

func transportPlanForOutputTest() app.ReversePlan {
	return app.ReversePlan{
		ID:                     "rplan_transport_output",
		Status:                 "planned",
		Mode:                   "issue_label_transport",
		DestinationCredential:  "private-destination-credential",
		DestinationConfig:      map[string]string{"private": "destination_config"},
		Action:                 "add_issue_labels",
		RecordCount:            1,
		PlanHash:               "plan-hash",
		TransportConnectionID:  "conn_transport_output",
		TransportBindingSHA256: "transport_binding",
		CreatedAt:              time.Unix(1, 0).UTC(),
		ExpiresAt:              time.Unix(2, 0).UTC(),
		ApprovalToken:          "raw-approval-token",
		ApprovalTokenHash:      "approval_token_hash",
		ConfirmationPolicy:     connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func TestETLTransportSafeOutputOmitsApprovalAndDestinationInternals(t *testing.T) {
	plan := transportPlanForOutputTest()
	var stdout bytes.Buffer
	if err := writeETLTransportPlan(&stdout, true, plan); err != nil {
		t.Fatal(err)
	}
	if err := writeETLTransportPreview(&stdout, true, plan, connectors.WritePreview{
		RecordsStaged: 1,
		Action:        plan.Action,
		Digest:        "safe-preview-digest",
		Warnings:      []string{"resolved request: POST https://private.example.test/issues/123/labels"},
		ApprovalTarget: connectors.WriteApprovalTarget{
			Connector: "github", Operation: plan.Action, Method: "POST", Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		},
	}); err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, forbidden := range []string{"raw-approval-token", "approval_token_hash", "approval_grant", "plan_seal", "private-destination-credential", "destination_config", "transport_binding", "private.example.test", "issues/123/labels", "approval_target"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("transport output leaked %q", forbidden)
		}
	}
	for _, want := range []string{"ETLTransportPlan", "ETLTransportPreview", "approval_token_issued", "safe-preview-digest"} {
		if !strings.Contains(output, want) {
			t.Fatalf("transport output missing %q", want)
		}
	}
}

func TestETLTransportPlanAndPreviewShowBoundedTargetCopyCapacity(t *testing.T) {
	plan := transportPlanForOutputTest()
	plan.TargetCopyWorkers = 2
	plan.TargetCopyWorkerMaximum = 8
	var stdout bytes.Buffer
	if err := writeETLTransportPlan(&stdout, false, plan); err != nil {
		t.Fatal(err)
	}
	if err := writeETLTransportPreview(&stdout, false, plan, connectors.WritePreview{RecordsStaged: 1, Action: plan.Action, Digest: "safe-preview-digest"}); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Count(stdout.String(), "Target COPY workers: 2 (declared pool maximum 8)"), 2; got != want {
		t.Fatalf("target COPY capacity lines = %d, want %d in plan and preview:\n%s", got, want, stdout.String())
	}

	stdout.Reset()
	if err := writeETLTransportPlan(&stdout, true, plan); err != nil {
		t.Fatal(err)
	}
	if err := writeETLTransportPreview(&stdout, true, plan, connectors.WritePreview{RecordsStaged: 1, Action: plan.Action, Digest: "safe-preview-digest"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"target_copy_workers": 2`, `"target_copy_worker_maximum": 8`} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("machine-readable target COPY capacity output missing %q:\n%s", want, stdout.String())
		}
	}
}
