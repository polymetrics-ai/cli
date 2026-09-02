package engine

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"
)

func TestCloudTrailSigV4AuthenticatorDeterministicVector(t *testing.T) {
	request, err := http.NewRequest(http.MethodPost, "https://cloudtrail.us-east-1.amazonaws.com/", bytes.NewReader([]byte(`{"MaxResults":1}`)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-amz-json-1.1")
	request.Header.Set("X-Amz-Target", "com.amazonaws.cloudtrail.v20131101.CloudTrail_20131101.LookupEvents")
	authenticator := &cloudTrailSigV4Authenticator{
		accessKeyID:     "AKIDEXAMPLE",
		secretAccessKey: "test-signing-secret",
		now:             func() time.Time { return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC) },
	}
	if err := authenticator.Apply(context.Background(), request); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	const wantAuthorization = "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260102/us-east-1/cloudtrail/aws4_request, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date;x-amz-target, Signature=d862f34a76b5d25334e1b5806bc65872cd8145c3e41d7418ac2810bcf557711a"
	if got := request.Header.Get("Authorization"); got != wantAuthorization {
		t.Fatalf("Authorization = %q", got)
	}
}
