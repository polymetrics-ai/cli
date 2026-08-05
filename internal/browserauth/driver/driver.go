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
	"net/url"
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

	// Timeout is the default readiness timeout used by Flow. A caller that
	// uses Session directly can still supply a more specific WaitFor timeout.
	Timeout time.Duration
}

func (cfg Config) validate() error {
	if len(cfg.RequiredCookies) == 0 {
		return errors.New("driver: required_cookies must name at least one cookie (never capture \"everything on the domain\")")
	}
	return nil
}

// FlowConfig binds the controlled browser session to the credential it may
// produce. It is deliberately static connector configuration, not a generic
// browser scripting surface: the user completes the provider's login in the
// visible page and Flow waits only for the declared minimum cookies.
type FlowConfig struct {
	Browser Config

	// Origin is the HTTPS provider origin that may consume the captured
	// session. It contains no path, query, fragment, or user info, so a later
	// native connector transport can pin requests to one origin.
	Origin string

	// CSRFHeader and CSRFValueCookie are either both empty or both set. When
	// set, Flow copies the value of that already-declared minimum cookie into
	// SessionCredential.CSRFValue for the connector's typed transport.
	CSRFHeader      string
	CSRFValueCookie string
}

// Flow is browserauth's browser-session credential outcome. It launches a
// real, user-visible browser through Config, waits for a user-completed login,
// captures only Config.RequiredCookies, and closes the controlled browser.
// It never inspects or enters a password.
type Flow struct {
	cfg    FlowConfig
	origin string
}

// NewFlow validates one browser-session credential flow. Browser availability
// is resolved only at Login time, so validating connector metadata does not
// spawn a browser or download Chromium.
func NewFlow(cfg FlowConfig) (*Flow, error) {
	if err := cfg.Browser.validate(); err != nil {
		return nil, err
	}
	origin, err := normalizeOrigin(cfg.Origin)
	if err != nil {
		return nil, err
	}
	if (cfg.CSRFHeader == "") != (cfg.CSRFValueCookie == "") {
		return nil, errors.New("driver: csrf_header and csrf_value_cookie must be set together")
	}
	if cfg.CSRFValueCookie != "" && !containsCookie(cfg.Browser.RequiredCookies, cfg.CSRFValueCookie) {
		return nil, fmt.Errorf("driver: csrf_value_cookie %q must be declared in required_cookies", cfg.CSRFValueCookie)
	}
	return &Flow{cfg: cfg, origin: origin}, nil
}

func normalizeOrigin(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("driver: origin must be an HTTPS origin without path, query, fragment, or user info: %q", raw)
	}
	return "https://" + u.Host, nil
}

func containsCookie(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
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

const defaultSessionWaitTimeout = 2 * time.Minute

// New resolves a browser per cfg, launches it (leakless disabled — see the
// package doc comment), opens one page, and returns a Session bound to it.
// The caller owns Close()ing the returned Session, which also terminates the
// launched browser process.
func New(ctx context.Context, cfg Config) (Session, error) {
	session, _, err := newSession(ctx, cfg)
	return session, err
}

func newSession(ctx context.Context, cfg Config) (Session, Resolution, error) {
	if err := cfg.validate(); err != nil {
		return nil, Resolution{}, err
	}
	res, err := Resolve(cfg)
	if err != nil {
		return nil, Resolution{}, err
	}

	l := launcher.New().
		Bin(res.Path).
		Headless(cfg.Headless).
		Leakless(false).
		Context(ctx)

	controlURL, err := l.Launch()
	if err != nil {
		return nil, Resolution{}, fmt.Errorf("driver: launch browser: %w", err)
	}

	browser := rod.New().Context(ctx).ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		l.Kill()
		return nil, Resolution{}, fmt.Errorf("driver: connect to browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		_ = browser.Close()
		l.Kill()
		return nil, Resolution{}, fmt.Errorf("driver: open page: %w", err)
	}

	if cfg.LoginURL != "" {
		if err := page.Context(ctx).Navigate(cfg.LoginURL); err != nil {
			_ = browser.Close()
			l.Kill()
			return nil, Resolution{}, fmt.Errorf("driver: navigate to login URL: %w", err)
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
	}, res, nil
}

func (f *Flow) Name() string { return "browser_session_capture" }

// Login launches a controlled real browser, lets the user authenticate in it,
// and returns the resulting browser session credential. The browser is always
// closed after capture or failure; subsequent connector execution reuses only
// the encrypted local credential, never the browser process.
func (f *Flow) Login(ctx context.Context) (browserauth.Credential, error) {
	session, resolution, err := newSession(ctx, f.cfg.Browser)
	if err != nil {
		return browserauth.Credential{}, err
	}
	return f.loginWithSession(ctx, session, resolution)
}

func (f *Flow) loginWithSession(ctx context.Context, session Session, resolution Resolution) (credential browserauth.Credential, err error) {
	defer func() {
		if closeErr := session.Close(); closeErr != nil && err == nil {
			credential = browserauth.Credential{}
			err = fmt.Errorf("driver: close browser after session capture: %w", closeErr)
		}
	}()

	timeout := f.cfg.Browser.Timeout
	if timeout <= 0 {
		timeout = defaultSessionWaitTimeout
	}
	if err := session.WaitFor(ctx, timeout, requiredCookiesReady(f.cfg.Browser.RequiredCookies)); err != nil {
		return browserauth.Credential{}, err
	}
	cookies, err := session.GetCookies(ctx)
	if err != nil {
		return browserauth.Credential{}, err
	}
	minimum, err := selectRequiredCookies(cookies, f.cfg.Browser.RequiredCookies)
	if err != nil {
		return browserauth.Credential{}, err
	}

	csrfValue := ""
	if f.cfg.CSRFValueCookie != "" {
		for _, cookie := range minimum {
			if cookie.Name == f.cfg.CSRFValueCookie {
				csrfValue = cookie.Value
				break
			}
		}
	}

	capturedAt := time.Now().UTC()
	return browserauth.Credential{Session: &browserauth.SessionCredential{
		Cookies:        minimum,
		CSRFHeader:     f.cfg.CSRFHeader,
		CSRFValue:      csrfValue,
		Origin:         f.origin,
		FingerprintRef: resolution.Version,
		CapturedAt:     capturedAt,
		ExpiresHint:    earliestCookieExpiry(minimum),
	}}, nil
}

func requiredCookiesReady(required []string) func(Snapshot) bool {
	return func(snapshot Snapshot) bool {
		for _, name := range required {
			if strings.TrimSpace(snapshot.Cookies[name]) == "" {
				return false
			}
		}
		return true
	}
}

func selectRequiredCookies(cookies []browserauth.Cookie, required []string) ([]browserauth.Cookie, error) {
	byName := make(map[string]browserauth.Cookie, len(cookies))
	for _, cookie := range cookies {
		byName[cookie.Name] = cookie
	}
	minimum := make([]browserauth.Cookie, 0, len(required))
	for _, name := range required {
		cookie, ok := byName[name]
		if !ok || strings.TrimSpace(cookie.Value) == "" {
			return nil, fmt.Errorf("driver: required session cookie %q was not captured", name)
		}
		minimum = append(minimum, cookie)
	}
	return minimum, nil
}

func earliestCookieExpiry(cookies []browserauth.Cookie) *time.Time {
	var earliest *time.Time
	for _, cookie := range cookies {
		if cookie.Expires == nil {
			continue
		}
		if earliest == nil || cookie.Expires.Before(*earliest) {
			expires := *cookie.Expires
			earliest = &expires
		}
	}
	return earliest
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
		timeout = defaultSessionWaitTimeout
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
