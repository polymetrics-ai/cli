package connectors

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	coordinationIdentifierMaxLength = 128
	coordinationSaltMinLength       = 16
	rateScopeSubjectMaxLength       = 512
)

// CredentialBinding is the protected, non-secret metadata required to build a
// coordination identity. Its BindingID is never suitable for regular output,
// coordinator storage, or logs; callers receive only the derived projections.
//
// It intentionally has no secret or approval-revision field. Credential
// equality is an explicit project-state relationship, never an inference from
// credential material.
type CredentialBinding struct {
	BindingID      string
	ProviderFamily string
	AuthProfile    string
}

// RateScopeKind is the provider-declared subject class for a rate-limit
// policy. Add a kind only with public provider evidence; there is no implicit
// credential/token fallback.
type RateScopeKind string

const (
	RateScopeKindAccount          RateScopeKind = "account"
	RateScopeKindApplication      RateScopeKind = "application"
	RateScopeKindEndpointResource RateScopeKind = "endpoint_resource"
	RateScopeKindIP               RateScopeKind = "ip"
	RateScopeKindInstallation     RateScopeKind = "installation"
	RateScopeKindUser             RateScopeKind = "user"
	RateScopeKindWorkspace        RateScopeKind = "workspace"
)

// RateLimitScope is the explicitly declared, non-secret rate-budget subject
// supplied by a provider policy. It is intentionally separate from a
// credential binding because a provider budget can be broader or narrower than
// an authentication cohort.
type RateLimitScope struct {
	PolicyID string
	Kind     RateScopeKind
	Subject  string
}

// AuthCohortKey is an opaque projection for verified-authentication fencing.
// It is intentionally a distinct type from RateLimitScopeKey so a caller
// cannot accidentally use an authentication cohort as a rate-limit budget.
type AuthCohortKey string

// RateLimitScopeKey is an opaque projection for one explicit provider
// rate-limit policy scope. It is intentionally a distinct type from
// AuthCohortKey.
type RateLimitScopeKey string

// CoordinationIdentity contains only opaque coordination projections. Its
// fields remain private so JSON or fmt-based output cannot reveal a binding
// preimage or declared scope metadata by accident.
type CoordinationIdentity struct {
	authCohortKey AuthCohortKey
	rateScopeSeed string
}

// NewCoordinationIdentity builds the single identity plumbing shared by
// authentication fencing and rate-limit scopes. salt must be project-stable,
// randomly generated protected state; it is never derived from a credential
// secret or approval revision.
func NewCoordinationIdentity(salt []byte, binding CredentialBinding) (CoordinationIdentity, error) {
	if len(salt) < coordinationSaltMinLength {
		return CoordinationIdentity{}, errors.New("coordination salt is invalid")
	}
	if err := validateCredentialBinding(binding); err != nil {
		return CoordinationIdentity{}, err
	}

	return CoordinationIdentity{
		authCohortKey: AuthCohortKey(projection(salt, "auth-cohort-v1", binding.BindingID)),
		rateScopeSeed: projection(
			salt,
			"rate-scope-seed-v1",
			binding.BindingID,
			binding.ProviderFamily,
			binding.AuthProfile,
		),
	}, nil
}

// AuthCohortKey returns the opaque projection used by the verified-authentication
// fence. It is stable for the durable binding and intentionally independent of
// approval evidence and credential rotation.
func (i CoordinationIdentity) AuthCohortKey() AuthCohortKey {
	return i.authCohortKey
}

// RateScopeKey derives an opaque projection for one explicit provider policy
// scope. A caller cannot get a rate key without a policy ID, supported scope
// kind, and non-secret subject.
func (i CoordinationIdentity) RateScopeKey(scope RateLimitScope) (RateLimitScopeKey, error) {
	if i.authCohortKey == "" || i.rateScopeSeed == "" {
		return "", errors.New("coordination identity is unavailable")
	}
	if err := validateRateLimitScope(scope); err != nil {
		return "", err
	}
	return RateLimitScopeKey(projection(
		[]byte(i.rateScopeSeed),
		"rate-scope-v1",
		scope.PolicyID,
		string(scope.Kind),
		scope.Subject,
	)), nil
}

func validateCredentialBinding(binding CredentialBinding) error {
	if err := validateCoordinationIdentifier(binding.BindingID, "credential binding"); err != nil {
		return err
	}
	if err := validateCoordinationIdentifier(binding.ProviderFamily, "provider family"); err != nil {
		return err
	}
	return validateCoordinationIdentifier(binding.AuthProfile, "auth profile")
}

func validateRateLimitScope(scope RateLimitScope) error {
	if err := validateCoordinationIdentifier(scope.PolicyID, "rate policy ID"); err != nil {
		return err
	}
	if !supportedRateScopeKind(scope.Kind) {
		return errors.New("rate scope kind is unsupported")
	}
	if strings.TrimSpace(scope.Subject) == "" {
		return errors.New("rate scope subject is required")
	}
	if len(scope.Subject) > rateScopeSubjectMaxLength {
		return errors.New("rate scope subject violates maximum length constraint")
	}
	if !utf8.ValidString(scope.Subject) || containsUnsafeControlCharacter(scope.Subject) {
		return errors.New("rate scope subject violates character constraint")
	}
	return nil
}

func validateCoordinationIdentifier(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	if len(value) > coordinationIdentifierMaxLength {
		return errors.New(field + " violates maximum length constraint")
	}
	for _, r := range value {
		if !isCoordinationIdentifierRune(r) {
			return errors.New(field + " violates identifier constraint")
		}
	}
	if strings.Contains(value, "..") {
		return errors.New(field + " violates identifier constraint")
	}
	return nil
}

func isCoordinationIdentifierRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '-' || r == '_' || r == '.':
		return true
	default:
		return false
	}
}

func supportedRateScopeKind(kind RateScopeKind) bool {
	switch kind {
	case RateScopeKindAccount,
		RateScopeKindApplication,
		RateScopeKindEndpointResource,
		RateScopeKindIP,
		RateScopeKindInstallation,
		RateScopeKindUser,
		RateScopeKindWorkspace:
		return true
	default:
		return false
	}
}

func containsUnsafeControlCharacter(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return true
		}
	}
	return false
}

func projection(key []byte, domain string, parts ...string) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(domain))
	for _, part := range parts {
		_, _ = mac.Write([]byte{0})
		_, _ = mac.Write([]byte(part))
	}
	return hex.EncodeToString(mac.Sum(nil))
}
