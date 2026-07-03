// Package nexusdatasets implements the nexus-datasets bundle's AuthHook
// (docs/migration/quarantine.json: AUTH_COMPLEX, "complex auth (hook
// needed)"): an HMAC-SHA256 request-signing connsdk.Authenticator, porting
// legacy internal/connectors/nexus-datasets/nexus_datasets.go's hmacAuth
// field-for-field (~90 lines, well under the Tier-2 soft target,
// conventions.md §1). Only one hook interface is implemented (AuthHook).
//
// The engine's declarative auth dialect has no mode that computes a
// signature over the outgoing request's own method/path/timestamp — this is
// exactly the AUTH_COMPLEX (signature auth) Tier-2 trigger named in
// conventions.md §1's Tier-2 table ("signature auth (SigV4, HMAC)"). Secret
// values (secret_key, api_key) flow ONLY into the outgoing request headers
// or the HMAC digest; they are never logged and never appear in an error
// string.
package nexusdatasets

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("nexus-datasets", func() engine.Hooks { return New() })
}

// Hooks is the nexus-datasets hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
}

// New returns a fresh nexus-datasets Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "nexus-datasets" }

// Authenticator resolves the HMAC-SHA256 connsdk.Authenticator for spec
// (mode "custom", hook "nexus-datasets"). Unlike a templated AuthSpec field
// (gmail's token_url/client_id shape), the four Infor Nexus credentials are
// read directly off cfg here by their fixed legacy key names — the same
// approach conventions.md documents for a custom hook that owns its own
// interpolation rather than routing through AuthSpec's generic fields, none
// of which map onto "access key id" / "HMAC secret key" naturally.
func (h *Hooks) Authenticator(_ context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	accessKeyID := strings.TrimSpace(cfg.Config["access_key_id"])
	if accessKeyID == "" {
		accessKeyID = strings.TrimSpace(cfg.Secrets["access_key_id"])
	}
	if accessKeyID == "" {
		return nil, errors.New("nexus-datasets oauth: access_key_id is required")
	}

	userID := strings.TrimSpace(cfg.Config["user_id"])
	if userID == "" {
		userID = strings.TrimSpace(cfg.Secrets["user_id"])
	}
	if userID == "" {
		return nil, errors.New("nexus-datasets oauth: user_id is required")
	}

	secretKey := strings.TrimSpace(cfg.Secrets["secret_key"])
	if secretKey == "" {
		secretKey = strings.TrimSpace(cfg.Config["secret_key"])
	}
	if secretKey == "" {
		return nil, errors.New("nexus-datasets oauth: secret_key is required")
	}

	apiKey := strings.TrimSpace(cfg.Secrets["api_key"])
	if apiKey == "" {
		apiKey = strings.TrimSpace(cfg.Config["api_key"])
	}
	if apiKey == "" {
		return nil, errors.New("nexus-datasets oauth: api_key is required")
	}

	return &hmacAuth{
		accessKeyID: accessKeyID,
		userID:      userID,
		secretKey:   secretKey,
		apiKey:      apiKey,
		now:         h.Now,
	}, nil
}

// hmacAuth implements connsdk.Authenticator for the Infor Nexus HMAC-SHA256
// signing scheme, mirroring legacy nexus_datasets.go's hmacAuth
// field-for-field: sign the canonical request (method, path, timestamp)
// keyed by the secret key, and set it plus the three Infor identity headers
// on every request. The exact upstream canonicalization may differ across
// Infor Nexus deployments (legacy's own caveat, carried forward verbatim);
// this implements the common HMAC-SHA256 scheme legacy already shipped. The
// secret key itself never appears on the wire (only its HMAC digest does)
// and is never logged.
type hmacAuth struct {
	accessKeyID string
	userID      string
	secretKey   string
	apiKey      string

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time
}

func (a *hmacAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

// Apply signs req and sets the Infor Nexus identity + signature headers.
// Signing happens fresh on every call (legacy computes a new timestamp per
// request, never caching the signature — matching nexus_datasets.go's Apply
// exactly).
func (a *hmacAuth) Apply(_ context.Context, req *http.Request) error {
	ts := strconv.FormatInt(a.timeNow().UTC().Unix(), 10)
	canonical := strings.Join([]string{req.Method, req.URL.Path, ts}, "\n")
	mac := hmac.New(sha256.New, []byte(a.secretKey))
	mac.Write([]byte(canonical))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Infor-AccessKeyId", a.accessKeyID)
	req.Header.Set("X-Infor-UserId", a.userID)
	req.Header.Set("X-Infor-ApiKey", a.apiKey)
	req.Header.Set("X-Infor-Timestamp", ts)
	req.Header.Set("Authorization", fmt.Sprintf("InforNexus %s:%s", a.accessKeyID, sig))
	return nil
}
