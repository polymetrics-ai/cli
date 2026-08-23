package engine

import (
	"errors"
	"net/http"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func providerResponseReceiptFromResponse(b Bundle, response *connsdk.Response, secrets map[string]string) *connectors.ProviderResponseReceipt {
	if response == nil {
		return nil
	}
	return providerResponseReceipt(b, response.Status, response.Header, response.Body, secrets)
}

func providerResponseReceiptFromHTTPError(b Bundle, err error, secrets map[string]string) *connectors.ProviderResponseReceipt {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) || httpErr == nil {
		return nil
	}
	raw := append([]byte(nil), httpErr.RawBody...)
	if raw == nil && httpErr.Body != "" {
		raw = []byte(httpErr.Body)
	}
	return providerResponseReceipt(b, httpErr.Status, httpErr.Header, raw, secrets)
}

func providerResponseReceipt(b Bundle, status int, headers http.Header, raw []byte, secrets map[string]string) *connectors.ProviderResponseReceipt {
	receipt := connectors.ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           status,
		Headers:          completeProviderResponseHeaders(b, headers),
		BodyPresent:      len(raw) != 0,
		BodyBytes:        int64(len(raw)),
	}
	if len(raw) != 0 {
		receipt.BodyRaw, receipt.BodyRawEncoding = writeProviderRawBody(raw)
		if decoded, err := decodeDirectReadBody(raw, -1); err == nil {
			receipt.Body = decoded
		} else {
			receipt.Body = receipt.BodyRaw
		}
	}
	// Engine direct-read results are also returned by native adapters, so this
	// boundary cannot rely on commandrunner being the sole public projection.
	// The sanitizer preserves provider receipt bytes and metadata for readback
	// binding; declaration-owned output-secret paths are applied separately.
	safe := connectors.SanitizeProviderResponseReceiptForOutput(receipt, secrets)
	return &safe
}

func completeProviderResponseHeaders(b Bundle, headers http.Header) map[string]connectors.OperationResponseHeader {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]connectors.OperationResponseHeader, len(headers))
	for name, values := range headers {
		// A response header name is provider-owned metadata, not a secret
		// classifier. Keep every name/value here; an equal configured credential
		// value remains ordinary provider output without an explicit declaration.
		out[name] = connectors.OperationResponseHeader{Values: append([]string(nil), values...)}
	}
	return out
}
