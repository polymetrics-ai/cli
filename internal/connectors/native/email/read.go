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

type imapConnection struct {
	client  *imapclient.Client
	ctx     context.Context
	stop    func() bool
	aborted bool
}

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
	if err := client.client.Noop().Wait(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
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

func (c Connector) openIMAP(ctx context.Context, connection connectionConfig) (*imapConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: connection.timeout}
	tlsConfig := &tls.Config{ServerName: connection.imapHost, MinVersion: tls.VersionTLS12}
	options := &imapclient.Options{
		TLSConfig: tlsConfig,
		// DebugWriter remains nil because the library documents that it would
		// include LOGIN credentials in raw protocol traffic.
	}
	var (
		conn net.Conn
		err  error
	)
	switch connection.imapSecurity {
	case securityTLS:
		tlsConfig = tlsConfig.Clone()
		if tlsConfig.NextProtos == nil {
			tlsConfig.NextProtos = []string{"imap"}
		}
		options.TLSConfig = tlsConfig
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", c.imapAddress(connection))
	case securitySTARTTLS, securityNone:
		conn, err = dialer.DialContext(ctx, "tcp", c.imapAddress(connection))
	default:
		return nil, errors.New("email config imap_security must satisfy its declared enum constraint")
	}
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("connect to IMAP server: %w", err)
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	if ctxErr := ctx.Err(); ctxErr != nil {
		_ = stop()
		_ = conn.Close()
		return nil, ctxErr
	}
	var client *imapclient.Client
	switch connection.imapSecurity {
	case securityTLS, securityNone:
		client = imapclient.New(conn, options)
	case securitySTARTTLS:
		client, err = imapclient.NewStartTLS(conn, options)
	}
	if err != nil {
		_ = stop()
		_ = conn.Close()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("connect to IMAP server: %w", err)
	}
	opened := &imapConnection{client: client, ctx: ctx, stop: stop}
	if err := client.Login(connection.username, connection.password).Wait(); err != nil {
		closeIMAP(opened)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, errors.New("IMAP authentication failed")
	}
	return opened, nil
}

func closeIMAP(connection *imapConnection) {
	if connection == nil || connection.client == nil {
		return
	}
	if !connection.aborted && connection.ctx.Err() == nil {
		_ = connection.client.Logout().Wait()
	}
	abortIMAP(connection)
}

func abortIMAP(connection *imapConnection) {
	if connection == nil || connection.client == nil {
		return
	}
	connection.aborted = true
	if connection.stop != nil {
		_ = connection.stop()
	}
	_ = connection.client.Close()
}

func readMailboxes(ctx context.Context, connection *imapConnection, limit int, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if connection == nil || connection.client == nil {
		return errors.New("IMAP connection is unavailable")
	}
	client := connection.client
	command := client.List("", "*", nil)
	for emitted := 0; emitted < limit; emitted++ {
		mailbox := command.Next()
		if err := ctx.Err(); err != nil {
			abortIMAP(connection)
			return err
		}
		if mailbox == nil {
			if err := command.Close(); err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}
				return errors.New("IMAP LIST failed")
			}
			return nil
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
			abortIMAP(connection)
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		abortIMAP(connection)
		return err
	}
	abortIMAP(connection)
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
