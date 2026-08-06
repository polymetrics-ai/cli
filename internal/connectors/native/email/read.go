package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultMessageLimit = 100
	maxMessageLimit     = 1000
	uidSearchWindowSize = 1000
	maxBodyPartBytes    = 1 << 20
	maxBodyParts        = 32
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
	connection, err := resolveConnectionConfig(req.Config)
	if err != nil {
		return err
	}
	client, err := c.openIMAP(ctx, connection)
	if err != nil {
		return err
	}
	defer closeIMAP(client)

	switch req.Stream {
	case mailboxesStream:
		return readMailboxes(ctx, client, boundedMessageLimit(req.Limit), emit)
	case messagesStream:
		return readMessages(ctx, client, connection.mailbox, req.State, req.Limit, emit)
	default:
		return connectors.ErrUnsupportedOperation
	}
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

func readMessages(ctx context.Context, client *imapclient.Client, mailbox string, state map[string]string, requestedLimit int, emit func(connectors.Record) error) error {
	selected, err := client.Select(mailbox, &imap.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return errors.New("IMAP mailbox selection failed")
	}
	prior, err := decodeCursor(stateValue(state, "cursor"), mailbox)
	if err != nil {
		return err
	}
	// A changed UIDVALIDITY means the server has replaced this mailbox's UID
	// namespace. RFC 9051 requires clients to discard the old UID state; start
	// from the new namespace rather than silently comparing dates.
	lowerUID := prior.uid
	if prior.uidValidity != 0 && prior.uidValidity != selected.UIDValidity {
		lowerUID = 0
	}
	uids, err := searchUIDsAfter(client, lowerUID, selected.UIDNext, boundedMessageLimit(requestedLimit))
	if err != nil {
		return err
	}
	if len(uids) == 0 {
		return nil
	}
	metadata, err := fetchMessageMetadata(client, uids)
	if err != nil {
		return err
	}
	for _, message := range metadata {
		if err := ctx.Err(); err != nil {
			return err
		}
		if uint32(message.UID) <= lowerUID {
			continue
		}
		parts, err := fetchBodyParts(client, message.UID, message.BodyStructure)
		if err != nil {
			return err
		}
		record := connectors.Record{
			"mailbox":      mailbox,
			"uid_validity": strconv.FormatUint(uint64(selected.UIDValidity), 10),
			"uid":          strconv.FormatUint(uint64(message.UID), 10),
			"imap_cursor":  encodeCursor(mailbox, selected.UIDValidity, uint32(message.UID)),
			"envelope":     envelopeRecord(message.Envelope),
			"flags":        flagsRecord(message.Flags),
			"body_parts":   parts,
		}
		if message.InternalDate.IsZero() {
			record["internal_date"] = nil
		} else {
			record["internal_date"] = message.InternalDate.UTC().Format(time.RFC3339Nano)
		}
		if message.RFC822Size <= 0 {
			record["size"] = nil
		} else {
			record["size"] = message.RFC822Size
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func stateValue(state map[string]string, key string) string {
	if state == nil {
		return ""
	}
	return state[key]
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

func searchUIDsAfter(client *imapclient.Client, lowerUID uint32, uidNext imap.UID, limit int) ([]imap.UID, error) {
	if lowerUID == math.MaxUint32 || limit <= 0 {
		return nil, nil
	}
	if uidNext == 0 {
		return nil, errors.New("IMAP mailbox did not provide UIDNEXT")
	}
	highestUID := uint32(uidNext) - 1
	if lowerUID >= highestUID {
		return nil, nil
	}
	start := lowerUID + 1
	filtered := make([]imap.UID, 0, limit)
	for start <= highestUID && len(filtered) < limit {
		stop := uidSearchWindowEnd(start, highestUID)
		var set imap.UIDSet
		set.AddRange(imap.UID(start), imap.UID(stop))
		data, err := client.UIDSearch(&imap.SearchCriteria{UID: []imap.UIDSet{set}}, nil).Wait()
		if err != nil {
			return nil, errors.New("IMAP UID SEARCH failed")
		}
		uids := append([]imap.UID(nil), data.AllUIDs()...)
		sort.Slice(uids, func(i, j int) bool { return uids[i] < uids[j] })
		for _, uid := range uids {
			if uint32(uid) < start || uint32(uid) > stop {
				continue
			}
			filtered = append(filtered, uid)
			if len(filtered) == limit {
				break
			}
		}
		if stop == highestUID {
			break
		}
		start = stop + 1
	}
	return filtered, nil
}

func uidSearchWindowEnd(start, highest uint32) uint32 {
	if highest-start < uidSearchWindowSize-1 {
		return highest
	}
	return start + uidSearchWindowSize - 1
}

func fetchMessageMetadata(client *imapclient.Client, uids []imap.UID) ([]*imapclient.FetchMessageBuffer, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	set := imap.UIDSetNum(uids...)
	result, err := client.Fetch(set, &imap.FetchOptions{
		UID:           true,
		Envelope:      true,
		Flags:         true,
		InternalDate:  true,
		RFC822Size:    true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
	}).Collect()
	if err != nil {
		return nil, errors.New("IMAP UID FETCH metadata failed")
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UID < result[j].UID })
	return result, nil
}

type bodyPartRequest struct {
	section     *imap.FetchItemBodySection
	path        string
	mediaType   string
	filename    string
	disposition string
	size        int64
}

func fetchBodyParts(client *imapclient.Client, uid imap.UID, structure imap.BodyStructure) ([]connectors.Record, error) {
	requests := bodyPartRequests(structure)
	if len(requests) == 0 {
		return []connectors.Record{}, nil
	}
	sections := make([]*imap.FetchItemBodySection, 0, len(requests))
	bySection := make(map[string]bodyPartRequest, len(requests))
	for _, request := range requests {
		sections = append(sections, request.section)
		bySection[bodySectionKey(request.section)] = request
	}
	command := client.Fetch(imap.UIDSetNum(uid), &imap.FetchOptions{UID: true, BodySection: sections})
	parts := make([]connectors.Record, 0, len(requests))
	var fetchErr error
	for {
		message := command.Next()
		if message == nil {
			break
		}
		for {
			item := message.Next()
			if item == nil {
				break
			}
			body, ok := item.(imapclient.FetchItemDataBodySection)
			if !ok || body.Section == nil {
				continue
			}
			request, ok := bySection[bodySectionKey(body.Section)]
			if !ok {
				continue
			}
			content, err := readBoundedLiteral(body.Literal, maxBodyPartBytes)
			if err != nil {
				fetchErr = err
				break
			}
			part := connectors.Record{
				"path":           request.path,
				"content_type":   request.mediaType,
				"declared_size":  request.size,
				"bytes_returned": len(content),
				"content_base64": base64.StdEncoding.EncodeToString(content),
				"truncated":      request.size > int64(len(content)),
			}
			if request.filename != "" {
				part["filename"] = request.filename
			}
			if request.disposition != "" {
				part["disposition"] = request.disposition
			}
			parts = append(parts, part)
		}
		if fetchErr != nil {
			break
		}
	}
	if err := command.Close(); err != nil && fetchErr == nil {
		fetchErr = err
	}
	if fetchErr != nil {
		return nil, errors.New("IMAP body-part fetch exceeded its configured bound or failed")
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i]["path"].(string) < parts[j]["path"].(string) })
	return parts, nil
}

func readBoundedLiteral(literal imap.LiteralReader, maximum int) ([]byte, error) {
	if literal == nil {
		return nil, nil
	}
	if literal.Size() > int64(maximum) {
		return nil, errors.New("IMAP body part exceeds configured bound")
	}
	content, err := io.ReadAll(io.LimitReader(literal, int64(maximum)+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maximum {
		return nil, errors.New("IMAP body part exceeds configured bound")
	}
	return content, nil
}

func bodyPartRequests(structure imap.BodyStructure) []bodyPartRequest {
	partial := func(part []int) *imap.FetchItemBodySection {
		return &imap.FetchItemBodySection{
			Part:    append([]int(nil), part...),
			Peek:    true,
			Partial: &imap.SectionPartial{Offset: 0, Size: maxBodyPartBytes},
		}
	}
	if structure == nil {
		section := partial(nil)
		section.Specifier = imap.PartSpecifierText
		return []bodyPartRequest{{section: section, path: "TEXT", mediaType: "application/octet-stream"}}
	}
	if single, ok := structure.(*imap.BodyStructureSinglePart); ok {
		section := partial(nil)
		section.Specifier = imap.PartSpecifierText
		return []bodyPartRequest{{
			section: section, path: "TEXT", mediaType: single.MediaType(), filename: single.Filename(),
			disposition: bodyDisposition(single), size: int64(single.Size),
		}}
	}
	requests := make([]bodyPartRequest, 0)
	structure.Walk(func(path []int, part imap.BodyStructure) bool {
		if len(requests) >= maxBodyParts {
			return false
		}
		single, ok := part.(*imap.BodyStructureSinglePart)
		if !ok {
			return true
		}
		section := partial(path)
		requests = append(requests, bodyPartRequest{
			section: section, path: dottedPartPath(path), mediaType: single.MediaType(), filename: single.Filename(),
			disposition: bodyDisposition(single), size: int64(single.Size),
		})
		return true
	})
	return requests
}

func bodyDisposition(part imap.BodyStructure) string {
	if disposition := part.Disposition(); disposition != nil {
		return disposition.Value
	}
	return ""
}

func dottedPartPath(path []int) string {
	parts := make([]string, 0, len(path))
	for _, value := range path {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ".")
}

func bodySectionKey(section *imap.FetchItemBodySection) string {
	return dottedPartPath(section.Part) + "|" + string(section.Specifier)
}

func envelopeRecord(envelope *imap.Envelope) any {
	if envelope == nil {
		return nil
	}
	record := connectors.Record{
		"subject":     envelope.Subject,
		"from":        addressesRecord(envelope.From),
		"sender":      addressesRecord(envelope.Sender),
		"reply_to":    addressesRecord(envelope.ReplyTo),
		"to":          addressesRecord(envelope.To),
		"cc":          addressesRecord(envelope.Cc),
		"bcc":         addressesRecord(envelope.Bcc),
		"in_reply_to": append([]string(nil), envelope.InReplyTo...),
		"message_id":  envelope.MessageID,
	}
	if envelope.Date.IsZero() {
		record["date"] = nil
	} else {
		record["date"] = envelope.Date.UTC().Format(time.RFC3339Nano)
	}
	return record
}

func addressesRecord(addresses []imap.Address) []connectors.Record {
	result := make([]connectors.Record, 0, len(addresses))
	for _, address := range addresses {
		result = append(result, connectors.Record{"name": address.Name, "address": address.Addr()})
	}
	return result
}

func flagsRecord(flags []imap.Flag) []string {
	result := make([]string, 0, len(flags))
	for _, flag := range flags {
		result = append(result, string(flag))
	}
	sort.Strings(result)
	return result
}
