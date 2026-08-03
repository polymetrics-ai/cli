package vault

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// RotateKey generates a fresh 32-byte data key, re-encrypts every flat
// credential, namespaced credential, and blob currently in the vault under
// the new key, and persists it via the vault's own KeyProtector (F3' gap
// "no rotation, no attested delete" — a compromised key stops protecting
// anything after rotation, without forcing every credential to be
// re-authenticated).
//
// Every re-encrypted item is staged to a "<path>.rotate-tmp" file before
// anything is committed; if staging fails partway, nothing on disk changes
// and the old key stays authoritative. The new key is persisted once
// staging succeeds, then every staged file is renamed over its original.
// Known limitation, stated rather than hidden: the final rename loop is a
// sequence of per-file atomic renames, not one atomic transaction — a crash
// after the new key is persisted but before every rename completes can
// leave a mixed old/new-key state on disk. On a local single-user CLI vault
// this window is a bounded sequence of fast local renames, not a network
// operation, so the risk is low; it is not zero.
func (v *Vault) RotateKey(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	type item struct {
		path      string
		aad       []byte
		plaintext []byte
	}
	var items []item

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
		items = append(items, item{path: v.path(id), aad: []byte(id), plaintext: plaintext})
	}

	connectors, err := v.listConnectors()
	if err != nil {
		return err
	}
	for _, connector := range connectors {
		namespaces, err := v.List(ctx, connector)
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
				items = append(items, item{path: path, aad: ns.aad(), plaintext: plaintext})
			}
		}
	}

	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return fmt.Errorf("rotate key: generate new key: %w", err)
	}
	staging := &Vault{dir: v.dir, key: newKey}

	tmpPaths := make([]string, len(items))
	for i, it := range items {
		sealed, err := staging.seal(it.aad, it.plaintext)
		if err != nil {
			return fmt.Errorf("rotate key: encrypt %s: %w", it.path, err)
		}
		tmp := it.path + ".rotate-tmp"
		if err := os.WriteFile(tmp, sealed, 0o600); err != nil {
			return fmt.Errorf("rotate key: stage %s: %w", it.path, err)
		}
		tmpPaths[i] = tmp
	}

	storer, ok := v.protector.(keyStorer)
	if !ok {
		return fmt.Errorf("rotate key: protector %q does not support rotation", v.protector.Name())
	}
	if err := storer.StoreKey(v.dir, newKey); err != nil {
		return fmt.Errorf("rotate key: persist new key: %w", err)
	}

	for i, it := range items {
		if err := os.Rename(tmpPaths[i], it.path); err != nil {
			return fmt.Errorf("rotate key: commit %s: %w", it.path, err)
		}
	}

	v.key = newKey
	return nil
}
