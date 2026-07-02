package schedule

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Group C — golden rendering tests.

var nightlyManifest = Manifest{
	Name:      "nightly-leads",
	Cron:      "0 2 * * *",
	Flow:      "likely-customers",
	CreatedAt: time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC),
	UpdatedAt: time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC),
}

const testPmBin = "/usr/local/bin/pm"

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read golden %q: %v", name, err)
	}
	return string(data)
}

// C-1 — launchd plist golden.
func TestRenderPlist_Golden(t *testing.T) {
	got, err := renderPlist(nightlyManifest, testPmBin)
	if err != nil {
		t.Fatalf("renderPlist: %v", err)
	}
	want := readGolden(t, "launchd_nightly.golden")
	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Fatalf("plist mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// C-2 — systemd service golden.
func TestRenderService_Golden(t *testing.T) {
	got, err := renderService(nightlyManifest, testPmBin)
	if err != nil {
		t.Fatalf("renderService: %v", err)
	}
	want := readGolden(t, "systemd_nightly.service.golden")
	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Fatalf("service mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// C-2 — systemd timer golden.
func TestRenderTimer_Golden(t *testing.T) {
	got, err := renderTimer(nightlyManifest)
	if err != nil {
		t.Fatalf("renderTimer: %v", err)
	}
	want := readGolden(t, "systemd_nightly.timer.golden")
	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Fatalf("timer mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// C-2 — cronToOnCalendar table-driven.
func TestCronToOnCalendar(t *testing.T) {
	cases := []struct{ expr, want string }{
		{"0 2 * * *", "*-*-* 02:00:00"},
		{"*/15 * * * *", "*-*-* *:0/15:00"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.expr, func(t *testing.T) {
			got, err := cronToOnCalendar(tc.expr)
			if err != nil {
				t.Fatalf("cronToOnCalendar(%q): %v", tc.expr, err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// C-3 — crontab line golden.
func TestRenderCrontabLine_Golden(t *testing.T) {
	got, err := renderCrontabLine(nightlyManifest, testPmBin)
	if err != nil {
		t.Fatalf("renderCrontabLine: %v", err)
	}
	want := strings.TrimSpace(readGolden(t, "crontab_nightly.golden"))
	if strings.TrimSpace(got) != want {
		t.Fatalf("crontab line mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRenderScheduleCommandsIncludeRoot(t *testing.T) {
	manifest := nightlyManifest
	manifest.Root = "/tmp/polymetrics root"
	manifest.Flow = "nightly-flow"
	pmBin := "/opt/Poly Metrics/pm"

	crontab, err := renderCrontabLine(manifest, pmBin)
	if err != nil {
		t.Fatalf("renderCrontabLine: %v", err)
	}
	wantShellCommand := "'/opt/Poly Metrics/pm' --root '/tmp/polymetrics root' flow run nightly-flow --json"
	if !strings.Contains(crontab, wantShellCommand) {
		t.Fatalf("crontab command missing rooted flow run\ngot:  %q\nwant containing: %q", crontab, wantShellCommand)
	}

	service, err := renderService(manifest, pmBin)
	if err != nil {
		t.Fatalf("renderService: %v", err)
	}
	if !strings.Contains(service, "ExecStart="+wantShellCommand) {
		t.Fatalf("systemd command missing rooted flow run\ngot:\n%s\nwant ExecStart containing: %q", service, wantShellCommand)
	}

	plist, err := renderPlist(manifest, pmBin)
	if err != nil {
		t.Fatalf("renderPlist: %v", err)
	}
	for _, fragment := range []string{
		"<string>/opt/Poly Metrics/pm</string>",
		"<string>--root</string>",
		"<string>/tmp/polymetrics root</string>",
		"<string>nightly-flow</string>",
	} {
		if !strings.Contains(plist, fragment) {
			t.Fatalf("launchd plist missing %q\ngot:\n%s", fragment, plist)
		}
	}
}

// C-3 — removeCrontabLine removes sentinel.
func TestRemoveCrontabLine_RemovesSentinel(t *testing.T) {
	sentinel := "# pm-schedule-nightly-leads"
	content := "# existing line\n0 2 * * *  /usr/local/bin/pm flow run likely-customers --json  " + sentinel + "\n# another line\n"
	got, err := removeCrontabLine(content, "nightly-leads")
	if err != nil {
		t.Fatalf("removeCrontabLine: %v", err)
	}
	if strings.Contains(got, sentinel) {
		t.Fatalf("sentinel still present after removal: %q", got)
	}
	if !strings.Contains(got, "# existing line") || !strings.Contains(got, "# another line") {
		t.Fatalf("other lines were removed unexpectedly: %q", got)
	}
}

// C-3 — removeCrontabLine no-op when absent.
func TestRemoveCrontabLine_NoopWhenAbsent(t *testing.T) {
	content := "# some crontab\n*/5 * * * *  echo hello\n"
	got, err := removeCrontabLine(content, "nightly-leads")
	if err != nil {
		t.Fatalf("removeCrontabLine: %v", err)
	}
	if got != content {
		t.Fatalf("content changed when sentinel was absent\ngot:  %q\nwant: %q", got, content)
	}
}

// C-3 — removeCrontabLine idempotent.
func TestRemoveCrontabLine_Idempotent(t *testing.T) {
	sentinel := "# pm-schedule-nightly-leads"
	content := "0 2 * * *  /usr/local/bin/pm flow run likely-customers --json  " + sentinel + "\n"
	once, err := removeCrontabLine(content, "nightly-leads")
	if err != nil {
		t.Fatalf("first remove: %v", err)
	}
	twice, err := removeCrontabLine(once, "nightly-leads")
	if err != nil {
		t.Fatalf("second remove: %v", err)
	}
	if once != twice {
		t.Fatalf("remove is not idempotent\nonce:  %q\ntwice: %q", once, twice)
	}
}
