package awscloudtrail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"

	"polymetrics/internal/connectors/connsdk"
)

// sigV4Signer applies an AWS Signature Version 4 Authorization header to each
// request. It is a focused, stdlib-only implementation of the subset of SigV4
// CloudTrail needs (a single POST with a known body). No third-party AWS SDK is
// pulled in. The secret access key is only ever fed into HMAC; it is never
// logged. It satisfies connsdk.Authenticator.
type sigV4Signer struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	service         string
	// now is injectable for deterministic tests; defaults to time.Now.
	now func() time.Time
}

// compile-time assertion that the signer is a connsdk.Authenticator.
var _ connsdk.Authenticator = (*sigV4Signer)(nil)

func (s *sigV4Signer) clock() time.Time {
	if s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

// Apply signs req in place. connsdk reads the already-encoded body off the
// request, so we must compute the payload hash from req.Body via GetBody, which
// connsdk-built requests always provide for bodies created from bytes.Reader.
func (s *sigV4Signer) Apply(_ context.Context, req *http.Request) error {
	now := s.clock()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	payloadHash := s.payloadHash(req)

	host := req.URL.Host
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	// Canonical headers: include host, content-type, and the x-amz-* set
	// CloudTrail expects, sorted by lowercase name.
	signedHeaderNames := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	headerValues := map[string]string{
		"host":                 host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	if ct := req.Header.Get("Content-Type"); ct != "" {
		signedHeaderNames = append(signedHeaderNames, "content-type")
		headerValues["content-type"] = ct
	}
	if target := req.Header.Get("X-Amz-Target"); target != "" {
		signedHeaderNames = append(signedHeaderNames, "x-amz-target")
		headerValues["x-amz-target"] = target
	}
	sort.Strings(signedHeaderNames)

	var canonicalHeaders strings.Builder
	for _, name := range signedHeaderNames {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(headerValues[name]))
		canonicalHeaders.WriteString("\n")
	}
	signedHeaders := strings.Join(signedHeaderNames, ";")

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQuery := req.URL.Query().Encode()

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := strings.Join([]string{dateStamp, s.region, s.service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hashHex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := s.signingKey(dateStamp)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authorization := "AWS4-HMAC-SHA256 " +
		"Credential=" + s.accessKeyID + "/" + credentialScope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature
	req.Header.Set("Authorization", authorization)
	return nil
}

// payloadHash returns the hex SHA256 of the request body, reading it via GetBody
// so the body remains available for the actual send.
func (s *sigV4Signer) payloadHash(req *http.Request) string {
	if req.GetBody == nil {
		return hashHex([]byte{})
	}
	body, err := req.GetBody()
	if err != nil {
		return hashHex([]byte{})
	}
	defer body.Close()
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 4096)
	for {
		n, readErr := body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if readErr != nil {
			break
		}
	}
	return hashHex(buf)
}

// signingKey derives the SigV4 signing key by chaining HMAC over the date,
// region, service, and the aws4_request terminator.
func (s *sigV4Signer) signingKey(dateStamp string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+s.secretAccessKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte(s.service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
