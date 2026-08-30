package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

// EphemeralCredential describes one certification-only credential. Its secret
// values live only in a process-local CertificationEphemeralSession; callers
// must never serialize this value or pass it through command arguments.
type EphemeralCredential struct {
	Name      string
	Connector string
	Config    map[string]string
	Secrets   map[string]string
}

// CertificationEphemeralSession is a root-scoped, in-memory credential source
// for the certification runner. It deliberately has no persistence operations:
// sessions are registered before the runner opens its ephemeral project and
// are removed after the run completes.
type CertificationEphemeralSession struct {
	root     string
	approval *projectWriteApprovalAuthority

	mu          sync.RWMutex
	credentials map[string]EphemeralCredential
	closed      bool
}

var certificationEphemeralSessions = struct {
	sync.Mutex
	byRoot map[string]*CertificationEphemeralSession
}{byRoot: make(map[string]*CertificationEphemeralSession)}

// BeginCertificationEphemeralCredentials starts a process-local credential
// session for root. InitProject and Open recognize the session and must not
// create a vault/key or a persisted credential profile while it is active.
func BeginCertificationEphemeralCredentials(root string, credentials ...EphemeralCredential) (*CertificationEphemeralSession, error) {
	key, err := certificationEphemeralRoot(root)
	if err != nil {
		return nil, err
	}
	approval, err := newEphemeralProjectWriteApprovalAuthority(filepath.Join(key, ".polymetrics"))
	if err != nil {
		return nil, err
	}
	session := &CertificationEphemeralSession{
		root:        key,
		approval:    approval,
		credentials: make(map[string]EphemeralCredential, len(credentials)),
	}
	for _, credential := range credentials {
		if err := session.Put(credential); err != nil {
			return nil, err
		}
	}

	certificationEphemeralSessions.Lock()
	defer certificationEphemeralSessions.Unlock()
	if _, exists := certificationEphemeralSessions.byRoot[key]; exists {
		return nil, fmt.Errorf("certification ephemeral credentials already active for root %q", key)
	}
	certificationEphemeralSessions.byRoot[key] = session
	return session, nil
}

// Put adds or replaces a session credential without writing a profile or a
// secret file. It is intentionally available only to the certification package
// through this internal app package.
func (s *CertificationEphemeralSession) Put(credential EphemeralCredential) error {
	if err := safety.ValidateIdentifier(credential.Name, "ephemeral credential"); err != nil {
		return err
	}
	if err := safety.ValidateIdentifier(credential.Connector, "ephemeral connector"); err != nil {
		return err
	}
	if err := connectors.RejectLegacyConnectorName(credential.Connector); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("certification ephemeral credential session is closed")
	}
	s.credentials[credential.Name] = cloneEphemeralCredential(credential)
	return nil
}

// Close unregisters the session and clears its maps. It is idempotent so every
// certification exit path can defer it safely.
func (s *CertificationEphemeralSession) Close() {
	if s == nil {
		return
	}
	certificationEphemeralSessions.Lock()
	if certificationEphemeralSessions.byRoot[s.root] == s {
		delete(certificationEphemeralSessions.byRoot, s.root)
	}
	certificationEphemeralSessions.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	for name, credential := range s.credentials {
		for key := range credential.Secrets {
			credential.Secrets[key] = ""
		}
		delete(s.credentials, name)
	}
	if s.approval != nil {
		s.approval.key = [32]byte{}
		s.approval.signer = nil
		s.approval.dir = ""
		s.approval = nil
	}
	s.closed = true
}

func activeCertificationEphemeralSession(root string) *CertificationEphemeralSession {
	key, err := certificationEphemeralRoot(root)
	if err != nil {
		return nil
	}
	certificationEphemeralSessions.Lock()
	defer certificationEphemeralSessions.Unlock()
	return certificationEphemeralSessions.byRoot[key]
}

func certificationEphemeralRoot(root string) (string, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve certification ephemeral root: %w", err)
	}
	return filepath.Clean(abs), nil
}

func (s *CertificationEphemeralSession) credential(name string) (CredentialMeta, map[string]string, bool) {
	if s == nil {
		return CredentialMeta{}, nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	credential, ok := s.credentials[name]
	if !ok || s.closed {
		return CredentialMeta{}, nil, false
	}
	fields := make([]string, 0, len(credential.Secrets))
	for field := range credential.Secrets {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return CredentialMeta{
		ID:           "certification-" + credential.Name,
		Name:         credential.Name,
		Connector:    credential.Connector,
		Config:       cloneStringMap(credential.Config),
		SecretFields: fields,
	}, cloneStringMap(credential.Secrets), true
}

// publicCredential returns only immutable credential metadata for command
// preflight. Unlike credential, it never copies or reads the secret map.
func (s *CertificationEphemeralSession) publicCredential(name string) (CredentialMeta, bool) {
	if s == nil {
		return CredentialMeta{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	credential, ok := s.credentials[name]
	if !ok || s.closed {
		return CredentialMeta{}, false
	}
	return CredentialMeta{
		ID:        "certification-" + credential.Name,
		Name:      credential.Name,
		Connector: credential.Connector,
		Config:    cloneStringMap(credential.Config),
	}, true
}

// HasCredential reports whether name is currently registered without exposing
// its secret values. Certification stages use this as their observable proof
// that an intake was retained only by the process-local session.
func (s *CertificationEphemeralSession) HasCredential(name string) bool {
	_, _, ok := s.credential(name)
	return ok
}

func (s *CertificationEphemeralSession) writeApproval() *projectWriteApprovalAuthority {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil
	}
	return s.approval
}

func cloneEphemeralCredential(credential EphemeralCredential) EphemeralCredential {
	credential.Config = cloneStringMap(credential.Config)
	credential.Secrets = cloneStringMap(credential.Secrets)
	return credential
}

type ephemeralCertificationSecretStore struct{}

func (ephemeralCertificationSecretStore) PutSecret(context.Context, string, string) error {
	return errors.New("certification ephemeral credentials cannot persist rotated secrets")
}

var _ connectors.SecretStore = ephemeralCertificationSecretStore{}
