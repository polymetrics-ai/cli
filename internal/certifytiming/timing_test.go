package certifytiming

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

var certifyHarnessTarget = Target{
	Name:       "certify-harness",
	GoPackage:  "./internal/connectors/certify",
	ImportPath: "polymetrics.ai/internal/connectors/certify",
}

func TestParseTargetAcceptsPassingEventsAndSortsSlowTests(t *testing.T) {
	result, err := ParseTarget(certifyHarnessTarget, strings.NewReader(passEvents(certifyHarnessTarget.ImportPath)))
	if err != nil {
		t.Fatalf("ParseTarget() error = %v", err)
	}
	if result.Elapsed != 3*time.Second {
		t.Errorf("Elapsed = %s, want 3s", result.Elapsed)
	}
	if len(result.SlowTests) != 2 {
		t.Fatalf("SlowTests = %#v, want 2 tests", result.SlowTests)
	}
	if got := result.SlowTests[0]; got.Name != "TestSlow" || got.Elapsed != 2*time.Second {
		t.Errorf("SlowTests[0] = %#v, want TestSlow at 2s", got)
	}
}

func TestParseTargetRejectsFailedPackage(t *testing.T) {
	stream := `{"Action":"run","Package":"polymetrics.ai/internal/connectors/certify","Test":"TestBroken"}
{"Action":"fail","Package":"polymetrics.ai/internal/connectors/certify","Test":"TestBroken","Elapsed":1}
{"Action":"fail","Package":"polymetrics.ai/internal/connectors/certify","Elapsed":1}
`
	_, err := ParseTarget(certifyHarnessTarget, strings.NewReader(stream))
	if err == nil || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("ParseTarget() error = %v, want failed-package error", err)
	}
}

func TestParseTargetRejectsMalformedEvent(t *testing.T) {
	_, err := ParseTarget(certifyHarnessTarget, strings.NewReader("not json\n"))
	if err == nil || !strings.Contains(err.Error(), "parse go test event") {
		t.Fatalf("ParseTarget() error = %v, want malformed-event error", err)
	}
}

func TestParseTargetRejectsMissingPackageCompletion(t *testing.T) {
	stream := `{"Action":"run","Package":"polymetrics.ai/internal/connectors/certify","Test":"TestSlow"}
{"Action":"pass","Package":"polymetrics.ai/internal/connectors/certify","Test":"TestSlow","Elapsed":2}
`
	_, err := ParseTarget(certifyHarnessTarget, strings.NewReader(stream))
	if err == nil || !strings.Contains(err.Error(), "completion") {
		t.Fatalf("ParseTarget() error = %v, want missing-completion error", err)
	}
}

func TestRunRetainsRawEventsAndPrintsSummary(t *testing.T) {
	var out bytes.Buffer
	_, err := Run(context.Background(), []Target{certifyHarnessTarget}, func(_ context.Context, _ Target, raw io.Writer) error {
		_, writeErr := io.WriteString(raw, passEvents(certifyHarnessTarget.ImportPath))
		return writeErr
	}, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"Action":"pass"`) {
		t.Errorf("Run() output omitted raw go test event: %s", got)
	}
	if !strings.Contains(got, "certify timing summary") || !strings.Contains(got, "TestSlow") || !strings.Contains(got, "wall_elapsed=") {
		t.Errorf("Run() output omitted timing summary or slow test: %s", got)
	}
}

func TestGoTestArgsAreColdAndTargeted(t *testing.T) {
	cliTarget := Target{
		Name:       "certify-cli",
		GoPackage:  "./internal/cli",
		ImportPath: "polymetrics.ai/internal/cli",
		Run:        "^TestCertifyCLI",
	}
	got := strings.Join(GoTestArgs(cliTarget), " ")
	for _, want := range []string{"test", "-count=1", "-json", "-run ^TestCertifyCLI", "./internal/cli"} {
		if !strings.Contains(got, want) {
			t.Errorf("GoTestArgs() = %q, want %q", got, want)
		}
	}
}

func TestCheckDurationBudgetRejectsControlledDuplicateFixture(t *testing.T) {
	report := Report{Targets: []TargetResult{
		{Name: "certify-harness", Elapsed: 1 * time.Second, WallElapsed: 4 * time.Second},
		{Name: "certify-cli-duplicate", Elapsed: 1 * time.Second, WallElapsed: 4 * time.Second},
	}}
	err := CheckDurationBudget(report, 7*time.Second)
	if err == nil {
		t.Fatal("CheckDurationBudget() error = nil, want duplicate fixture to exceed budget")
	}
	if !strings.Contains(err.Error(), "observed 8.000s") || !strings.Contains(err.Error(), "allowed 7.000s") {
		t.Errorf("duration budget error = %v, want observed and allowed values", err)
	}
}

func TestReportTotalWallElapsedSumsMeasuredTargets(t *testing.T) {
	report := Report{Targets: []TargetResult{
		{Name: "certify-harness", WallElapsed: 4 * time.Second},
		{Name: "certify-cli", WallElapsed: 5 * time.Second},
	}}
	if got := report.TotalWallElapsed(); got != 9*time.Second {
		t.Errorf("TotalWallElapsed() = %s, want 9s", got)
	}
}

func TestCheckDurationBudgetRejectsMissingWallMeasurement(t *testing.T) {
	report := Report{Targets: []TargetResult{{Name: "certify-harness", Elapsed: time.Second}}}
	err := CheckDurationBudget(report, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "no measured wall duration") {
		t.Fatalf("CheckDurationBudget() error = %v, want missing wall measurement", err)
	}
}

func passEvents(pkg string) string {
	return `{"Action":"run","Package":"` + pkg + `","Test":"TestFast"}
{"Action":"pass","Package":"` + pkg + `","Test":"TestFast","Elapsed":1}
{"Action":"run","Package":"` + pkg + `","Test":"TestSlow"}
{"Action":"pass","Package":"` + pkg + `","Test":"TestSlow","Elapsed":2}
{"Action":"pass","Package":"` + pkg + `","Elapsed":3}
`
}
