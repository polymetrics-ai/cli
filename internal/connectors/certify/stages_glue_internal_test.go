package certify

import "testing"

func TestFullSweepFlowAndScheduleNamesAreStreamScoped(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "sample"}}
	if got := rc.flowName(); got != "cert_flow_sample" {
		t.Fatalf("default flowName = %q", got)
	}
	if got := rc.flowTable(); got != "cert_flow_sample" {
		t.Fatalf("default flowTable = %q", got)
	}
	if got := rc.flowConnectionName(); got != "cert_flow_conn_sample" {
		t.Fatalf("default flowConnectionName = %q", got)
	}
	if got := rc.scheduleName(); got != "cert-schedule-sample" {
		t.Fatalf("default scheduleName = %q", got)
	}

	rc.currentStream = "pull requests"
	if got := rc.flowName(); got != "cert_flow_sample_pull_requests" {
		t.Fatalf("stream flowName = %q", got)
	}
	if got := rc.flowTable(); got != "cert_flow_sample_pull_requests" {
		t.Fatalf("stream flowTable = %q", got)
	}
	if got := rc.flowConnectionName(); got != "cert_flow_conn_sample_pull_requests" {
		t.Fatalf("stream flowConnectionName = %q", got)
	}
	if got := rc.scheduleName(); got != "cert-schedule-sample-pull-requests" {
		t.Fatalf("stream scheduleName = %q", got)
	}
}

func TestValidateInstalledScheduleLineRequiresQuotedRoot(t *testing.T) {
	const (
		sentinel = "# pm-schedule-cert-schedule-sample"
		cron     = "0 3 * * *"
		root     = "/tmp/cert root"
		flow     = "cert_flow_sample"
	)
	valid := "0 3 * * *  pm --root '/tmp/cert root' flow run cert_flow_sample --json  # pm-schedule-cert-schedule-sample\n"
	if got := validateInstalledScheduleLine(valid, sentinel, cron, root, flow); got != "" {
		t.Fatalf("validateInstalledScheduleLine(valid) = %q, want empty", got)
	}
	unquoted := "0 3 * * *  pm --root /tmp/cert root flow run cert_flow_sample --json  # pm-schedule-cert-schedule-sample\n"
	if got := validateInstalledScheduleLine(unquoted, sentinel, cron, root, flow); got == "" {
		t.Fatal("validateInstalledScheduleLine(unquoted) = empty, want root validation failure")
	}
}
