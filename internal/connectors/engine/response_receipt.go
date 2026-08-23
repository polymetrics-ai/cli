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
	// Receipts are internal immutable evidence. Configured and declared secret
	// masking happens only when commandrunner projects this value for public
	// output; mutating it here loses the proof needed by retries and App state.
	_ = secrets
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
	return &receipt
}

func completeProviderResponseHeaders(b Bundle, headers http.Header) map[string]connectors.OperationResponseHeader {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]connectors.OperationResponseHeader, len(headers))
	for name, values := range headers {
		out[name] = connectors.OperationResponseHeader{Values: append([]string(nil), values...)}
	}
	return out
}
