package transportpolicy

import (
	"context"
	"errors"
	"net/http"
)

var ErrRedirectRefused = errors.New("redirect refused; preview and approve the redirected target")

type destructiveContextKey struct{}

func MarkDestructive(ctx context.Context) context.Context {
	return context.WithValue(ctx, destructiveContextKey{}, true)
}

func IsDestructive(ctx context.Context) bool {
	required, _ := ctx.Value(destructiveContextKey{}).(bool)
	return required
}

func HTTPClient(ctx context.Context, client *http.Client) *http.Client {
	return httpClient(ctx, client, ErrRedirectRefused)
}

func HTTPClientRetainingRedirectResponse(ctx context.Context, client *http.Client) *http.Client {
	return httpClient(ctx, client, http.ErrUseLastResponse)
}

func httpClient(ctx context.Context, client *http.Client, redirectError error) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	if !IsDestructive(ctx) {
		return client
	}
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error {
		return redirectError
	}
	return &clone
}
