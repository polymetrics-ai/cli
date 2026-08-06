package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/emersion/go-imap/v2/imapclient"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultMessageLimit = 100
	maxMessageLimit     = 1000
)

// Check authenticates to each configured protocol without changing mailbox
// data or submitting a message. Authentication errors are intentionally
// summarized rather than forwarding a server string that could contain an
// echoed credential.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	connection, err := resolveConnectionConfig(cfg)
	if err != nil {
		return err
	}
	client, err := c.openIMAP(ctx, connection)
	if err != nil {
		return err
	}
	defer closeIMAP(client)
	if err := client.Noop().Wait(); err != nil {
		return errors.New("email IMAP check failed")
	}
	return c.checkSMTP(ctx, connection)
}

// Read dispatches only the IMAP streams declared in the bundle. SMTP is never
// used here: it is a send-only protocol and has no read capability.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if req.Stream != mailboxesStream {
		return connectors.ErrUnsupportedOperation
	}
	connection, err := resolveConnectionConfig(req.Config)
	if err != nil {
		return err
	}
	client, err := c.openIMAP(ctx, connection)
	if err != nil {
		return err
	}
	defer closeIMAP(client)

	return readMailboxes(ctx, client, boundedMessageLimit(req.Limit), emit)
}

func (c Connector) openIMAP(ctx context.Context, connection connectionConfig) (*imapclient.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	options := &imapclient.Options{
		TLSConfig: &tls.Config{ServerName: connection.imapHost, MinVersion: tls.VersionTLS12},
		Dialer:    &net.Dialer{Timeout: connection.timeout},
		// DebugWriter remains nil because the library documents that it would
		// include LOGIN credentials in raw protocol traffic.
	}
	var (
		client *imapclient.Client
		err    error
	)
	switch connection.imapSecurity {
	case securityTLS:
		client, err = imapclient.DialTLS(c.imapAddress(connection), options)
	case securitySTARTTLS:
		client, err = imapclient.DialStartTLS(c.imapAddress(connection), options)
	case securityNone:
		client, err = imapclient.DialInsecure(c.imapAddress(connection), options)
	default:
		return nil, errors.New("email config imap_security must satisfy its declared enum constraint")
	}
	if err != nil {
		return nil, fmt.Errorf("connect to IMAP server: %w", err)
	}
	if err := client.Login(connection.username, connection.password).Wait(); err != nil {
		_ = client.Close()
		return nil, errors.New("IMAP authentication failed")
	}
	return client, nil
}

func closeIMAP(client *imapclient.Client) {
	if client == nil {
		return
	}
	_ = client.Logout().Wait()
	_ = client.Close()
}

func readMailboxes(ctx context.Context, client *imapclient.Client, limit int, emit func(connectors.Record) error) error {
	command := client.List("", "*", nil)
	for emitted := 0; emitted < limit; emitted++ {
		mailbox := command.Next()
		if mailbox == nil {
			break
		}
		if err := ctx.Err(); err != nil {
			_ = command.Close()
			return err
		}
		attributes := make([]string, 0, len(mailbox.Attrs))
		for _, attribute := range mailbox.Attrs {
			attributes = append(attributes, string(attribute))
		}
		record := connectors.Record{
			"name":       mailbox.Mailbox,
			"attributes": attributes,
		}
		if mailbox.Delim == 0 {
			record["delimiter"] = nil
		} else {
			record["delimiter"] = string(mailbox.Delim)
		}
		if err := emit(record); err != nil {
			_ = command.Close()
			return err
		}
	}
	if err := command.Close(); err != nil {
		return errors.New("IMAP LIST failed")
	}
	return nil
}

func boundedMessageLimit(limit int) int {
	if limit <= 0 {
		return defaultMessageLimit
	}
	if limit > maxMessageLimit {
		return maxMessageLimit
	}
	return limit
}
