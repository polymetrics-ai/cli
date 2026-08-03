// Package driver drives a controlled browser to capture a login session for
// the -web connectors — the cookie/session mechanism half of browserauth
// (report §3.3-§3.4). It is never used to fetch data, only to obtain a
// credential once, interactively.
//
// Library choice: go-rod/rod, chosen over chromedp because rod's launcher
// both LOCATES an installed Chrome/Chromium/Edge and DOWNLOADS a pinned
// Chromium revision when none is present — the cookie path needs that
// anyway, since captured cookies and the fingerprint descriptor must come
// from the same browser build. leakless (rod's platform-specific supervisor
// binary) is explicitly disabled: it would drop an extra binary onto disk,
// an unacceptable supply-chain addition for a CLI whose own AGENTS.md calls
// it "dependency-free ETL"; explicit context cancellation (ctx passed to
// every call, Close() killing the launched process) replaces it.
// playwright-go was disqualified outright — it downloads and drives a
// Node.js driver, which the Go-only-CLI rule forbids outright.
//
// The password rule is enforced mechanically, not by policy: Session's
// entire surface is Navigate, WaitFor, GetCookies, and Close. There is no
// Input, Keyboard, SetFiles, or Type method anywhere in this package.
// guard_test.go greps this package's source for those symbols and fails
// the build if any appear, so the boundary survives a careless future edit
// — the tool cannot type a password into a page because the code to do so
// does not exist.
package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"polymetrics.ai/internal/browserauth"
)

// Resolution describes the browser build a Session is bound to. It is
// recorded as SessionCredential.FingerprintRef because captured cookies are
// only coherent with the browser build that produced them.
type Resolution struct {
	// Source is exactly one of "user_specified", "installed", or
	// "pinned_download" — which of the three resolution steps produced Path.
	Source string
	Path   string
	// Version identifies the exact browser build, and is never empty for a
	// successful Resolve. For a pinned download it is the pinned revision.
	// Otherwise it is the binary's own `--version` output when the binary
	// reports one, and failing that a stat fingerprint of the binary
	// ("<abs path>@size:<n>,mtime:<unixnano>") — which still changes
	// whenever the browser is upgraded in place, the coherence signal
	// callers actually need.
	Version string
}

// Config configures one browser-driven session capture.
type Config struct {
	// BrowserPath, when set, is used unconditionally (resolution step 1:
	// a user-specified binary).
	BrowserPath string

	// AllowDownload opts into resolution step 3: downloading a pinned
	// Chromium revision into rod's local cache when no browser is found at
	// BrowserPath or already installed on the host (step 2). This never
	// happens without explicit opt-in — no silent download.
	AllowDownload bool

	// LoginURL is the real provider login page to navigate to. New opens it
	// as soon as the page exists; leave it empty to have New leave the page
	// blank and navigate later via Session.Navigate. It is never pre-filled
	// with any credential — the user completes the login themselves in the
	// visible page.
	LoginURL string

	// RequiredCookies is the named minimum this connector's mechanism
	// needs. GetCookies returns only cookies whose name appears here —
	// never "every cookie on the domain".
	RequiredCookies []string

	// Headless controls whether the launched browser is headless. Default
	// false (headed): a captured session needs the user to see and
	// interact with a real login page. Tests that drive a fake local login
	// page may set this true.
	Headless bool

	// Timeout bounds Session.Launch and is the default for WaitFor when its
	// own timeout argument is <= 0.
	Timeout time.Duration
}

func (cfg Config) validate() error {
	if len(cfg.RequiredCookies) == 0 {
		return errors.New("driver: required_cookies must name at least one cookie (never capture \"everything on the domain\")")
	}
	return nil
}

// lookInstalledBrowser is a package variable (not a direct launcher.LookPath
// call) so tests can override the "installed browser" resolution step —
// launcher.LookPath checks hardcoded per-OS install paths that a test
// cannot control via $PATH alone, and the machine running the test may or
// may not actually have a browser installed.
var lookInstalledBrowser = launcher.LookPath

// Resolve implements the three-step browser resolution order documented on
// Config.AllowDownload. It never starts a browser session, but it is not a
// pure query: on the user_specified and installed branches it briefly
// executes the resolved binary as `<path> --version` (bounded by
// browserVersionProbeTimeout, with captured output capped) to fill
// Resolution.Version. Callers that must not spawn a process from a
// config-supplied path should not call Resolve.
func Resolve(cfg Config) (Resolution, error) {
	if cfg.BrowserPath != "" {
		if _, err := os.Stat(cfg.BrowserPath); err != nil {
			return Resolution{}, fmt.Errorf("driver: configured browser path %q: %w", cfg.BrowserPath, err)
		}
		return Resolution{Source: "user_specified", Path: cfg.BrowserPath, Version: browserBuildID(cfg.BrowserPath)}, nil
	}
	if found, ok := lookInstalledBrowser(); ok {
		return Resolution{Source: "installed", Path: found, Version: browserBuildID(found)}, nil
	}
	if !cfg.AllowDownload {
		return Resolution{}, errors.New("driver: no installed Chrome/Chromium/Edge found; set Config.AllowDownload to fetch a pinned Chromium revision (this package never downloads a browser silently)")
	}
	b := launcher.NewBrowser()
	path, err := b.Get()
	if err != nil {
		return Resolution{}, fmt.Errorf("driver: download pinned Chromium revision %d: %w", b.Revision, err)
	}
	return Resolution{Source: "pinned_download", Path: path, Version: fmt.Sprintf("chromium-%d", b.Revision)}, nil
}

const (
	// browserVersionProbeTimeout bounds the `<binary> --version` probe so a
	// wedged or non-responsive binary cannot stall resolution.
	browserVersionProbeTimeout = 5 * time.Second
	// maxBrowserVersionOutput caps what the probe keeps of the child's
	// stdout. A version line is tens of bytes; a binary that floods stdout
	// must not make Resolve allocate without bound.
	maxBrowserVersionOutput = 4096
)

// browserBuildID identifies the exact build at path. Chrome-family binaries
// print a parseable version for `--version` on Unix; where that is
// unavailable or empty — Windows, plus wrapper scripts and stripped builds —
// identity falls back to the binary's size and modification time, which
// still change whenever the browser is upgraded in place.
func browserBuildID(path string) string {
	if v := probeBrowserVersion(path); v != "" {
		return v
	}
	return binaryStatFingerprint(path)
}

// probeBrowserVersion returns the binary's self-reported version, or "" when
// it reports none. It is skipped outright on Windows, where Chrome-family
// binaries are GUI-subsystem executables that answer --version with no
// stdout at all, so every call would pay for a process spawn that cannot
// succeed.
func probeBrowserVersion(path string) string {
	if runtime.GOOS == "windows" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), browserVersionProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	out := &cappedBuffer{limit: maxBrowserVersionOutput}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// binaryStatFingerprint identifies a binary by absolute path, size, and
// modification time. It deliberately does not hash the file's contents:
// Resolve runs on every session capture and a browser binary is large
// enough that reading all of it merely to notice an upgrade is not worth
// the cost.
func binaryStatFingerprint(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	info, err := os.Stat(path)
	if err != nil {
		return abs
	}
	return fmt.Sprintf("%s@size:%d,mtime:%d", abs, info.Size(), info.ModTime().UnixNano())
}

// cappedBuffer keeps at most limit bytes and silently discards the rest,
// while still reporting every byte as written so the child process never
// sees a short write.
type cappedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if room := c.limit - c.buf.Len(); room > 0 {
		if len(p) > room {
			p = p[:room]
		}
		c.buf.Write(p)
	}
	return n, nil
}

func (c *cappedBuffer) String() string { return c.buf.String() }

// Snapshot is a point-in-time read of the driven page, passed to a WaitFor
// readiness predicate. Cookies contains only Config.RequiredCookies —
// consistent with GetCookies never returning the full jar.
type Snapshot struct {
	URL     string
	Cookies map[string]string
}

// Session is the ONLY surface this package exposes for driving a browser.
type Session interface {
	Navigate(ctx context.Context, url string) error
	WaitFor(ctx context.Context, timeout time.Duration, ready func(Snapshot) bool) error
	GetCookies(ctx context.Context) ([]browserauth.Cookie, error)
	Close() error
}

// New resolves a browser per cfg, launches it (leakless disabled — see the
// package doc comment), opens one page, and returns a Session bound to it.
// The caller owns Close()ing the returned Session, which also terminates the
// launched browser process.
func New(ctx context.Context, cfg Config) (Session, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	res, err := Resolve(cfg)
	if err != nil {
		return nil, err
	}

	l := launcher.New().
		Bin(res.Path).
		Headless(cfg.Headless).
		Leakless(false).
		Context(ctx)

	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("driver: launch browser: %w", err)
	}

	browser := rod.New().Context(ctx).ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		l.Kill()
		return nil, fmt.Errorf("driver: connect to browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		_ = browser.Close()
		l.Kill()
		return nil, fmt.Errorf("driver: open page: %w", err)
	}

	if cfg.LoginURL != "" {
		if err := page.Context(ctx).Navigate(cfg.LoginURL); err != nil {
			_ = browser.Close()
			l.Kill()
			return nil, fmt.Errorf("driver: navigate to login URL: %w", err)
		}
	}

	names := make(map[string]bool, len(cfg.RequiredCookies))
	for _, n := range cfg.RequiredCookies {
		names[n] = true
	}

	return &rodSession{
		launcher:   l,
		browser:    browser,
		page:       page,
		names:      names,
		resolution: res,
	}, nil
}

type rodSession struct {
	launcher   *launcher.Launcher
	browser    *rod.Browser
	page       *rod.Page
	names      map[string]bool
	resolution Resolution

	closeOnce sync.Once
}

func (s *rodSession) Navigate(ctx context.Context, url string) error {
	if err := s.page.Context(ctx).Navigate(url); err != nil {
		return fmt.Errorf("driver: navigate: %w", err)
	}
	return nil
}

func (s *rodSession) WaitFor(ctx context.Context, timeout time.Duration, ready func(Snapshot) bool) error {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		snap, err := s.snapshot(ctx)
		if err == nil && ready(snap) {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("driver: timed out waiting for session readiness")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (s *rodSession) snapshot(ctx context.Context) (Snapshot, error) {
	info, err := s.page.Context(ctx).Info()
	if err != nil {
		return Snapshot{}, fmt.Errorf("driver: page info: %w", err)
	}
	cookies, err := s.GetCookies(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	m := make(map[string]string, len(cookies))
	for _, c := range cookies {
		m[c.Name] = c.Value
	}
	return Snapshot{URL: info.URL, Cookies: m}, nil
}

// GetCookies returns only cookies whose name is in Config.RequiredCookies —
// the named minimum, never the full jar (report §3.4).
func (s *rodSession) GetCookies(ctx context.Context) ([]browserauth.Cookie, error) {
	raw, err := s.page.Context(ctx).Cookies(nil)
	if err != nil {
		return nil, fmt.Errorf("driver: get cookies: %w", err)
	}
	out := make([]browserauth.Cookie, 0, len(s.names))
	for _, c := range raw {
		if !s.names[c.Name] {
			continue
		}
		cookie := browserauth.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
		}
		if c.Expires > 0 {
			t := c.Expires.Time()
			cookie.Expires = &t
		}
		out = append(out, cookie)
	}
	return out, nil
}

func (s *rodSession) Close() error {
	s.closeOnce.Do(func() {
		if s.browser != nil {
			_ = s.browser.Close()
		}
		if s.launcher != nil {
			s.launcher.Kill()
		}
	})
	return nil
}
