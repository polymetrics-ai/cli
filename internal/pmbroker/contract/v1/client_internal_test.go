package contractv1

import (
	"context"
	"strings"
	"testing"
)

func TestNewHTTPClientUsesDefaultTimeout(t *testing.T) {
	t.Parallel()

	client, err := NewHTTPClient("https://pm-broker.example", AuthenticatorFunc(func(context.Context) (Authorization, error) {
		return NewAuthorization("PMBroker", strings.Repeat("x", 24))
	}))
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}
	if client.httpClient.Timeout != defaultBrokerHTTPTimeout {
		t.Fatalf("http client timeout = %s, want %s", client.httpClient.Timeout, defaultBrokerHTTPTimeout)
	}
}
