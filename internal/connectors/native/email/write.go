package email

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	mail "net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/safety"
)

const (
	maxAttachmentBytes      = 10 << 20
	maxAttachmentTotalBytes = 25 << 20
)

type sendMessage struct {
	from            string
	to              []mail.Address
	cc              []mail.Address
	bcc             []mail.Address
	subject         string
	body            string
	bodyContentType string
	attachments     []string
}

type attachment struct {
	filename string
	content  []byte
}

type preparedSend struct {
	shared      engine.PreparedWrite
	connection  connectionConfig
	message     sendMessage
	smtpAddress string
}

// ValidateWrite proves that only the one typed SMTP action and one typed
// record can enter the shared plan lifecycle. It does not read attachment
// bytes: those are read while making the preview and are then bound by the
// prepared-write digest before execution.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if req.Action != sendAction {
		return connectors.ErrUnsupportedOperation
	}
	if len(records) != 1 {
		return errors.New("email send_message accepts exactly one message; SMTP submission is non-batchable")
	}
	connection, err := resolveConnectionConfig(req.Config)
	if err != nil {
		return err
	}
	_, err = decodeSendMessage(records[0], connection.fromAddress)
	return err
}

// DryRunWrite creates the exact MIME bytes before the approval is minted.
// The complete unmasked DATA payload is deliberately included in warnings so
// the existing preview surface displays what will actually be submitted.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	prepared, err := c.prepareSend(ctx, req, records)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return engine.PreviewPreparedWrite(prepared.shared)
}

// Write regenerates the prepared bytes and delegates digest/evidence checking
// to the shared destructive-write gate before opening SMTP. This makes a file
// change after preview, a recipient change, or any MIME drift fail closed.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	prepared, err := c.prepareSend(ctx, req, records)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	preview, err := engine.PreviewPreparedWrite(prepared.shared)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	result := connectors.WriteResult{}
	err = engine.ExecutePreparedWrite(ctx, prepared.shared, req.Approval, preview.Digest, func(executeCtx context.Context) error {
		if err := submitSMTP(executeCtx, prepared.connection, prepared.smtpAddress, prepared.message, []byte(prepared.shared.Requests[0].Body)); err != nil {
			return err
		}
		result.RecordsWritten = 1
		return nil
	})
	if err != nil && result.RecordsWritten == 0 {
		result.RecordsFailed = len(records)
	}
	return result, err
}

func (c Connector) prepareSend(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (preparedSend, error) {
	if err := ctx.Err(); err != nil {
		return preparedSend{}, err
	}
	if req.Action != sendAction {
		return preparedSend{}, connectors.ErrUnsupportedOperation
	}
	if len(records) != 1 {
		return preparedSend{}, errors.New("email send_message accepts exactly one message; SMTP submission is non-batchable")
	}
	connection, err := resolveConnectionConfig(req.Config)
	if err != nil {
		return preparedSend{}, err
	}
	message, err := decodeSendMessage(records[0], connection.fromAddress)
	if err != nil {
		return preparedSend{}, err
	}
	attachments, err := loadAttachments(req.Config.ProjectDir, message.attachments)
	if err != nil {
		return preparedSend{}, err
	}
	payload, err := buildMIME(message, attachments)
	if err != nil {
		return preparedSend{}, err
	}
	envelopeRecipients := message.envelopeRecipients()
	endpoint := (&url.URL{Scheme: "smtp", Host: connection.smtpAddress()}).String()
	prepared := engine.PreparedWrite{
		Target: engine.DestructiveTarget{
			Connector:     c.Name(),
			Operation:     sendAction,
			Method:        "SMTP",
			MutationClass: "create",
			Destructive:   true,
			Confirmation:  connectors.ConfirmationKindDestructive,
		},
		CredentialRevision:  req.Config.CredentialRevision,
		ConfigurationDigest: req.Config.ConfigurationDigest,
		ApprovalScope:       req.Config.WriteApprovalScope,
		Batchable:           false,
		RecordsStaged:       1,
		Action:              sendAction,
		Warnings: []string{
			"SMTP envelope sender (unmasked): " + message.from,
			"SMTP envelope recipients (unmasked): " + strings.Join(envelopeRecipients, ", "),
			"SMTP DATA payload (unmasked):\n" + string(payload),
		},
		Definition: map[string]any{
			"protocol":            "SMTP",
			"action":              sendAction,
			"envelope_from":       message.from,
			"envelope_recipients": envelopeRecipients,
		},
		Requests: []engine.PreparedRequest{{
			Method:      "SMTP",
			URL:         endpoint,
			Target:      "MAIL FROM:<" + message.from + ">; RCPT TO:<" + strings.Join(envelopeRecipients, ">, <") + ">; DATA",
			ContentType: "message/rfc822",
			BodyFormat:  "rfc5322",
			Body:        string(payload),
		}},
	}
	return preparedSend{shared: prepared, connection: connection, message: message, smtpAddress: c.smtpAddress(connection)}, nil
}

func decodeSendMessage(record connectors.Record, defaultFrom string) (sendMessage, error) {
	for field := range record {
		switch field {
		case "to", "cc", "bcc", "subject", "body", "body_content_type", "attachments":
		default:
			return sendMessage{}, fmt.Errorf("email send_message does not accept field %q", field)
		}
	}
	to, err := recordAddresses(record, "to", true)
	if err != nil {
		return sendMessage{}, err
	}
	cc, err := recordAddresses(record, "cc", false)
	if err != nil {
		return sendMessage{}, err
	}
	bcc, err := recordAddresses(record, "bcc", false)
	if err != nil {
		return sendMessage{}, err
	}
	subject, ok := record["subject"].(string)
	if !ok {
		return sendMessage{}, errors.New("email send_message requires string field subject")
	}
	if containsControl(subject) {
		return sendMessage{}, errors.New("email send_message subject must not contain control characters")
	}
	body, ok := record["body"].(string)
	if !ok {
		return sendMessage{}, errors.New("email send_message requires string field body")
	}
	bodyContentType := "text/plain"
	if raw, present := record["body_content_type"]; present {
		var ok bool
		bodyContentType, ok = raw.(string)
		if ok && containsControl(bodyContentType) {
			return sendMessage{}, errors.New("email send_message body_content_type must not contain control characters")
		}
		if !ok || (bodyContentType != "text/plain" && bodyContentType != "text/html") {
			return sendMessage{}, errors.New("email send_message body_content_type must be text/plain or text/html")
		}
	}
	attachments, err := recordStringArray(record, "attachments", false)
	if err != nil {
		return sendMessage{}, err
	}
	return sendMessage{
		from: defaultFrom, to: to, cc: cc, bcc: bcc, subject: subject, body: body,
		bodyContentType: bodyContentType, attachments: attachments,
	}, nil
}

func recordAddresses(record connectors.Record, field string, required bool) ([]mail.Address, error) {
	values, err := recordStringArray(record, field, required)
	if err != nil {
		return nil, err
	}
	addresses := make([]mail.Address, 0, len(values))
	for _, value := range values {
		if containsControl(value) {
			return nil, fmt.Errorf("email send_message %s must not contain control characters", field)
		}
		address, err := mail.ParseAddress(value)
		if err != nil || address.Address == "" {
			return nil, fmt.Errorf("email send_message %s contains an invalid email address", field)
		}
		addresses = append(addresses, *address)
	}
	return addresses, nil
}

func recordStringArray(record connectors.Record, field string, required bool) ([]string, error) {
	raw, present := record[field]
	if !present {
		if required {
			return nil, fmt.Errorf("email send_message requires field %s", field)
		}
		return nil, nil
	}
	var values []string
	switch typed := raw.(type) {
	case []string:
		values = append([]string(nil), typed...)
	case []any:
		values = make([]string, 0, len(typed))
		for _, value := range typed {
			stringValue, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("email send_message %s must be an array of strings", field)
			}
			values = append(values, stringValue)
		}
	default:
		return nil, fmt.Errorf("email send_message %s must be an array of strings", field)
	}
	if required && len(values) == 0 {
		return nil, fmt.Errorf("email send_message requires at least one %s recipient", field)
	}
	for _, value := range values {
		if containsControl(value) {
			return nil, fmt.Errorf("email send_message %s must not contain control characters", field)
		}
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("email send_message %s must not contain an empty value", field)
		}
	}
	return values, nil
}

func (message sendMessage) envelopeRecipients() []string {
	addresses := make([]string, 0, len(message.to)+len(message.cc)+len(message.bcc))
	for _, group := range [][]mail.Address{message.to, message.cc, message.bcc} {
		for _, address := range group {
			addresses = append(addresses, address.Address)
		}
	}
	return addresses
}

func loadAttachments(runtimeRoot string, paths []string) ([]attachment, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	stagingRoot, err := attachmentStagingRoot(runtimeRoot)
	if err != nil {
		return nil, err
	}
	rootPath, err := filepath.Abs(stagingRoot)
	if err != nil {
		return nil, errors.New("email attachment staging root is unavailable")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, errors.New("email attachment staging root is unavailable")
	}
	defer func() { _ = root.Close() }()
	attachments := make([]attachment, 0, len(paths))
	total := int64(0)
	for _, path := range paths {
		relative, err := attachmentStagingRelativePath(path)
		if err != nil {
			return nil, err
		}
		file, err := root.Open(relative)
		if err != nil {
			return nil, errors.New("email attachment path cannot be resolved")
		}
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			_ = file.Close()
			return nil, errors.New("email attachment path must be a readable regular file")
		}
		if info.Size() > maxAttachmentBytes {
			_ = file.Close()
			return nil, fmt.Errorf("email attachment exceeds the %d-byte per-file limit", maxAttachmentBytes)
		}
		if total+info.Size() > maxAttachmentTotalBytes {
			_ = file.Close()
			return nil, fmt.Errorf("email attachments exceed the %d-byte total limit", maxAttachmentTotalBytes)
		}
		content, readErr := io.ReadAll(io.LimitReader(file, maxAttachmentBytes+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			return nil, errors.New("email attachment could not be read")
		}
		if len(content) > maxAttachmentBytes {
			return nil, fmt.Errorf("email attachment exceeds the %d-byte per-file limit", maxAttachmentBytes)
		}
		if int64(len(content)) != info.Size() {
			return nil, errors.New("email attachment changed while preparing the preview")
		}
		total += int64(len(content))
		attachments = append(attachments, attachment{filename: filepath.Base(relative), content: content})
	}
	return attachments, nil
}

func attachmentStagingRoot(runtimeRoot string) (string, error) {
	if strings.TrimSpace(runtimeRoot) == "" {
		return "", errors.New("email attachment staging root is unavailable")
	}
	return runtimeRoot, nil
}

func attachmentStagingRelativePath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("email attachment path is required")
	}
	if err := safety.RejectDangerousChars(raw, "email attachment path"); err != nil {
		return "", err
	}
	clean := filepath.Clean(filepath.FromSlash(raw))
	if filepath.IsAbs(clean) || !filepath.IsLocal(clean) {
		return "", errors.New("email attachment path must be relative to the attachment staging root")
	}
	return clean, nil
}

func buildMIME(message sendMessage, attachments []attachment) ([]byte, error) {
	var body bytes.Buffer
	if len(attachments) == 0 {
		writeHeader(&body, "From", message.from)
		writeAddressesHeader(&body, "To", message.to)
		if len(message.cc) > 0 {
			writeAddressesHeader(&body, "Cc", message.cc)
		}
		writeHeader(&body, "Subject", mime.QEncoding.Encode("utf-8", message.subject))
		writeHeader(&body, "MIME-Version", "1.0")
		writeHeader(&body, "Content-Type", message.bodyContentType+`; charset="UTF-8"`)
		writeHeader(&body, "Content-Transfer-Encoding", "quoted-printable")
		body.WriteString("\r\n")
		if err := writeQuotedPrintable(&body, normalizeCRLF(message.body)); err != nil {
			return nil, err
		}
		return body.Bytes(), nil
	}
	boundary := deterministicBoundary(message, attachments)
	writeHeader(&body, "From", message.from)
	writeAddressesHeader(&body, "To", message.to)
	if len(message.cc) > 0 {
		writeAddressesHeader(&body, "Cc", message.cc)
	}
	writeHeader(&body, "Subject", mime.QEncoding.Encode("utf-8", message.subject))
	writeHeader(&body, "MIME-Version", "1.0")
	writeHeader(&body, "Content-Type", `multipart/mixed; boundary="`+boundary+`"`)
	body.WriteString("\r\n")
	body.WriteString("--" + boundary + "\r\n")
	writeHeader(&body, "Content-Type", message.bodyContentType+`; charset="UTF-8"`)
	writeHeader(&body, "Content-Transfer-Encoding", "quoted-printable")
	body.WriteString("\r\n")
	if err := writeQuotedPrintable(&body, normalizeCRLF(message.body)); err != nil {
		return nil, err
	}
	if !bytes.HasSuffix(body.Bytes(), []byte("\r\n")) {
		body.WriteString("\r\n")
	}
	for _, attachment := range attachments {
		body.WriteString("--" + boundary + "\r\n")
		mediaType := mime.FormatMediaType("application/octet-stream", map[string]string{"name": attachment.filename})
		disposition := mime.FormatMediaType("attachment", map[string]string{"filename": attachment.filename})
		writeHeader(&body, "Content-Type", mediaType)
		writeHeader(&body, "Content-Transfer-Encoding", "base64")
		writeHeader(&body, "Content-Disposition", disposition)
		body.WriteString("\r\n")
		writeBase64Lines(&body, attachment.content)
	}
	body.WriteString("--" + boundary + "--\r\n")
	return body.Bytes(), nil
}

func writeHeader(buffer *bytes.Buffer, name, value string) {
	buffer.WriteString(name)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString("\r\n")
}

func writeAddressesHeader(buffer *bytes.Buffer, name string, addresses []mail.Address) {
	values := make([]string, 0, len(addresses))
	for _, address := range addresses {
		values = append(values, address.String())
	}
	writeHeader(buffer, name, strings.Join(values, ", "))
}

func writeQuotedPrintable(buffer *bytes.Buffer, content string) error {
	encoder := quotedprintable.NewWriter(buffer)
	if _, err := encoder.Write([]byte(content)); err != nil {
		return err
	}
	return encoder.Close()
}

func normalizeCRLF(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}

func writeBase64Lines(buffer *bytes.Buffer, content []byte) {
	encoded := base64.StdEncoding.EncodeToString(content)
	for len(encoded) > 76 {
		buffer.WriteString(encoded[:76])
		buffer.WriteString("\r\n")
		encoded = encoded[76:]
	}
	buffer.WriteString(encoded)
	buffer.WriteString("\r\n")
}

func deterministicBoundary(message sendMessage, attachments []attachment) string {
	hash := sha256.New()
	for _, value := range append([]string{message.from, message.subject, message.bodyContentType, message.body}, message.envelopeRecipients()...) {
		hash.Write([]byte(value))
		hash.Write([]byte{0})
	}
	for _, attachment := range attachments {
		hash.Write([]byte(attachment.filename))
		hash.Write([]byte{0})
		hash.Write(attachment.content)
		hash.Write([]byte{0})
	}
	return "pm-" + fmt.Sprintf("%x", hash.Sum(nil))[:40]
}

func (c Connector) checkSMTP(ctx context.Context, connection connectionConfig) error {
	client, err := openSMTP(ctx, connection, c.smtpAddress(connection))
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()
	return authenticateSMTP(client, connection)
}

func submitSMTP(ctx context.Context, connection connectionConfig, address string, message sendMessage, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	client, err := openSMTP(ctx, connection, address)
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()
	if err := authenticateSMTP(client, connection); err != nil {
		return err
	}
	if err := client.Mail(message.from); err != nil {
		return errors.New("SMTP MAIL command failed")
	}
	for _, recipient := range message.envelopeRecipients() {
		if err := client.Rcpt(recipient); err != nil {
			return errors.New("SMTP RCPT command failed")
		}
	}
	writer, err := client.Data()
	if err != nil {
		return errors.New("SMTP DATA command failed")
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return errors.New("SMTP DATA payload write failed")
	}
	if err := writer.Close(); err != nil {
		return errors.New("SMTP server rejected the message")
	}
	return nil
}

func openSMTP(ctx context.Context, connection connectionConfig, address string) (*smtp.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: connection.timeout}
	tlsConfig := &tls.Config{ServerName: connection.smtpHost, MinVersion: tls.VersionTLS12}
	var (
		conn net.Conn
		err  error
	)
	switch connection.smtpSecurity {
	case securityTLS:
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	case securitySTARTTLS, securityNone:
		conn, err = dialer.DialContext(ctx, "tcp", address)
	default:
		return nil, errors.New("email config smtp_security must satisfy its declared enum constraint")
	}
	if err != nil {
		return nil, fmt.Errorf("connect to SMTP server: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(connection.timeout))
	client, err := smtp.NewClient(conn, connection.smtpHost)
	if err != nil {
		_ = conn.Close()
		return nil, errors.New("SMTP handshake failed")
	}
	if connection.smtpSecurity == securitySTARTTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			_ = client.Close()
			return nil, errors.New("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, errors.New("SMTP STARTTLS failed")
		}
	}
	return client, nil
}

func authenticateSMTP(client *smtp.Client, connection connectionConfig) error {
	ok, mechanisms := client.Extension("AUTH")
	if !ok {
		return errors.New("SMTP server does not advertise password authentication")
	}
	available := map[string]bool{}
	for _, mechanism := range strings.Fields(strings.ToUpper(mechanisms)) {
		available[mechanism] = true
	}
	var auth smtp.Auth
	switch {
	case available["PLAIN"]:
		auth = smtp.PlainAuth("", connection.smtpUsername, connection.password, connection.smtpHost)
	case available["LOGIN"]:
		auth = &loginAuth{username: connection.smtpUsername, password: connection.password, host: connection.smtpHost}
	default:
		return errors.New("SMTP server does not advertise a supported password authentication mechanism")
	}
	if err := client.Auth(auth); err != nil {
		return errors.New("SMTP authentication failed")
	}
	return nil
}

type loginAuth struct {
	username string
	password string
	host     string
	step     int
}

func (auth *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !isLoopbackHost(auth.host) {
		return "", nil, errors.New("LOGIN authentication requires TLS for a non-loopback host")
	}
	return "LOGIN", nil, nil
}

func (auth *loginAuth) Next(_ []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch auth.step {
	case 0:
		auth.step++
		return []byte(auth.username), nil
	case 1:
		auth.step++
		return []byte(auth.password), nil
	default:
		return nil, errors.New("unexpected SMTP LOGIN challenge")
	}
}

var _ connectors.WriteValidator = Connector{}
var _ connectors.DryRunWriter = Connector{}
