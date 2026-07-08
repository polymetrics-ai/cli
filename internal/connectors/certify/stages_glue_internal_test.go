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
	if got := rc.scheduleName(); got != "cert-schedule-sample-pull_requests" {
		t.Fatalf("stream scheduleName = %q", got)
	}
}
