package pinterest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// decodeTokenResponse reads the access token and expiry from an OAuth token
// response. expires_in is decoded with json.Number so it tolerates either a
// JSON number or string. A missing or non-positive expiry falls back to one
// hour, matching Pinterest's default access-token lifetime.
func decodeTokenResponse(resp *http.Response) (string, time.Duration, error) {
	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", 0, fmt.Errorf("pinterest oauth: decode token response: %w", err)
	}
	if out.AccessToken == "" {
		return "", 0, errors.New("pinterest oauth: token response missing access_token")
	}
	ttl := time.Hour
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	return out.AccessToken, ttl, nil
}

// parsePositiveInt parses a strictly-positive integer.
func parsePositiveInt(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 1 {
		return 0, errors.New("must be >= 1")
	}
	return value, nil
}
