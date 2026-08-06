// Package email implements the protocol-level Email connector.
//
// The connector is deliberately split by protocol responsibility: IMAP is
// the read path (mailbox and message streams) and SMTP is a typed, approval
// gated send-only write path. It is not a wrapper around Gmail or Outlook's
// HTTP APIs, and SMTP never backs a read capability.
package email

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	mailboxesStream = "mailboxes"
	messagesStream  = "messages"
	sendAction      = "send_message"
)

// Connector is a Tier-3 native connector because IMAP and SMTP are stateful
// wire protocols that the declarative HTTP engine cannot execute. Identity,
// configuration constraints, schemas, and command metadata remain in the
// email defs bundle through engine.Base.
type Connector struct {
	engine.Base

	// The address overrides are package-private test seams for local protocol
	// doubles. They are never populated from configuration or command input:
	// production dials only the validated ordinary-mail-client host and port.
	imapAddressOverride string
	smtpAddressOverride string
}

func (c Connector) imapAddress(connection connectionConfig) string {
	if c.imapAddressOverride != "" {
		return c.imapAddressOverride
	}
	return connection.imapAddress()
}

func (c Connector) smtpAddress(connection connectionConfig) string {
	if c.smtpAddressOverride != "" {
		return c.smtpAddressOverride
	}
	return connection.smtpAddress()
}

// New constructs the Email connector from its embedded bundle. A failure to
// load the bundle is a build invariant violation rather than a user runtime
// error, matching the other native connectors.
func New() Connector {
	b, err := engine.Load(defs.FS, "email")
	if err != nil {
		panic("native/email: failed to load defs/email bundle: " + err.Error())
	}
	return Connector{Base: engine.NewBase(b)}
}

// Manifest supplies the typed native write contract consumed by the shared
// command plan/preview/approval gate. SMTP send is explicitly non-batchable
// and destructive once the server accepts DATA.
func (c Connector) Manifest() connectors.Manifest {
	batchable := false
	return connectors.Manifest{
		Metadata: c.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "imap_host", Required: true},
			{Name: "imap_port", Required: true},
			{Name: "imap_security", Required: true},
			{Name: "smtp_host", Required: true},
			{Name: "smtp_port", Required: true},
			{Name: "smtp_security", Required: true},
			{Name: "username", Required: true},
			{Name: "smtp_username"},
			{Name: "from_address"},
			{Name: "connection_timeout_seconds", Default: "30"},
			{Name: "mailbox", Default: "INBOX"},
		},
		SecretFields: []connectors.SecretField{{Name: "password", Required: true}},
		Streams: []connectors.Stream{
			{
				Name:        mailboxesStream,
				Description: "Mailboxes returned by IMAP LIST.",
				PrimaryKey:  []string{"name"},
				Fields: []connectors.Field{
					{Name: "name", Type: "string"},
					{Name: "delimiter", Type: "string"},
					{Name: "attributes", Type: "array"},
				},
			},
			{
				Name:         messagesStream,
				Description:  "Messages from one IMAP mailbox, keyed and incremented by mailbox UIDVALIDITY plus UID. Hard deletions are not observable by polling.",
				PrimaryKey:   []string{"mailbox", "uid_validity", "uid"},
				CursorFields: []string{"imap_cursor"},
				Fields: []connectors.Field{
					{Name: "mailbox", Type: "string"},
					{Name: "uid_validity", Type: "string"},
					{Name: "uid", Type: "string"},
					{Name: "imap_cursor", Type: "string"},
					{Name: "envelope", Type: "object"},
					{Name: "flags", Type: "array"},
					{Name: "internal_date", Type: "string"},
					{Name: "size", Type: "integer"},
					{Name: "body_parts", Type: "array"},
				},
			},
		},
		WriteActions: []connectors.WriteActionSpec{{
			Name:           sendAction,
			Description:    "Submit one RFC 5322/MIME message through SMTP.",
			RequiredFields: []string{"to", "subject", "body"},
			OptionalFields: []string{"cc", "bcc", "body_content_type", "attachments"},
			Method:         "SMTP",
			Path:           "MAIL/RCPT/DATA",
			Risk:           "submits one externally visible SMTP message; it cannot be undone after the server accepts DATA",
			Batchable:      &batchable,
			Confirm:        string(connectors.ConfirmationKindDestructive),
		}},
		SyncModes:       []string{"full_refresh_append", "incremental_append"},
		SourceSyncModes: []string{"full_refresh", "incremental"},
		Risk: connectors.RiskSpec{
			Read:     "polled IMAP reads; hard deletion is not observable by polling",
			Write:    "SMTP send-only submission",
			Approval: "plan, unmasked preview, typed destructive confirmation, and approval are required before SMTP submission",
		},
	}
}

// Catalog returns the connector's fixed protocol-level streams. It does not
// contact a mail server; credentials are only used by Check and Read.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: c.Manifest().Streams}, nil
}

// InitialState establishes the scalar state shape used by generic ETL state
// persistence. The mailbox name is part of the encoded value, so state cannot
// accidentally be reused for a different mailbox.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if stream != messagesStream {
		return map[string]string{}, nil
	}
	mailbox, err := mailboxFromConfig(cfg.Config)
	if err != nil {
		return nil, err
	}
	return map[string]string{"cursor": encodeCursor(mailbox, 0, 0)}, nil
}

var _ connectors.Connector = Connector{}
var _ connectors.ManifestProvider = Connector{}
var _ connectors.StatefulReader = Connector{}
