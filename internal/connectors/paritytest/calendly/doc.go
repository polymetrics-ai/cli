// Package calendlyparity is the engine-vs-legacy parity suite for the
// calendly bundle (wave1-pilot P-4). It lives in its own package (SPEC
// wave1-pilot §6) rather than internal/connectors/engine so that ten
// parallel pilot agents never collide on a shared Go test package/dir.
package calendlyparity
