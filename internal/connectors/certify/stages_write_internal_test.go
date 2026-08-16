package certify

import "testing"

func TestPostgresProfileSkipsDirectWriteProtocolWithoutPairing(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "postgres", Write: true}}
	report := Report{}
	if err := stageWritePlanPreview(rc, &report); err != nil {
		t.Fatalf("stageWritePlanPreview() error = %v", err)
	}
	if rc.write != nil {
		t.Fatalf("PostgreSQL direct write protocol = %#v, want no unrelated outbox self-test", rc.write)
	}
	if len(report.Stages) != 1 || report.Stages[0].Name != "write_plan_preview" || report.Stages[0].Status != "skipped" {
		t.Fatalf("PostgreSQL direct write stage = %#v, want explicit skip", report.Stages)
	}
}
