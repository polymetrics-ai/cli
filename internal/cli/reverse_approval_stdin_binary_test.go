package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestReverseETLApprovalUsesBoundedStdin proves that a reverse approval stays
// on the stdin channel throughout a real process invocation. The test emits
// only the checked process command line and sanitized evidence, never a token.
func TestReverseETLApprovalUsesBoundedStdin(t *testing.T) {
	t.Setenv("PM_REVERSE_APPROVAL_ENV_PROBE", "visible")
	binary := buildTransportPM(t)
	root := setupReverseApprovalStdinProject(t)

	t.Run("successful run keeps token out of every observable carrier", func(t *testing.T) {
		planID, token := planReverseApprovalStdinRun(t, root, "stdin-success")
		receipt := reverseApprovalReceiptPath(root, "stdin-success")

		command := exec.Command(binary,
			"reverse", "run", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		)
		stdin, err := command.StdinPipe()
		if err != nil {
			t.Fatalf("create approval stdin pipe: %v", err)
		}
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		if err := command.Start(); err != nil {
			t.Fatalf("start reverse run: %v", err)
		}

		commandLine := reverseApprovalProcessCommandLine(t, command.Process.Pid)
		assertReverseApprovalTokenAbsentFromArgv(t, commandLine, token)
		processEnvironment := reverseApprovalProcessEnvironment(t, command.Process.Pid)
		if !strings.Contains(processEnvironment, "PM_REVERSE_APPROVAL_ENV_PROBE=visible") {
			t.Fatal("live reverse process environment was not observable")
		}
		assertReverseApprovalTokenAbsentFromEnvironment(t, processEnvironment, token)
		t.Logf("reverse approval process argv: %s", commandLine)

		if _, err := io.WriteString(stdin, token+"\n"); err != nil {
			t.Fatalf("write approval stdin: %v", err)
		}
		if err := stdin.Close(); err != nil {
			t.Fatalf("close approval stdin: %v", err)
		}
		if err := command.Wait(); err != nil {
			t.Fatalf("approved reverse run failed without exposing its output: %v", err)
		}

		capturedLogs := stdout.String() + stderr.String()
		assertReverseApprovalTokenAbsentFromLogs(t, capturedLogs, token)
		assertReverseApprovalTokenAbsentFromProjectFiles(t, root, token)
		assertReverseApprovalTokenAbsentFromReceipt(t, receipt, token)
		assertReverseApprovalRunSucceeded(t, stdout.Bytes())

		beforeReplay, err := os.Stat(receipt)
		if err != nil {
			t.Fatalf("stat successful reverse receipt: %v", err)
		}
		replayOutput, replayErr := runReverseApprovalPM(binary, token+"\n",
			"reverse", "run", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		)
		if replayErr == nil {
			t.Fatal("replayed approval unexpectedly succeeded")
		}
		assertReverseApprovalTokenAbsentFromLogs(t, replayOutput, token)
		afterReplay, err := os.Stat(receipt)
		if err != nil {
			t.Fatalf("stat reverse receipt after replay: %v", err)
		}
		if afterReplay.Size() != beforeReplay.Size() {
			t.Fatal("replayed approval wrote another receipt")
		}
		emitReverseApprovalCarrierEvidence(t, commandLine, token)
	})

	for _, tc := range []struct {
		name  string
		args  []string
		stdin string
		want  string
	}{
		{name: "empty stdin", args: []string{"--approval-token-stdin"}, want: "approval token stdin must contain one bounded line"},
		{name: "oversized stdin", args: []string{"--approval-token-stdin"}, stdin: strings.Repeat("x", maxApprovalTokenStdinBytes+1) + "\n", want: "approval token stdin is too large"},
		{name: "malformed stdin", args: []string{"--approval-token-stdin"}, stdin: "line-one\nline-two\n", want: "approval token stdin must contain exactly one line"},
		{name: "valued marker", args: []string{"--approval-token-stdin=argv-value"}, want: "--approval-token-stdin must be a bare stdin marker"},
		{name: "retired argv carrier", args: []string{"--approve", "argv-value"}, want: "approval tokens must be supplied with --approval-token-stdin"},
	} {
		t.Run(tc.name+" rejects before a receipt", func(t *testing.T) {
			name := "stdin-reject-" + strings.ReplaceAll(tc.name, " ", "-")
			planID, token := planReverseApprovalStdinRun(t, root, name)
			receipt := reverseApprovalReceiptPath(root, name)
			args := append([]string{"reverse", "run", planID}, tc.args...)
			args = append(args, "--root", root, "--json")
			output, err := runReverseApprovalPM(binary, tc.stdin, args...)
			if err == nil {
				t.Fatal("invalid approval input unexpectedly succeeded")
			}
			if !strings.Contains(output, tc.want) {
				t.Fatalf("invalid approval input did not return the expected refusal %q", tc.want)
			}
			assertReverseApprovalTokenAbsentFromLogs(t, output, token)
			if _, err := os.Stat(receipt); !os.IsNotExist(err) {
				t.Fatalf("invalid approval input wrote a receipt: %v", err)
			}
		})
	}
}

func setupReverseApprovalStdinProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("PM_SAMPLE_TOKEN", "sample-token")
	for _, args := range [][]string{
		{"init", "--root", root, "--json"},
		{"credentials", "add", "sample-local", "--connector", "sample", "--from-env", "token=PM_SAMPLE_TOKEN", "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"credentials", "add", "outbox-local", "--connector", "outbox", "--config", "path=" + filepath.Join(root, ".polymetrics", "outbox"), "--root", root, "--json"},
		{"connections", "create", "sample-to-warehouse", "--source", "sample:sample-local", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--primary-key", "id", "--cursor", "updated_at", "--table", "sample_customers", "--root", root, "--json"},
		{"etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--root", root, "--json"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("set up reverse approval project: command failed with exit %d", code)
		}
	}
	return root
}

func planReverseApprovalStdinRun(t *testing.T, root, name string) (string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := []string{
		"reverse", "plan", name,
		"--source-table", "sample_customers",
		"--destination", "outbox:outbox-local",
		"--map", "id:external_id",
		"--map", "name:full_name",
		"--map", "email:email",
		"--root", root,
	}
	if code := Run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("plan reverse approval stdin run: command failed with exit %d", code)
	}
	planID := reverseApprovalField(t, stdout.String(), `Created reverse plan (\S+)`)
	token := reverseApprovalField(t, stdout.String(), `Approval token: (\S+)`)
	return planID, token
}

func reverseApprovalField(t *testing.T, text, pattern string) string {
	t.Helper()
	parts := regexpMustCompile(t, pattern).FindStringSubmatch(text)
	if len(parts) != 2 {
		t.Fatal("reverse plan output did not include the expected non-secret field")
	}
	return parts[1]
}

func regexpMustCompile(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	return regexp.MustCompile(pattern)
}

func reverseApprovalReceiptPath(root, name string) string {
	return filepath.Join(root, ".polymetrics", "outbox", name+".jsonl")
}

func reverseApprovalProcessCommandLine(t *testing.T, pid int) string {
	t.Helper()
	output, err := exec.Command("ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		t.Fatalf("inspect live reverse process argv: %v", err)
	}
	commandLine := strings.TrimSpace(string(output))
	if commandLine == "" {
		t.Fatal("live reverse process command line is empty")
	}
	return commandLine
}

func reverseApprovalProcessEnvironment(t *testing.T, pid int) string {
	t.Helper()
	output, err := exec.Command("ps", "eww", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		t.Fatalf("inspect live reverse process environment: %v", err)
	}
	return string(output)
}

func runReverseApprovalPM(binary, stdin string, args ...string) (string, error) {
	command := exec.Command(binary, args...)
	command.Stdin = strings.NewReader(stdin)
	output, err := command.CombinedOutput()
	return string(output), err
}

func assertReverseApprovalRunSucceeded(t *testing.T, output []byte) {
	t.Helper()
	var envelope struct {
		Kind string `json:"kind"`
		Run  struct {
			Status           string `json:"status"`
			RecordsSucceeded int    `json:"records_succeeded"`
		} `json:"run"`
	}
	if err := json.Unmarshal(output, &envelope); err != nil {
		t.Fatalf("decode successful reverse run output: %v", err)
	}
	if envelope.Kind != "ReverseRun" || envelope.Run.Status != "completed" || envelope.Run.RecordsSucceeded != 3 {
		t.Fatalf("unexpected successful reverse run envelope: kind=%q status=%q succeeded=%d", envelope.Kind, envelope.Run.Status, envelope.Run.RecordsSucceeded)
	}
}

func assertReverseApprovalTokenAbsentFromArgv(t *testing.T, commandLine, token string) {
	t.Helper()
	if strings.Contains(commandLine, token) {
		t.Fatal("approval token appeared in the live process argv")
	}
}

func assertReverseApprovalTokenAbsentFromEnvironment(t *testing.T, environment, token string) {
	t.Helper()
	if strings.Contains(environment, token) {
		t.Fatal("approval token appeared in the live process environment")
	}
}

func assertReverseApprovalTokenAbsentFromProjectFiles(t *testing.T, root, token string) {
	t.Helper()
	project := filepath.Join(root, ".polymetrics")
	err := filepath.WalkDir(project, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(contents, []byte(token)) {
			return fmt.Errorf("approval token found in project file")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("assert approval token absent from project files: %v", err)
	}
}

func assertReverseApprovalTokenAbsentFromLogs(t *testing.T, logs, token string) {
	t.Helper()
	if strings.Contains(logs, token) {
		t.Fatal("approval token appeared in captured logs")
	}
}

func assertReverseApprovalTokenAbsentFromReceipt(t *testing.T, receipt, token string) {
	t.Helper()
	contents, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatalf("read reverse receipt: %v", err)
	}
	if bytes.Contains(contents, []byte(token)) {
		t.Fatal("approval token appeared in the reverse receipt")
	}
}

func emitReverseApprovalCarrierEvidence(t *testing.T, commandLine, token string) {
	t.Helper()
	evidence := struct {
		Kind              string `json:"kind"`
		StdinCarrier      bool   `json:"stdin_carrier"`
		ArgvObserved      string `json:"argv_observed"`
		ReplayRejected    bool   `json:"replay_rejected"`
		ReceiptTokenFree  bool   `json:"receipt_token_free"`
		ProjectTokenFree  bool   `json:"project_token_free"`
		EnvironmentAbsent bool   `json:"environment_token_absent"`
	}{
		Kind:              "ReverseApprovalCarrierEvidence",
		StdinCarrier:      true,
		ArgvObserved:      commandLine,
		ReplayRejected:    true,
		ReceiptTokenFree:  true,
		ProjectTokenFree:  true,
		EnvironmentAbsent: true,
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("encode reverse approval evidence: %v", err)
	}
	if bytes.Contains(encoded, []byte(token)) {
		t.Fatal("approval token appeared in emitted evidence")
	}
	t.Logf("reverse_approval_evidence=%s", encoded)
}
