package driver

// SetLookInstalledBrowserForTest overrides the "installed browser"
// resolution step for the duration of a test, restoring the real
// launcher.LookPath afterwards. Needed because that step otherwise checks
// hardcoded per-OS install paths a test cannot control via $PATH, and
// whether the test machine has a real browser installed is not something a
// deterministic test should depend on.
func SetLookInstalledBrowserForTest(fn func() (string, bool)) (restore func()) {
	prev := lookInstalledBrowser
	lookInstalledBrowser = fn
	return func() { lookInstalledBrowser = prev }
}
