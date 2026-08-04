package vault

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrWriteApprovalConsumed = errors.New("write approval has already been consumed")

type WriteApprovalRoot struct {
	dir string
	key [sha256.Size]byte
}

func (v *Vault) WriteApprovalRoot() WriteApprovalRoot {
	if v == nil {
		return WriteApprovalRoot{}
	}
	mac := hmac.New(sha256.New, v.key)
	_, _ = mac.Write([]byte("polymetrics/write-approval-root/v2"))
	var key [sha256.Size]byte
	copy(key[:], mac.Sum(nil))
	return WriteApprovalRoot{dir: v.dir, key: key}
}

func (r WriteApprovalRoot) Valid() bool {
	return strings.TrimSpace(r.dir) != "" && r.key != [sha256.Size]byte{}
}

func (r WriteApprovalRoot) AuthorityID() (string, error) {
	return r.Authenticate([]byte("authority-id-v1"))
}

func (r WriteApprovalRoot) Authenticate(payload []byte) (string, error) {
	if !r.Valid() {
		return "", errors.New("write approval root is invalid")
	}
	mac := hmac.New(sha256.New, r.key[:])
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (r WriteApprovalRoot) Consumed(approvalID string) (bool, error) {
	path, err := r.consumptionMarkerPath(approvalID)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("inspect write approval consumption marker: %w", err)
}

func (r WriteApprovalRoot) Consume(approvalID, nonce, grantMAC string, consumedAt time.Time) error {
	if !r.Valid() {
		return errors.New("write approval root is invalid")
	}
	if strings.TrimSpace(approvalID) == "" || strings.TrimSpace(nonce) == "" || strings.TrimSpace(grantMAC) == "" {
		return errors.New("write approval consumption identity is incomplete")
	}
	path, err := r.consumptionMarkerPath(approvalID)
	if err != nil {
		return err
	}
	markerID := strings.TrimSuffix(filepath.Base(path), ".used")
	nonceID, err := r.Authenticate(append([]byte("consumed-nonce-v1\x00"), []byte(nonce)...))
	if err != nil {
		return err
	}
	dir := filepath.Join(r.dir, "write-approval-consumed")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create write approval consumption directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return ErrWriteApprovalConsumed
	}
	if err != nil {
		return fmt.Errorf("create write approval consumption marker: %w", err)
	}
	record := struct {
		Version     int       `json:"version"`
		AuthorityID string    `json:"authority_id"`
		MarkerID    string    `json:"marker_id"`
		NonceID     string    `json:"nonce_id"`
		GrantMAC    string    `json:"grant_mac"`
		ConsumedAt  time.Time `json:"consumed_at"`
		MAC         string    `json:"mac"`
	}{Version: 1, MarkerID: markerID, NonceID: nonceID, GrantMAC: grantMAC, ConsumedAt: consumedAt.UTC()}
	record.AuthorityID, err = r.AuthorityID()
	if err == nil {
		var unsigned []byte
		unsigned, err = json.Marshal(record)
		if err == nil {
			record.MAC, err = r.Authenticate(append([]byte("consumed-record-v1\x00"), unsigned...))
		}
	}
	var payload []byte
	if err == nil {
		payload, err = json.Marshal(record)
		payload = append(payload, '\n')
	}
	if err == nil {
		_, err = file.Write(payload)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("write write approval consumption marker: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close write approval consumption marker: %w", closeErr)
	}
	if dirHandle, openErr := os.Open(dir); openErr == nil {
		_ = dirHandle.Sync()
		_ = dirHandle.Close()
	}
	return nil
}

func (r WriteApprovalRoot) consumptionMarkerPath(approvalID string) (string, error) {
	if !r.Valid() {
		return "", errors.New("write approval root is invalid")
	}
	if strings.TrimSpace(approvalID) == "" {
		return "", errors.New("write approval identity is required")
	}
	markerID, err := r.Authenticate(append([]byte("consumed-plan-v2\x00"), []byte(approvalID)...))
	if err != nil {
		return "", err
	}
	return filepath.Join(r.dir, "write-approval-consumed", markerID+".used"), nil
}
