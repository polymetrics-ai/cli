package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Namespace identifies one credential/session/blob within the vault's
// per-connector/per-profile isolation scheme (F3' gap "flat ID namespace").
// Connector and Profile scope isolation; Kind distinguishes what is stored
// for that (connector, profile) pair — e.g. "oauth", "session",
// "risk_acceptance", or a native connector's own blob name (a whatsmeow
// device store, say).
type Namespace struct {
	Connector string
	Profile   string
	Kind      string
}

func (n Namespace) validate() error {
	if err := validateNamespaceComponent("connector", n.Connector); err != nil {
		return err
	}
	if err := ValidateProfile(n.Profile); err != nil {
		return err
	}
	return validateNamespaceComponent("kind", n.Kind)
}

// ValidateProfile applies the vault namespace profile grammar.
func ValidateProfile(profile string) error {
	return validateNamespaceComponent("profile", profile)
}

func validateNamespaceComponent(label, v string) error {
	if v == "" {
		return fmt.Errorf("namespace %s is required", label)
	}
	if strings.Contains(v, "..") || strings.ContainsAny(v, `/\`) {
		return fmt.Errorf("invalid namespace %s %q", label, v)
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return fmt.Errorf("invalid namespace %s %q", label, v)
		}
	}
	return nil
}

func (n Namespace) dir(root string) string {
	return filepath.Join(root, "ns", n.Connector, n.Profile)
}

// aad binds ciphertext to its exact namespace so a file cannot be moved to a
// different connector/profile/kind and still decrypt.
func (n Namespace) aad() []byte {
	return []byte("ns:" + n.Connector + "/" + n.Profile + "/" + n.Kind)
}

const (
	credentialExt = ".enc"
	blobExt       = ".blob"
)

// PutNamespaced stores a credential map under ns, isolated from every other
// connector/profile pair.
func (v *Vault) PutNamespaced(ctx context.Context, ns Namespace, secret map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ns.validate(); err != nil {
		return err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	plaintext, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("marshal secret bundle: %w", err)
	}
	return v.writeNamespaced(ns, credentialExt, plaintext)
}

// GetNamespaced returns the credential map stored under ns.
func (v *Vault) GetNamespaced(ctx context.Context, ns Namespace) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ns.validate(); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	plaintext, err := v.readNamespaced(ns, credentialExt)
	if err != nil {
		return nil, err
	}
	var out map[string]string
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, fmt.Errorf("decode secret bundle: %w", err)
	}
	if out == nil {
		out = map[string]string{}
	}
	return out, nil
}

// DeleteNamespaced removes the credential map stored under ns, if any.
func (v *Vault) DeleteNamespaced(ctx context.Context, ns Namespace) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ns.validate(); err != nil {
		return err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.deleteNamespacedFile(ns, credentialExt)
}

// PutBlob stores opaque bytes under ns — for structured session state a
// map[string]string cannot represent, such as a whatsmeow device store.
func (v *Vault) PutBlob(ctx context.Context, ns Namespace, blob []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ns.validate(); err != nil {
		return err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.writeNamespaced(ns, blobExt, blob)
}

// GetBlob returns the opaque bytes stored under ns.
func (v *Vault) GetBlob(ctx context.Context, ns Namespace) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ns.validate(); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.readNamespaced(ns, blobExt)
}

// DeleteBlob removes the opaque bytes stored under ns, if any.
func (v *Vault) DeleteBlob(ctx context.Context, ns Namespace) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ns.validate(); err != nil {
		return err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.deleteNamespacedFile(ns, blobExt)
}

// List returns every Namespace stored for connector, across all profiles and
// kinds, sorted by profile then kind. Each logical namespace appears exactly
// once no matter how many of its backing files exist — a (profile, kind) that
// holds both a credential and a blob is one Namespace, and callers such as
// RotateKey already probe both extensions per entry themselves.
func (v *Vault) List(ctx context.Context, connector string) ([]Namespace, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateNamespaceComponent("connector", connector); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.listNamespaces(connector)
}

func (v *Vault) listNamespaces(connector string) ([]Namespace, error) {
	root := filepath.Join(v.dir, "ns", connector)
	profiles, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list vault namespace %s: %w", connector, err)
	}

	var out []Namespace
	seen := make(map[Namespace]bool)
	for _, p := range profiles {
		if !p.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, p.Name()))
		if err != nil {
			return nil, fmt.Errorf("list vault namespace %s/%s: %w", connector, p.Name(), err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			var kind string
			switch {
			case strings.HasSuffix(name, credentialExt):
				kind = strings.TrimSuffix(name, credentialExt)
			case strings.HasSuffix(name, blobExt):
				kind = strings.TrimSuffix(name, blobExt)
			default:
				continue
			}
			ns := Namespace{Connector: connector, Profile: p.Name(), Kind: kind}
			if seen[ns] {
				continue
			}
			seen[ns] = true
			out = append(out, ns)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Profile != out[j].Profile {
			return out[i].Profile < out[j].Profile
		}
		return out[i].Kind < out[j].Kind
	})
	return out, nil
}

// DeleteAll removes every stored entry for (connector, profile) — the
// per-account isolation boundary's revocation primitive: logging out of one
// profile can never leave residue that a later List/Get for that profile
// would still see. Returns the number of entries removed.
func (v *Vault) DeleteAll(ctx context.Context, connector, profile string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if err := validateNamespaceComponent("connector", connector); err != nil {
		return 0, err
	}
	if err := ValidateProfile(profile); err != nil {
		return 0, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	dir := filepath.Join(v.dir, "ns", connector, profile)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("delete all %s/%s: %w", connector, profile, err)
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	if err := os.RemoveAll(dir); err != nil {
		return 0, fmt.Errorf("delete all %s/%s: %w", connector, profile, err)
	}
	return count, nil
}

func (v *Vault) writeNamespaced(ns Namespace, ext string, plaintext []byte) error {
	dir := ns.dir(v.dir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create namespace directory: %w", err)
	}
	sealed, err := v.seal(ns.aad(), plaintext)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, ns.Kind+ext)
	if err := os.WriteFile(path, sealed, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (v *Vault) readNamespaced(ns Namespace, ext string) ([]byte, error) {
	path := filepath.Join(ns.dir(v.dir), ns.Kind+ext)
	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	plaintext, err := v.open(ns.aad(), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt %s: %w", path, err)
	}
	return plaintext, nil
}

func (v *Vault) deleteNamespacedFile(ns Namespace, ext string) error {
	path := filepath.Join(ns.dir(v.dir), ns.Kind+ext)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (v *Vault) listConnectors() ([]string, error) {
	root := filepath.Join(v.dir, "ns")
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list vault namespaces: %w", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
