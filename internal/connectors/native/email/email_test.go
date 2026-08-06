package email

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"strings"
	"testing"
	"time"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
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
		{"messages", "list"},
		{"message", "send"},
	} {
		if err := commandrunner.Preflight(c, path); err != nil {
			t.Fatalf("Preflight(%q): %v", path, err)
		}
	}
}

func TestMessagesReadUsesUIDValidityCursorAndBoundedBodyParts(t *testing.T) {
	address, appendMessage := startIMAPFixture(t)
	appendMessage(t, "From: sender@example.invalid\r\nTo: reader@example.invalid\r\nSubject: first\r\nContent-Type: text/plain\r\n\r\n"+strings.Repeat("x", maxBodyPartBytes+128))

	c := New()
	c.imapAddressOverride = address
	records := readRecords(t, c, connectors.ReadRequest{Stream: messagesStream, Config: testRuntimeConfig(), Limit: 1})
	if len(records) != 1 {
		t.Fatalf("first messages Read emitted %d records, want 1", len(records))
	}
	first := records[0]
	if first["uid"] != "1" || first["uid_validity"] != "1" {
		t.Fatalf("first cursor identity = uid=%v uid_validity=%v, want 1/1", first["uid"], first["uid_validity"])
	}
	cursor, ok := first["imap_cursor"].(string)
	if !ok || cursor == "" {
		t.Fatalf("imap_cursor = %#v, want encoded UIDVALIDITY+UID cursor", first["imap_cursor"])
	}
	parts, ok := first["body_parts"].([]connectors.Record)
	if !ok || len(parts) != 1 {
		t.Fatalf("body_parts = %#v, want one bounded leaf part", first["body_parts"])
	}
	content, err := base64.StdEncoding.DecodeString(parts[0]["content_base64"].(string))
	if err != nil {
		t.Fatalf("decode body part: %v", err)
	}
	if len(content) != maxBodyPartBytes || parts[0]["truncated"] != true {
		t.Fatalf("bounded body part = len %d truncated=%v, want %d/true", len(content), parts[0]["truncated"], maxBodyPartBytes)
	}

	appendMessage(t, "From: sender@example.invalid\r\nTo: reader@example.invalid\r\nSubject: second\r\nContent-Type: text/plain\r\n\r\nsecond body")
	records = readRecords(t, c, connectors.ReadRequest{
		Stream: messagesStream,
		Config: testRuntimeConfig(),
		State:  map[string]string{"cursor": cursor},
		Limit:  1,
	})
	if len(records) != 1 || records[0]["uid"] != "2" {
		t.Fatalf("incremental messages Read = %#v, want only UID 2", records)
	}
}

func TestMailboxesAreReachableThroughCommandRunner(t *testing.T) {
	address, _ := startIMAPFixture(t)
	c := New()
	c.imapAddressOverride = address
	var records []connectors.Record
	result, err := commandrunner.Run(context.Background(), c, commandrunner.Request{
		Path:   []string{"mailboxes", "list"},
		Config: testRuntimeConfig(),
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
	address, _ := startIMAPFixture(t)
	c := New()
	c.imapAddressOverride = address
	records := readRecords(t, c, connectors.ReadRequest{
		Stream: mailboxesStream,
		Config: testRuntimeConfig(),
		Limit:  1,
	})
	if len(records) != 1 {
		t.Fatalf("mailboxes Read emitted %d records with limit=1, want 1", len(records))
	}
}

func TestSendPreviewIsUnmaskedAndAttachmentDriftCannotDispatch(t *testing.T) {
	address, captured := startSMTPFixture(t)
	root := t.TempDir()
	attachmentPath := root + "/note.txt"
	if err := osWriteFile(attachmentPath, []byte("attachment payload")); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	c := New()
	cfg := testRuntimeConfig()
	cfg.ProjectDir = root
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

func TestMessageSendCommandBuildsTypedUnmaskedPreview(t *testing.T) {
	c := New()
	command, err := commandrunner.BuildWriteCommand(context.Background(), c, commandrunner.Request{
		Path:    []string{"message", "send"},
		Config:  testRuntimeConfig(),
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
	cfg := testRuntimeConfig()
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
		Config: testRuntimeConfig(),
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

func TestCursorRejectsDifferentMailboxAndPortErrorsDoNotEchoValues(t *testing.T) {
	encoded := encodeCursor("INBOX", 7, 9)
	if _, err := decodeCursor(encoded, "Archive"); err == nil || !strings.Contains(err.Error(), "different mailbox") {
		t.Fatalf("decodeCursor for another mailbox error = %v, want explicit mailbox-state rejection", err)
	}
	cfg := testRuntimeConfig()
	cfg.Config["imap_port"] = "999"
	_, err := resolveConnectionConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "imap_port") {
		t.Fatalf("resolveConnectionConfig invalid port error = %v, want imap_port constraint", err)
	}
	if strings.Contains(err.Error(), "999") || strings.Contains(err.Error(), cfg.Secrets["password"]) {
		t.Fatalf("configuration error exposed supplied value or secret: %q", err)
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

func testRuntimeConfig() connectors.RuntimeConfig {
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
		Secrets:             map[string]string{"password": "fixture-only-not-a-password"},
		CredentialRevision:  "fixture-email-credential-revision",
		ConfigurationDigest: "fixture-email-configuration-digest",
		WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
	}
}

func startIMAPFixture(t *testing.T) (string, func(*testing.T, string)) {
	t.Helper()
	mem := imapmemserver.New()
	user := imapmemserver.NewUser("reader@example.invalid", "fixture-only-not-a-password")
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
	appendMessage := func(t *testing.T, message string) {
		t.Helper()
		client, err := imapclient.DialInsecure(listener.Addr().String(), nil)
		if err != nil {
			t.Fatalf("dial local IMAP fixture: %v", err)
		}
		defer func() { _ = client.Close() }()
		if err := client.Login("reader@example.invalid", "fixture-only-not-a-password").Wait(); err != nil {
			t.Fatalf("login local IMAP fixture: %v", err)
		}
		appendCommand := client.Append("INBOX", int64(len(message)), nil)
		if _, err := appendCommand.Write([]byte(message)); err != nil {
			t.Fatalf("append local IMAP fixture: %v", err)
		}
		if err := appendCommand.Close(); err != nil {
			t.Fatalf("close local IMAP append: %v", err)
		}
		if _, err := appendCommand.Wait(); err != nil {
			t.Fatalf("wait local IMAP append: %v", err)
		}
	}
	return listener.Addr().String(), appendMessage
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
