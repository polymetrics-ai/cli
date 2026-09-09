package connsdk

import "context"

// RequestAdmission is an in-memory, per-physical-request authorization check.
// It carries neither credential material nor a route: the caller already chose
// a declaration-owned request, and the check only decides whether its current
// epoch may reach the send boundary.
type RequestAdmission func(context.Context) error

type requestAdmissionContextKey struct{}

// WithRequestAdmission attaches a durable-health check to one operation
// context. Requester invokes it immediately before every physical HTTP send,
// including retries and redirects. Nil checks intentionally preserve callers
// that have no production coordination runtime.
func WithRequestAdmission(ctx context.Context, admission RequestAdmission) context.Context {
	if ctx == nil || admission == nil {
		return ctx
	}
	return context.WithValue(ctx, requestAdmissionContextKey{}, admission)
}

// CheckRequestAdmission revalidates the current operation epoch at a physical
// request boundary. Native connector adapters use the same helper before their
// database query or statement boundaries.
func CheckRequestAdmission(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}
	admission, _ := ctx.Value(requestAdmissionContextKey{}).(RequestAdmission)
	if admission == nil {
		return nil
	}
	return admission(ctx)
}
