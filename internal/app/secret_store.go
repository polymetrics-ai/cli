package app

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

// credentialSecretStore is the vault-backed connectors.SecretStore for one
// credential. It exists so a provider-rotated secret — an OAuth2 refresh token
// the provider replaces on every exchange — survives the process.
//
// It writes to the project's EXISTING encrypted local credential store
// (internal/vault: AES-256-GCM, per-project key, 0600 files under
// .polymetrics/vault) and to nothing else. No new storage path, no plaintext,
// zero centralized custody: the value never leaves the machine.
//
// It is scoped to a single credential id, so a connector can only ever rotate
// its own credential.
type credentialSecretStore struct {
	app *App
	id  string
}

// storeMu serialises rotation writes. A rotation can be driven from a
// connector's own goroutine while a sync is in flight, and App carries no lock
// of its own; without this, two credentials rotating at once would race on
// a.state and the state file. It is package-level rather than per-store because
// the contended resources — a.state and the state file — are per-App, not
// per-credential, and rotation is rare enough that a single lock costs nothing.
var storeMu sync.Mutex

// credentialSecretStore returns a store scoped to the given credential id.
func (a *App) credentialSecretStore(id string) connectors.SecretStore {
	return &credentialSecretStore{app: a, id: id}
}

// PutSecret merges one rotated value into the credential's encrypted secret
// bundle, leaving every sibling secret untouched.
//
// No error it returns carries the secret value.
func (s *credentialSecretStore) PutSecret(ctx context.Context, key, value string) error {
	if err := safety.ValidateIdentifier(key, "secret key"); err != nil {
		return fmt.Errorf("rotate credential secret: %w", err)
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	secrets, err := s.app.vault.Get(ctx, s.id)
	if err != nil {
		return fmt.Errorf("rotate credential secret %q: %w", key, err)
	}
	if secrets == nil {
		secrets = map[string]string{}
	}
	if existing, ok := secrets[key]; ok && existing == value {
		// Nothing rotated; skip the re-encrypt and the state write.
		return nil
	}
	secrets[key] = value
	if err := s.app.vault.Put(ctx, s.id, secrets); err != nil {
		return fmt.Errorf("rotate credential secret %q: %w", key, err)
	}
	return s.app.recordSecretField(s.id, key)
}

// recordSecretField keeps CredentialMeta.SecretFields — a list of secret KEY
// NAMES, never values — consistent after a rotation introduces a key the
// credential did not previously carry, and stamps UpdatedAt.
func (a *App) recordSecretField(id, key string) error {
	for i := range a.state.Credentials {
		if a.state.Credentials[i].ID != id {
			continue
		}
		fields := a.state.Credentials[i].SecretFields
		for _, existing := range fields {
			if existing == key {
				a.state.Credentials[i].UpdatedAt = time.Now().UTC()
				return a.save()
			}
		}
		fields = append(fields, key)
		sort.Strings(fields)
		a.state.Credentials[i].SecretFields = fields
		a.state.Credentials[i].UpdatedAt = time.Now().UTC()
		return a.save()
	}
	return fmt.Errorf("rotate credential secret: credential %q not found", id)
}
