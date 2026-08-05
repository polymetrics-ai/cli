// Package shopify holds connector-local regression evidence for the Shopify
// definition bundle. The bundle remains declarative; this test proves the
// Admin API host constraint before the credential boundary consumes it.
package shopify

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(os.DirFS(".."), "shopify")
	if err != nil {
		t.Fatalf("load Shopify bundle: %v", err)
	}
	return bundle
}

func TestShopDomainUsesCanonicalAdminHost(t *testing.T) {
	connector := engine.New(loadBundle(t), nil)

	tests := []struct {
		name      string
		host      string
		wantError bool
	}{
		{
			name:      "rejects non myshopify host",
			host:      "fixture-shop.invalid.example",
			wantError: true,
		},
		{
			name: "accepts canonical myshopify host",
			host: "fixture-shop.myshopify.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := connectors.ValidateConfiguration(connector, map[string]string{"shop_domain": tt.host})
			if !tt.wantError {
				if err != nil {
					t.Fatalf("ValidateConfiguration(%q) error = %v", tt.host, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateConfiguration(%q) error = nil, want shop_domain pattern rejection", tt.host)
			}
			if !strings.Contains(err.Error(), "shop_domain") || !strings.Contains(err.Error(), "pattern") {
				t.Fatalf("ValidateConfiguration(%q) error = %q, want shop_domain and pattern", tt.host, err)
			}
			if strings.Contains(err.Error(), tt.host) {
				t.Fatalf("ValidateConfiguration(%q) error unexpectedly echoes host: %q", tt.host, err)
			}
		})
	}
}
