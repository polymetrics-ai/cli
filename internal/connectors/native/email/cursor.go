package email

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Cursor encodes the identity that RFC 9051 uses for incremental mailbox
// synchronization: the mailbox's UIDVALIDITY namespace marker and a UID in
// that namespace. It deliberately does not use INTERNALDATE or received-date
// timestamps, which are not mailbox identities or a reliable monotonic order.
//
// Fixed-width decimal fields preserve lexical ordering for the generic ETL
// cursor store. The mailbox component prevents accidental state reuse between
// distinct mailbox commands or credentials.
type cursor struct {
	mailbox     string
	uidValidity uint32
	uid         uint32
}

func encodeCursor(mailbox string, uidValidity, uid uint32) string {
	encodedMailbox := base64.RawURLEncoding.EncodeToString([]byte(mailbox))
	return fmt.Sprintf("imapv1:%s:%010d:%010d", encodedMailbox, uidValidity, uid)
}

func decodeCursor(raw, mailbox string) (cursor, error) {
	if strings.TrimSpace(raw) == "" {
		return cursor{mailbox: mailbox}, nil
	}
	parts := strings.Split(raw, ":")
	if len(parts) != 4 || parts[0] != "imapv1" {
		return cursor{}, errors.New("email messages cursor is invalid; reset the mailbox state before retrying")
	}
	decodedMailbox, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || string(decodedMailbox) == "" {
		return cursor{}, errors.New("email messages cursor is invalid; reset the mailbox state before retrying")
	}
	if string(decodedMailbox) != mailbox {
		return cursor{}, errors.New("email messages cursor belongs to a different mailbox; use a separate credential/state or reset this mailbox state")
	}
	uidValidity, err := parseCursorUint(parts[2])
	if err != nil {
		return cursor{}, errors.New("email messages cursor is invalid; reset the mailbox state before retrying")
	}
	uid, err := parseCursorUint(parts[3])
	if err != nil {
		return cursor{}, errors.New("email messages cursor is invalid; reset the mailbox state before retrying")
	}
	return cursor{mailbox: mailbox, uidValidity: uidValidity, uid: uid}, nil
}

func parseCursorUint(raw string) (uint32, error) {
	if len(raw) != 10 {
		return 0, errors.New("not fixed width")
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, errors.New("not decimal")
		}
	}
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}
