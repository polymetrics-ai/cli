package email

import (
	"bufio"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
)

// TestConnectorDeclaresExecutableProtocolCommandSurface is deliberately the
// first red checkpoint for the Email connector. A protocol connector must be
// reachable through the real command-runner preflight, not merely described
// by bundle metadata.
func TestConnectorDeclaresExecutableProtocolCommandSurface(t *testing.T) {
	c := New()
	for _, path := range [][]string{
		{"mailboxes", "list"},
		{"message", "send"},
	} {
		if err := commandrunner.Preflight(c, path); err != nil {
			t.Fatalf("Preflight(%q): %v", path, err)
		}
	}
}

func TestMessagesReadIsBlockedAndUndeclared(t *testing.T) {
	c := New()
	for _, stream := range c.Manifest().Streams {
		if stream.Name == "messages" {
			t.Fatal("Manifest() declares unavailable messages stream")
		}
	}
	for _, stream := range c.Definition().Streams {
		if stream.Name == "messages" {
			t.Fatal("Definition() declares unavailable messages stream")
		}
	}
	catalog, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	for _, stream := range catalog.Streams {
		if stream.Name == "messages" {
			t.Fatal("Catalog() declares unavailable messages stream")
		}
	}
	err = commandrunner.Preflight(c, []string{"messages", "list"})
	var blocked *commandrunner.BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Preflight(messages list) error = %T %v, want BlockedCommandError", err, err)
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "messages"}, func(connectors.Record) error { return nil }); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Read(messages) error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := any(c).(connectors.StatefulReader); ok {
		t.Fatal("Connector implements StatefulReader for unavailable messages")
	}
	if _, ok := any(&c).(connectors.StatefulReader); ok {
		t.Fatal("*Connector implements StatefulReader for unavailable messages")
	}
}

func TestMailboxesAreReachableThroughCommandRunner(t *testing.T) {
	fixture := startIMAPFixture(t)
	c := New()
	c.imapAddressOverride = fixture.address
	var records []connectors.Record
	result, err := commandrunner.Run(context.Background(), c, commandrunner.Request{
		Path:   []string{"mailboxes", "list"},
		Config: fixture.config,
		Limit:  10,
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("commandrunner.Run(mailboxes list): %v", err)
	}
	if result.Count != 2 || len(records) != 2 {
		t.Fatalf("mailboxes list emitted result=%+v records=%#v, want two local mailboxes", result, records)
	}
}

func TestMailboxListHonorsRequestedLimit(t *testing.T) {
	fixture := startIMAPFixture(t)
	c := New()
	c.imapAddressOverride = fixture.address
	records := readRecords(t, c, connectors.ReadRequest{
		Stream: mailboxesStream,
		Config: fixture.config,
		Limit:  1,
	})
	if len(records) != 1 {
		t.Fatalf("mailboxes Read emitted %d records with limit=1, want 1", len(records))
	}
}

func TestSendPreviewIsUnmaskedAndAttachmentDriftCannotDispatch(t *testing.T) {
	address, captured := startSMTPFixture(t)
	projectRoot := t.TempDir()
	runtimeRoot := filepath.Join(projectRoot, ".polymetrics")
	if err := os.MkdirAll(runtimeRoot, 0o700); err != nil {
		t.Fatalf("create attachment staging root: %v", err)
	}
	attachmentPath := filepath.Join(runtimeRoot, "note.txt")
	if err := osWriteFile(attachmentPath, []byte("attachment payload")); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	c := New()
	cfg := testRuntimeConfig(t)
	cfg.ProjectDir = runtimeRoot
	c.smtpAddressOverride = address
	request := connectors.WriteRequest{Action: sendAction, Config: cfg}
	records := []connectors.Record{{
		"to":                []string{"to@example.invalid"},
		"cc":                []string{"cc@example.invalid"},
		"bcc":               []string{"bcc@example.invalid"},
		"subject":           "visible subject",
		"body":              "visible body",
		"body_content_type": "text/plain",
		"attachments":       []string{"note.txt"},
	}}
	preview, err := c.DryRunWrite(context.Background(), request, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	previewText := strings.Join(preview.Warnings, "\n")
	for _, want := range []string{"visible subject", "visible body", "bcc@example.invalid", "YXR0YWNobWVudCBwYXlsb2Fk"} {
		if !strings.Contains(previewText, want) {
			t.Fatalf("unmasked preview missing %q: %s", want, previewText)
		}
	}
	if strings.Contains(strings.ToLower(previewText), "redacted") || strings.Contains(previewText, "***") {
		t.Fatalf("preview unexpectedly masks payload: %s", previewText)
	}
	if strings.Contains(previewText, "Bcc:") {
		t.Fatalf("MIME preview exposes a Bcc header instead of only the envelope: %s", previewText)
	}

	request.Approval = approvalForPreview(t, preview)
	if err := osWriteFile(attachmentPath, []byte("changed after preview")); err != nil {
		t.Fatalf("modify attachment: %v", err)
	}
	result, err := c.Write(context.Background(), request, records)
	if err == nil {
		t.Fatal("Write dispatched after attachment drift from the approved preview")
	}
	if result.RecordsWritten != 0 || captured.calls != 0 {
		t.Fatalf("attachment drift dispatched SMTP result=%+v calls=%d", result, captured.calls)
	}
}

func TestAttachmentsRequireRelativePathsWithinRuntimeStagingRoot(t *testing.T) {
	projectRoot := t.TempDir()
	runtimeRoot := filepath.Join(projectRoot, ".polymetrics")
	if err := os.MkdirAll(runtimeRoot, 0o700); err != nil {
		t.Fatalf("create attachment staging root: %v", err)
	}
	if err := osWriteFile(filepath.Join(runtimeRoot, "inside.txt"), []byte("staged attachment")); err != nil {
		t.Fatalf("write staged attachment: %v", err)
	}
	if err := osWriteFile(filepath.Join(projectRoot, "project-only.txt"), []byte("project attachment")); err != nil {
		t.Fatalf("write project-root attachment: %v", err)
	}
	outsideRoot := t.TempDir()
	outsidePath := filepath.Join(outsideRoot, "outside.txt")
	if err := osWriteFile(outsidePath, []byte("outside attachment")); err != nil {
		t.Fatalf("write outside attachment: %v", err)
	}
	if err := os.Symlink(outsidePath, filepath.Join(runtimeRoot, "escape.txt")); err != nil {
		t.Fatalf("create attachment symlink: %v", err)
	}

	c := New()
	cfg := testRuntimeConfig(t)
	cfg.ProjectDir = runtimeRoot
	request := connectors.WriteRequest{Action: sendAction, Config: cfg}
	for _, test := range []struct {
		name string
		path string
	}{
		{name: "project root", path: "project-only.txt"},
		{name: "absolute", path: filepath.Join(runtimeRoot, "inside.txt")},
		{name: "traversal", path: filepath.Join("..", "inside.txt")},
		{name: "symlink escape", path: "escape.txt"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := c.DryRunWrite(context.Background(), request, []connectors.Record{{
				"to":          []string{"to@example.invalid"},
				"subject":     "attachment boundary",
				"body":        "attachment boundary",
				"attachments": []string{test.path},
			}})
			if err == nil {
				t.Fatal("DryRunWrite accepted an attachment outside the runtime staging root")
			}
		})
	}
}

func TestMessageSendCommandBuildsTypedUnmaskedPreview(t *testing.T) {
	c := New()
	command, err := commandrunner.BuildWriteCommand(context.Background(), c, commandrunner.Request{
		Path:    []string{"message", "send"},
		Config:  testRuntimeConfig(t),
		Preview: true,
		Flags: map[string][]string{
			"to":      {"to@example.invalid"},
			"bcc":     {"bcc@example.invalid"},
			"subject": {"command subject"},
			"body":    {"command body"},
		},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand(message send): %v", err)
	}
	if command.Write != sendAction || command.ConfirmationChallenge != string(connectors.ConfirmationKindDestructive) {
		t.Fatalf("message send command = %+v, want destructive send_message", command)
	}
	if command.Preview == nil {
		t.Fatal("message send command did not build a preview")
	}
	previewText := strings.Join(command.Preview.Warnings, "\n")
	for _, want := range []string{"command subject", "command body", "bcc@example.invalid"} {
		if !strings.Contains(previewText, want) {
			t.Fatalf("command preview missing unmasked %q: %s", want, previewText)
		}
	}
}

func TestSendRequiresTypedApprovalBeforeSMTPDispatch(t *testing.T) {
	address, captured := startSMTPFixture(t)
	c := New()
	cfg := testRuntimeConfig(t)
	c.smtpAddressOverride = address
	request := connectors.WriteRequest{Action: sendAction, Config: cfg}
	records := []connectors.Record{{"to": []string{"to@example.invalid"}, "subject": "subject", "body": "body"}}
	result, err := c.Write(context.Background(), request, records)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "approval") {
		t.Fatalf("Write without approval error = %v, want approval rejection", err)
	}
	if result.RecordsWritten != 0 || captured.calls != 0 {
		t.Fatalf("Write without approval dispatched SMTP result=%+v calls=%d", result, captured.calls)
	}

	preview, err := c.DryRunWrite(context.Background(), request, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	request.Approval = approvalForPreview(t, preview)
	result, err = c.Write(context.Background(), request, records)
	if err != nil {
		t.Fatalf("approved Write: %v", err)
	}
	if result.RecordsWritten != 1 || captured.calls != 1 {
		t.Fatalf("approved Write result=%+v calls=%d, want one SMTP submission", result, captured.calls)
	}
	if !strings.Contains(captured.data, "Subject: subject\r\n") || !strings.Contains(captured.data, "body") {
		t.Fatalf("SMTP DATA = %q, want prepared MIME payload", captured.data)
	}
}

// Connector-command plans are persisted as JSON and later decoded back into a
// connectors.Record. JSON arrays become []any, so this regression test keeps
// a saved plan executable through its preview and approval lifecycle.
func TestSendMessageAcceptsPersistedJSONArrays(t *testing.T) {
	c := New()
	_, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{
		Action: sendAction,
		Config: testRuntimeConfig(t),
	}, []connectors.Record{{
		"to":          []any{"to@example.invalid"},
		"cc":          []any{"cc@example.invalid"},
		"bcc":         []any{"bcc@example.invalid"},
		"subject":     "saved command plan",
		"body":        "serialized recipient arrays remain typed",
		"attachments": []any{},
	}})
	if err != nil {
		t.Fatalf("DryRunWrite with JSON-decoded arrays: %v", err)
	}
}

func TestInvalidPortErrorsDoNotEchoValues(t *testing.T) {
	cfg := testRuntimeConfig(t)
	cfg.Config["imap_port"] = "999"
	_, err := resolveConnectionConfig(cfg)
	if err == nil {
		t.Fatal("resolveConnectionConfig accepted an invalid IMAP port")
	}
	if !strings.Contains(err.Error(), "imap_port") {
		t.Fatal("resolveConnectionConfig invalid-port error did not identify imap_port")
	}
	if strings.Contains(err.Error(), "999") || strings.Contains(err.Error(), cfg.Secrets["password"]) {
		t.Fatal("configuration error exposed a supplied value or secret")
	}
}

func TestResolveConnectionConfigRejectsRawCredentialControls(t *testing.T) {
	const marker = "email-control-probe"
	cases := []struct {
		name   string
		field  string
		mutate func(connectors.RuntimeConfig)
	}{
		{name: "IMAP host", field: "imap_host", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["imap_host"] = "\r\n" + marker }},
		{name: "IMAP port", field: "imap_port", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["imap_port"] = "\r\n" + marker }},
		{name: "IMAP security", field: "imap_security", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["imap_security"] = "\r\n" + marker }},
		{name: "SMTP host", field: "smtp_host", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["smtp_host"] = "\r\n" + marker }},
		{name: "SMTP port", field: "smtp_port", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["smtp_port"] = "\r\n" + marker }},
		{name: "SMTP security", field: "smtp_security", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["smtp_security"] = "\r\n" + marker }},
		{name: "username", field: "username", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["username"] = "\r\n" + marker }},
		{name: "SMTP username", field: "smtp_username", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["smtp_username"] = "\r\n" + marker }},
		{name: "from address", field: "from_address", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["from_address"] = "\r\n" + marker }},
		{name: "connection timeout", field: "connection_timeout_seconds", mutate: func(cfg connectors.RuntimeConfig) { cfg.Config["connection_timeout_seconds"] = "\r\n" + marker }},
		{name: "password", field: "password", mutate: func(cfg connectors.RuntimeConfig) {
			cfg.Secrets["password"] = "\r\n" + marker + cfg.Secrets["password"]
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testRuntimeConfig(t)
			secret := cfg.Secrets["password"]
			tc.mutate(cfg)
			_, err := resolveConnectionConfig(cfg)
			if err == nil {
				t.Fatalf("resolveConnectionConfig accepted a control character in %s", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) || !strings.Contains(strings.ToLower(err.Error()), "control") {
				t.Fatalf("resolveConnectionConfig(%s) error = %q, want field and control constraint", tc.field, err)
			}
			if strings.Contains(err.Error(), marker) || strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "\r") || strings.Contains(err.Error(), "\n") {
				t.Fatal("resolveConnectionConfig error exposed a supplied value or secret")
			}
		})
	}
}

func TestSendMessageRejectsRawControlsBeforeSMTPConstruction(t *testing.T) {
	const marker = "email-control-probe"
	cases := []struct {
		name   string
		field  string
		mutate func(connectors.RuntimeConfig, connectors.Record)
	}{
		{name: "primary recipient", field: "to", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) {
			record["to"] = []string{"to@example.invalid\r\n" + marker}
		}},
		{name: "carbon-copy recipient", field: "cc", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) {
			record["cc"] = []string{"cc@example.invalid\r\n" + marker}
		}},
		{name: "blind-copy recipient", field: "bcc", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) {
			record["bcc"] = []string{"bcc@example.invalid\r\n" + marker}
		}},
		{name: "subject", field: "subject", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) { record["subject"] = "subject\r\n" + marker }},
		{name: "body content type", field: "body_content_type", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) {
			record["body_content_type"] = "text/plain\r\n" + marker
		}},
		{name: "attachment", field: "attachments", mutate: func(_ connectors.RuntimeConfig, record connectors.Record) {
			record["attachments"] = []string{"attachment\r\n" + marker}
		}},
		{name: "from address", field: "from_address", mutate: func(cfg connectors.RuntimeConfig, _ connectors.Record) {
			cfg.Config["from_address"] = "from@example.invalid\r\n" + marker
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testRuntimeConfig(t)
			record := connectors.Record{
				"to":      []string{"to@example.invalid"},
				"subject": "subject",
				"body":    "body",
			}
			tc.mutate(cfg, record)
			_, err := New().DryRunWrite(context.Background(), connectors.WriteRequest{Action: sendAction, Config: cfg}, []connectors.Record{record})
			if err == nil {
				t.Fatalf("DryRunWrite accepted a control character in %s", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) || !strings.Contains(strings.ToLower(err.Error()), "control") {
				t.Fatalf("DryRunWrite(%s) error = %q, want field and control constraint", tc.field, err)
			}
			if strings.Contains(err.Error(), marker) || strings.Contains(err.Error(), "\r") || strings.Contains(err.Error(), "\n") {
				t.Fatal("DryRunWrite error exposed a supplied value")
			}
		})
	}
}

func readRecords(t *testing.T, connector Connector, request connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var records []connectors.Record
	if err := connector.Read(context.Background(), request, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", request.Stream, err)
	}
	return records
}

func testRuntimeConfig(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	return testRuntimeConfigWithSecret(t, transientTestSecret(t))
}

func testRuntimeConfigWithSecret(t *testing.T, secret string) connectors.RuntimeConfig {
	t.Helper()
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"imap_host":                  "127.0.0.1",
			"imap_port":                  "143",
			"imap_security":              "none",
			"smtp_host":                  "127.0.0.1",
			"smtp_port":                  "25",
			"smtp_security":              "none",
			"username":                   "reader@example.invalid",
			"from_address":               "reader@example.invalid",
			"connection_timeout_seconds": "5",
		},
		Secrets:             map[string]string{"password": secret},
		CredentialRevision:  "fixture-email-credential-revision",
		ConfigurationDigest: "fixture-email-configuration-digest",
		WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
	}
}

func transientTestSecret(t *testing.T) string {
	t.Helper()
	bytes := make([]byte, 32)
	if _, err := cryptorand.Read(bytes); err != nil {
		t.Fatalf("generate test secret: %v", err)
	}
	return hex.EncodeToString(bytes)
}

type imapFixture struct {
	address string
	config  connectors.RuntimeConfig
}

func startIMAPFixture(t *testing.T) imapFixture {
	t.Helper()
	mem := imapmemserver.New()
	secret := transientTestSecret(t)
	user := imapmemserver.NewUser("reader@example.invalid", secret)
	for _, mailbox := range []string{"INBOX", "Archive"} {
		if err := user.Create(mailbox, nil); err != nil {
			t.Fatalf("create local IMAP mailbox %s: %v", mailbox, err)
		}
	}
	mem.AddUser(user)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen IMAP fixture: %v", err)
	}
	server := imapserver.New(&imapserver.Options{
		NewSession: func(*imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return mem.NewSession(), nil, nil
		},
		Caps:         imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapIMAP4rev2: {}},
		InsecureAuth: true,
	})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	return imapFixture{
		address: listener.Addr().String(),
		config:  testRuntimeConfigWithSecret(t, secret),
	}
}

type smtpCapture struct {
	calls int
	data  string
}

func startSMTPFixture(t *testing.T) (string, *smtpCapture) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen SMTP fixture: %v", err)
	}
	capture := &smtpCapture{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := textproto.NewReader(bufio.NewReader(conn))
		writer := textproto.NewWriter(bufio.NewWriter(conn))
		if err := writer.PrintfLine("220 local fixture ready"); err != nil {
			return
		}
		if err := writer.W.Flush(); err != nil {
			return
		}
		for {
			line, err := reader.ReadLine()
			if err != nil {
				return
			}
			switch {
			case strings.HasPrefix(strings.ToUpper(line), "EHLO "):
				_, _ = fmt.Fprint(writer.W, "250-localhost\r\n250-AUTH PLAIN LOGIN\r\n250 OK\r\n")
				_ = writer.W.Flush()
			case strings.HasPrefix(strings.ToUpper(line), "AUTH PLAIN"):
				_ = writer.PrintfLine("235 authenticated")
				_ = writer.W.Flush()
			case strings.HasPrefix(strings.ToUpper(line), "MAIL FROM:"):
				_ = writer.PrintfLine("250 sender accepted")
				_ = writer.W.Flush()
			case strings.HasPrefix(strings.ToUpper(line), "RCPT TO:"):
				_ = writer.PrintfLine("250 recipient accepted")
				_ = writer.W.Flush()
			case strings.EqualFold(line, "DATA"):
				_ = writer.PrintfLine("354 send data")
				_ = writer.W.Flush()
				var data strings.Builder
				for {
					dataLine, err := reader.ReadLine()
					if err != nil {
						return
					}
					if dataLine == "." {
						break
					}
					data.WriteString(dataLine)
					data.WriteString("\r\n")
				}
				capture.calls++
				capture.data = data.String()
				_ = writer.PrintfLine("250 queued")
				_ = writer.W.Flush()
			case strings.EqualFold(line, "QUIT"):
				_ = writer.PrintfLine("221 bye")
				_ = writer.W.Flush()
				return
			default:
				_ = writer.PrintfLine("250 ok")
				_ = writer.W.Flush()
			}
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("SMTP fixture did not stop")
		}
	})
	return listener.Addr().String(), capture
}

func approvalForPreview(t *testing.T, preview connectors.WritePreview) *connectors.WriteApprovalEvidence {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority: %v", err)
	}
	token := "fixture-email-approval"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_fixture_email",
		PlanHash:      strings.Repeat("a", 64),
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant: %v", err)
	}
	evidence, err := authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: preview.Digest,
		ApprovalToken: token,
		Target:        preview.ApprovalTarget,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		t.Fatalf("VerifyWriteGrant: %v", err)
	}
	return evidence
}

func osWriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o600)
}
