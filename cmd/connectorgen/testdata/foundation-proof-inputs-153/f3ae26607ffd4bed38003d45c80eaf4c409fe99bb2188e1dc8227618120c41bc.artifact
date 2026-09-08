package engine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
)

const (
	cloudTrailSigV4Service = "cloudtrail"
	cloudTrailSigV4Region  = "us-east-1"
)

type cloudTrailSigV4Authenticator struct {
	accessKeyID     string
	secretAccessKey string
	sessionToken    string
	now             func() time.Time
}

var _ connsdk.Authenticator = (*cloudTrailSigV4Authenticator)(nil)

func buildCloudTrailSigV4Authenticator(spec AuthSpec, vars Vars) (connsdk.Authenticator, error) {
	accessKeyID, err := interpolateCredential(spec.AccessKeyID, vars)
	if err != nil {
		return nil, fmt.Errorf("aws_sigv4: access_key_id: %w", err)
	}
	if err := credential.RequireAuthenticationValue("AWS access key ID", accessKeyID); err != nil {
		return nil, fmt.Errorf("aws_sigv4: %w", err)
	}
	secretAccessKey, err := interpolateCredential(spec.SecretAccessKey, vars)
	if err != nil {
		return nil, fmt.Errorf("aws_sigv4: secret_access_key: %w", err)
	}
	if err := credential.RequireAuthenticationValue("AWS secret access key", secretAccessKey); err != nil {
		return nil, fmt.Errorf("aws_sigv4: %w", err)
	}
	service, err := Interpolate(spec.AWSService, vars)
	if err != nil {
		return nil, fmt.Errorf("aws_sigv4: aws_service: %w", err)
	}
	region, err := Interpolate(spec.AWSRegion, vars)
	if err != nil {
		return nil, fmt.Errorf("aws_sigv4: aws_region: %w", err)
	}
	if service != cloudTrailSigV4Service || region != cloudTrailSigV4Region {
		return nil, fmt.Errorf("aws_sigv4: only fixed CloudTrail %s authentication is supported", cloudTrailSigV4Region)
	}

	sessionToken := ""
	if spec.SessionToken != "" {
		sessionToken, err = interpolateCredential(spec.SessionToken, vars)
		if err != nil {
			return nil, fmt.Errorf("aws_sigv4: session_token: %w", err)
		}
		if err := credential.RequireAuthenticationValue("AWS session token", sessionToken); err != nil {
			return nil, fmt.Errorf("aws_sigv4: %w", err)
		}
	}
	return &cloudTrailSigV4Authenticator{
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		sessionToken:    sessionToken,
	}, nil
}

func (a *cloudTrailSigV4Authenticator) Apply(_ context.Context, request *http.Request) error {
	if request.Method != http.MethodPost || request.URL.Scheme != "https" || request.URL.Host != "cloudtrail.us-east-1.amazonaws.com" || request.URL.Path != "/" || request.URL.RawQuery != "" {
		return fmt.Errorf("aws_sigv4: request must be the fixed CloudTrail HTTPS POST route")
	}
	now := time.Now().UTC()
	if a.now != nil {
		now = a.now().UTC()
	}
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := awsPayloadHash(request)
	host := request.URL.Host

	request.Header.Set("X-Amz-Date", amzDate)
	request.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if a.sessionToken != "" {
		request.Header.Set("X-Amz-Security-Token", a.sessionToken)
	}

	headerNames := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	headerValues := map[string]string{
		"host":                 host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	for _, name := range []string{"Content-Type", "X-Amz-Target", "X-Amz-Security-Token"} {
		if value := request.Header.Get(name); value != "" {
			lower := strings.ToLower(name)
			headerNames = append(headerNames, lower)
			headerValues[lower] = value
		}
	}
	sort.Strings(headerNames)
	var canonicalHeaders strings.Builder
	for _, name := range headerNames {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(headerValues[name]))
		canonicalHeaders.WriteString("\n")
	}

	canonicalURI := request.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	signedHeaders := strings.Join(headerNames, ";")
	canonicalRequest := strings.Join([]string{
		request.Method,
		canonicalURI,
		request.URL.Query().Encode(),
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")
	credentialScope := strings.Join([]string{dateStamp, cloudTrailSigV4Region, cloudTrailSigV4Service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		awsHashHex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := awsHMAC([]byte("AWS4"+a.secretAccessKey), []byte(dateStamp))
	signingKey = awsHMAC(signingKey, []byte(cloudTrailSigV4Region))
	signingKey = awsHMAC(signingKey, []byte(cloudTrailSigV4Service))
	signingKey = awsHMAC(signingKey, []byte("aws4_request"))
	signature := hex.EncodeToString(awsHMAC(signingKey, []byte(stringToSign)))
	request.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+a.accessKeyID+"/"+credentialScope+", SignedHeaders="+signedHeaders+", Signature="+signature)
	return nil
}

func awsPayloadHash(request *http.Request) string {
	if request.GetBody == nil {
		return awsHashHex(nil)
	}
	body, err := request.GetBody()
	if err != nil {
		return awsHashHex(nil)
	}
	defer func() { _ = body.Close() }()
	data := make([]byte, 0, 1024)
	buffer := make([]byte, 4096)
	for {
		n, readErr := body.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if readErr != nil {
			break
		}
	}
	return awsHashHex(data)
}

func awsHMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func awsHashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
