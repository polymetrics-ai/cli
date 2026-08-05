package vault

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var rotateRename = os.Rename

type rotationItem struct {
	path       string
	aad        []byte
	plaintext  []byte
	ciphertext []byte
	temporary  string
}

// RotateKey generates a fresh 32-byte data key, re-encrypts every flat
// credential, namespaced credential, and blob currently in the vault under
// the new key, and persists it via the vault's own KeyProtector (F3' gap
// "no rotation, no attested delete" — a compromised key stops protecting
// anything after rotation, without forcing every credential to be
// re-authenticated).
//
// Every re-encrypted item is staged to a "<path>.rotate-tmp" file before
// anything is committed; if staging fails partway, nothing on disk changes
// and the old key stays authoritative.
func (v *Vault) RotateKey(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()

	var items []rotationItem

	flatIDs, err := v.listFlatIDs()
	if err != nil {
		return err
	}
	for _, id := range flatIDs {
		ciphertext, err := os.ReadFile(v.path(id))
		if err != nil {
			return fmt.Errorf("rotate key: read %s: %w", id, err)
		}
		plaintext, err := v.open([]byte(id), ciphertext)
		if err != nil {
			return fmt.Errorf("rotate key: decrypt %s: %w", id, err)
		}
		items = append(items, rotationItem{path: v.path(id), aad: []byte(id), plaintext: plaintext, ciphertext: ciphertext})
	}

	connectors, err := v.listConnectors()
	if err != nil {
		return err
	}
	for _, connector := range connectors {
		if err := ctx.Err(); err != nil {
			return err
		}
		namespaces, err := v.listNamespaces(connector)
		if err != nil {
			return err
		}
		for _, ns := range namespaces {
			for _, ext := range []string{credentialExt, blobExt} {
				path := filepath.Join(ns.dir(v.dir), ns.Kind+ext)
				ciphertext, err := os.ReadFile(path)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return fmt.Errorf("rotate key: read %s: %w", path, err)
				}
				plaintext, err := v.open(ns.aad(), ciphertext)
				if err != nil {
					return fmt.Errorf("rotate key: decrypt %s: %w", path, err)
				}
				items = append(items, rotationItem{path: path, aad: ns.aad(), plaintext: plaintext, ciphertext: ciphertext})
			}
		}
	}

	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return fmt.Errorf("rotate key: generate new key: %w", err)
	}
	staging := &Vault{dir: v.dir, key: newKey}

	defer func() {
		for _, it := range items {
			if it.temporary != "" {
				_ = os.Remove(it.temporary)
			}
		}
	}()
	for i, it := range items {
		sealed, err := staging.seal(it.aad, it.plaintext)
		if err != nil {
			return fmt.Errorf("rotate key: encrypt %s: %w", it.path, err)
		}
		items[i].temporary = it.path + ".rotate-tmp"
		if err := os.WriteFile(items[i].temporary, sealed, 0o600); err != nil {
			return fmt.Errorf("rotate key: stage %s: %w", it.path, err)
		}
	}

	storer, ok := v.protector.(keyStorer)
	if !ok {
		return errors.New("rotate key: configured protector does not support rotation")
	}

	for i, it := range items {
		if err := rotateRename(it.temporary, it.path); err != nil {
			if rollbackErr := restoreRotatedFiles(items[:i]); rollbackErr != nil {
				return errors.Join(fmt.Errorf("rotate key: commit %s: %w", it.path, err), rollbackErr)
			}
			return fmt.Errorf("rotate key: commit %s: %w", it.path, err)
		}
	}

	if err := storer.StoreKey(v.dir, newKey); err != nil {
		if rollbackErr := restoreRotatedFiles(items); rollbackErr != nil {
			return errors.Join(fmt.Errorf("rotate key: persist new key: %w", err), rollbackErr)
		}
		return fmt.Errorf("rotate key: persist new key: %w", err)
	}

	v.key = newKey
	return nil
}

func restoreRotatedFiles(items []rotationItem) error {
	var errs []error
	for _, it := range items {
		if err := os.WriteFile(it.path, it.ciphertext, 0o600); err != nil {
			errs = append(errs, fmt.Errorf("rotate key: restore %s: %w", it.path, err))
		}
	}
	return errors.Join(errs...)
}
