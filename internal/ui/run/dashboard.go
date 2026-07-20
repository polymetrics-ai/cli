// Package run renders deterministic inline dashboards for long-running run commands.
package run

import (
	"context"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/ui/styles"
)

const (
	layoutWide     = "wide"
	layoutStandard = "standard"
	layoutCompact  = "compact"
	layoutGuard    = "guard"
)

// Step is the command-layer view data for one dashboard row.
type Step struct {
	ID     string
	Kind   string
	Detail string
}

// Config controls deterministic dashboard rendering.
type Config struct {
	Title         string
	Name          string
	Steps         []Step
	Width         int
	Height        int
	ASCII         bool
	NoColor       bool
	ReducedMotion bool
	Accessible    bool
	StartedAt     time.Time
	Now           func() time.Time
	Cancel        context.CancelFunc
	ResumeCommand string
}

// Model is a small state machine for inline run dashboards.
type Model struct {
	cfg       Config
	glyphs    styles.Glyphs
	steps     []stepState
	stepIndex map[string]int

	startedAt      time.Time
	finalKind      events.Kind
	finalStatus    string
	finalMessage   string
	cancelling     bool
	terminal       bool
	lastActiveStep string
	selected       int
	helpOpen       bool
	gPending       bool
}

type stepState struct {
	Step
	status   string
	message  string
	counters events.Counters
}

// NewModel returns a dashboard model with every configured step pending.
func NewModel(cfg Config) *Model {
	if cfg.Title == "" {
		cfg.Title = "Run"
	}
	if cfg.Width == 0 {
		cfg.Width = 80
	}
	if cfg.Height == 0 {
		cfg.Height = 24
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	started := cfg.StartedAt
	if started.IsZero() {
		started = cfg.Now()
	}
	m := &Model{
		cfg:       cfg,
		glyphs:    styles.ResolveGlyphs(cfg.ASCII),
		stepIndex: make(map[string]int, len(cfg.Steps)),
		startedAt: started,
	}
	for i, step := range cfg.Steps {
		state := stepState{Step: sanitizeStep(step), status: "pending"}
		m.steps = append(m.steps, state)
		m.stepIndex[state.ID] = i
	}
	return m
}

// Apply applies one lifecycle or progress event to the model.
func (m *Model) Apply(event events.Event) {
	if m == nil {
		return
	}
	if !event.Time.IsZero() && m.startedAt.IsZero() {
		m.startedAt = event.Time
	}
	stepID := clean(event.StepID)
	if stepID != "" {
		m.applyStepEvent(stepID, event)
		return
	}
	m.applyRunEvent(event)
}

// HandleKey handles simple dashboard key commands. It returns true only when
// the model can quit immediately.
func (m *Model) HandleKey(key string) bool {
	if m == nil {
		return true
	}
	raw := strings.TrimSpace(key)
	if raw == "G" || strings.EqualFold(raw, "end") {
		m.selectLast()
		m.gPending = false
		return false
	}
	switch strings.ToLower(raw) {
	case "ctrl+c", "ctrl-c", "^c":
		m.cancelling = true
		m.gPending = false
		if m.cfg.Cancel != nil {
			m.cfg.Cancel()
		}
		return false
	case "j", "down":
		m.selectBy(1)
	case "k", "up":
		m.selectBy(-1)
	case "ctrl+d", "pagedown", "page down":
		m.selectBy(max(1, len(m.steps)/2))
	case "ctrl+u", "pageup", "page up":
		m.selectBy(-max(1, len(m.steps)/2))
	case "home":
		m.selected = 0
	case "g":
		if m.gPending {
			m.selected = 0
		}
		m.gPending = !m.gPending
		return false
	case "?":
		m.helpOpen = !m.helpOpen
	case "esc", "escape":
		m.helpOpen = false
	case "q":
		return m.terminal && !m.helpOpen
	}
	m.gPending = false
	return false
}

// SelectedStep returns the currently focused step identifier.
func (m *Model) SelectedStep() string {
	if m == nil || len(m.steps) == 0 {
		return ""
	}
	if m.selected < 0 || m.selected >= len(m.steps) {
		return ""
	}
	return m.steps[m.selected].ID
}

// Resize updates the responsive layout dimensions.
func (m *Model) Resize(width, height int) {
	if m == nil {
		return
	}
	m.cfg.Width = width
	m.cfg.Height = height
}

// Done reports whether a terminal lifecycle event has arrived.
func (m *Model) Done() bool {
	return m == nil || m.terminal
}

// View renders the current dashboard frame.
func (m *Model) View() string {
	if m == nil {
		return ""
	}
	layout := m.layout()
	if layout == layoutGuard {
		return m.guardView()
	}
	if m.cfg.Accessible {
		return m.accessibleView()
	}

	var b strings.Builder
	header := fmt.Sprintf("%s %s", clean(m.cfg.Title), clean(m.cfg.Name))
	if layout == layoutCompact {
		header += "  compact"
	}
	elapsed := m.elapsed()
	fmt.Fprintf(&b, "%s%s elapsed %s", header, spacesBetween(header, 18), formatDuration(elapsed))
	if rate := m.recordsRate(elapsed); rate != "" {
		fmt.Fprintf(&b, " · %s records/s", rate)
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, "%s\n", rule(m.cfg.Width, m.cfg.ASCII))
	for i := range m.steps {
		m.writeStep(&b, i)
	}
	fmt.Fprintf(&b, "%s\n", rule(m.cfg.Width, m.cfg.ASCII))
	if final := m.finalLine(); final != "" {
		fmt.Fprintf(&b, "%s\n", final)
	}
	if m.cancelling && !m.terminal {
		fmt.Fprintf(&b, "%s\n", "– cancelling — waiting for checkpoints and final status")
	}
	if selected := m.SelectedStep(); selected != "" {
		fmt.Fprintf(&b, "Selected: %s\n", selected)
	}
	if m.helpOpen {
		b.WriteString("Keys: up/k previous · down/j next · home/gg first · end/G last · ctrl+u/ctrl+d page · esc close help\n")
	}
	fmt.Fprintf(&b, "NORMAL · run · ctrl+c cancel (checkpoints kept) · ? help\n")
	return b.String()
}

func (m *Model) applyStepEvent(stepID string, event events.Event) {
	idx, ok := m.stepIndex[stepID]
	if !ok {
		idx = len(m.steps)
		m.stepIndex[stepID] = idx
		m.steps = append(m.steps, stepState{Step: Step{ID: stepID}, status: "pending"})
	}
	step := &m.steps[idx]
	step.counters = mergeCounters(step.counters, event.Counters)
	step.message = cleanError(event.Message)
	step.status = eventStatus(event)
	if step.status == "running" || step.status == "ok" || step.status == "failed" || step.status == "cancelled" {
		m.lastActiveStep = step.ID
	}
}

func (m *Model) applyRunEvent(event events.Event) {
	if event.Kind == events.KindStarted {
		m.finalStatus = "running"
		return
	}
	if !event.Lifecycle() {
		return
	}
	m.finalKind = event.Kind
	m.finalStatus = eventStatus(event)
	m.finalMessage = cleanError(event.Message)
	switch event.Kind {
	case events.KindCompleted, events.KindFailed, events.KindSkipped:
		m.terminal = true
	}
}

func (m *Model) writeStep(b *strings.Builder, idx int) {
	step := m.steps[idx]
	glyph, word := m.statusGlyphWord(step.status)
	detail := step.Detail
	if detail == "" && step.Kind != "" {
		detail = step.Kind
	}
	if detail != "" {
		fmt.Fprintf(b, "%s %-18s %-8s %s\n", glyph, step.ID, step.Kind, clean(detail))
	} else {
		fmt.Fprintf(b, "%s %s\n", glyph, step.ID)
	}
	if idx < len(m.steps)-1 {
		fmt.Fprintf(b, "%s\n", m.glyphs.RailVertical)
	}
	metrics := formatCounters(step.counters)
	if metrics != "" || step.message != "" || word == "failed" {
		line := metrics
		if step.message != "" {
			if line != "" {
				line += "  "
			}
			line += fmt.Sprintf("%s — %s", word, step.message)
		}
		if line != "" {
			fmt.Fprintf(b, "  %s\n", line)
		}
	}
}

func (m *Model) statusGlyphWord(status string) (string, string) {
	switch status {
	case "ok", "success", string(events.KindCompleted):
		return m.glyphs.OK, "ok"
	case "failed":
		return m.glyphs.Failed, "failed"
	case "cancelled", "canceled":
		return m.glyphs.Skipped, "cancelled"
	case "running", string(events.KindStarted), string(events.KindProgress):
		return m.glyphs.Running, "running"
	case "skipped":
		return m.glyphs.Skipped, "skipped"
	case "partial":
		return m.glyphs.Partial, "partial"
	default:
		return m.glyphs.Pending, "pending"
	}
}

func (m *Model) finalLine() string {
	if !m.terminal {
		return ""
	}
	if m.finalStatus == "cancelled" || m.finalStatus == "canceled" || strings.Contains(strings.ToLower(m.finalMessage), context.Canceled.Error()) {
		step := m.lastActiveStep
		if step == "" && len(m.steps) > 0 {
			step = m.steps[0].ID
		}
		resume := m.cfg.ResumeCommand
		if resume == "" {
			resume = fmt.Sprintf("pm %s run %s", strings.ToLower(clean(m.cfg.Title)), clean(m.cfg.Name))
		}
		return fmt.Sprintf("%s Cancelled after %s. Resume: %s", m.glyphs.Skipped, step, resume)
	}
	glyph, word := m.statusGlyphWord(m.finalStatus)
	if word == "ok" {
		return fmt.Sprintf("%s %s %s finished — %d steps, %s records, %s", glyph, clean(m.cfg.Title), clean(m.cfg.Name), len(m.steps), formatInt(m.totalRecords()), formatDuration(m.elapsed()))
	}
	if m.finalMessage != "" {
		return fmt.Sprintf("%s %s %s failed — %s", glyph, clean(m.cfg.Title), clean(m.cfg.Name), m.finalMessage)
	}
	return fmt.Sprintf("%s %s %s %s", glyph, clean(m.cfg.Title), clean(m.cfg.Name), word)
}

func (m *Model) accessibleView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s elapsed %s\n", clean(m.cfg.Title), clean(m.cfg.Name), formatDuration(m.elapsed()))
	for _, step := range m.steps {
		_, word := m.statusGlyphWord(step.status)
		fmt.Fprintf(&b, "step %s %s", step.ID, word)
		if counters := formatCounters(step.counters); counters != "" {
			fmt.Fprintf(&b, " %s", counters)
		}
		if step.message != "" {
			fmt.Fprintf(&b, " %s", step.message)
		}
		b.WriteByte('\n')
	}
	if final := m.finalLine(); final != "" {
		fmt.Fprintf(&b, "%s\n", final)
	}
	b.WriteString("mode normal focus run\n")
	return b.String()
}

func (m *Model) guardView() string {
	return fmt.Sprintf("Terminal too small: %dx%d. Recommended: 80x24. Plain fallback: pm %s run --plain\n", m.cfg.Width, m.cfg.Height, strings.ToLower(clean(m.cfg.Title)))
}

func (m *Model) layout() string {
	if m.cfg.Width < 60 || m.cfg.Height < 18 {
		return layoutGuard
	}
	if m.cfg.Width >= 120 {
		return layoutWide
	}
	if m.cfg.Width >= 80 {
		return layoutStandard
	}
	return layoutCompact
}

func (m *Model) elapsed() time.Duration {
	start := m.startedAt
	if start.IsZero() {
		start = m.cfg.StartedAt
	}
	if start.IsZero() {
		return 0
	}
	d := m.cfg.Now().Sub(start)
	if d < 0 {
		return 0
	}
	return d
}

func (m *Model) selectBy(delta int) {
	if len(m.steps) == 0 {
		m.selected = 0
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.steps) {
		m.selected = len(m.steps) - 1
	}
}

func (m *Model) selectLast() {
	if len(m.steps) == 0 {
		m.selected = 0
		return
	}
	m.selected = len(m.steps) - 1
}

func (m *Model) totalRecords() int64 {
	var total int64
	for _, step := range m.steps {
		if step.counters.RecordsWritten > 0 {
			total += step.counters.RecordsWritten
			continue
		}
		total += step.counters.RecordsRead
	}
	return total
}

func (m *Model) recordsRate(elapsed time.Duration) string {
	records := m.totalRecords()
	if records == 0 || elapsed <= 0 {
		return ""
	}
	return fmt.Sprintf("%.0f", float64(records)/elapsed.Seconds())
}

// BridgeOptions configures dashboard event throttling.
type BridgeOptions struct {
	Interval time.Duration
	Sink     events.Emitter
}

// Bridge wraps the event throttle used by dashboard command wiring.
type Bridge struct {
	throttle *events.Throttle
}

// NewBridge returns a lifecycle-preserving throttled event bridge.
func NewBridge(opts BridgeOptions) *Bridge {
	return &Bridge{throttle: events.NewThrottle(opts.Interval, opts.Sink)}
}

// Emit implements events.Emitter.
func (b *Bridge) Emit(ctx context.Context, event events.Event) {
	if b == nil || b.throttle == nil {
		return
	}
	b.throttle.Emit(ctx, event)
}

// Flush forwards the latest coalesced progress event.
func (b *Bridge) Flush(ctx context.Context) {
	if b == nil || b.throttle == nil {
		return
	}
	b.throttle.Flush(ctx)
}

// Dropped reports coalesced progress events.
func (b *Bridge) Dropped() uint64 {
	if b == nil || b.throttle == nil {
		return 0
	}
	return b.throttle.Dropped()
}

func sanitizeStep(step Step) Step {
	step.ID = clean(step.ID)
	step.Kind = clean(step.Kind)
	step.Detail = clean(step.Detail)
	return step
}

func eventStatus(event events.Event) string {
	status := strings.ToLower(clean(event.Status))
	if status == "canceled" {
		status = "cancelled"
	}
	if status != "" {
		switch status {
		case "success", "completed":
			return "ok"
		default:
			return status
		}
	}
	switch event.Kind {
	case events.KindStarted, events.KindProgress:
		return "running"
	case events.KindCompleted:
		return "ok"
	case events.KindFailed:
		return "failed"
	case events.KindSkipped:
		return "skipped"
	default:
		return "pending"
	}
}

func mergeCounters(base, next events.Counters) events.Counters {
	if next.RecordsRead != 0 {
		base.RecordsRead = next.RecordsRead
	}
	if next.RecordsTransformed != 0 {
		base.RecordsTransformed = next.RecordsTransformed
	}
	if next.RecordsWritten != 0 {
		base.RecordsWritten = next.RecordsWritten
	}
	if next.RecordsFailed != 0 {
		base.RecordsFailed = next.RecordsFailed
	}
	if next.Batches != 0 {
		base.Batches = next.Batches
	}
	if next.Completed != 0 {
		base.Completed = next.Completed
	}
	if next.Total != 0 {
		base.Total = next.Total
	}
	return base
}

func clean(value string) string {
	return safety.SanitizeTerminalLine(value)
}

func cleanError(value string) string {
	return safety.RedactErrorText(safety.SanitizeTerminalLine(value))
}

func formatCounters(c events.Counters) string {
	parts := make([]string, 0, 4)
	if c.RecordsRead != 0 || c.RecordsWritten != 0 {
		parts = append(parts, fmt.Sprintf("%s read → %s written", formatInt(c.RecordsRead), formatInt(c.RecordsWritten)))
	}
	if c.RecordsFailed != 0 {
		parts = append(parts, fmt.Sprintf("%s failed", formatInt(c.RecordsFailed)))
	}
	if c.Batches != 0 {
		parts = append(parts, fmt.Sprintf("%s batches", formatInt(c.Batches)))
	}
	return strings.Join(parts, " · ")
}

func formatInt(n int64) string {
	negative := n < 0
	if negative {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if negative {
		return "-" + s
	}
	return s
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int(d.Round(time.Second).Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
	hours := minutes / 60
	minutes %= 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func rule(width int, ascii bool) string {
	if width <= 0 {
		width = 80
	}
	if width > 72 {
		width = 72
	}
	if ascii {
		return strings.Repeat("-", width)
	}
	return strings.Repeat("─", width)
}

func spacesBetween(left string, minimum int) string {
	count := minimum
	if len(left) < 42 {
		count += 42 - len(left)
	}
	return strings.Repeat(" ", count)
}

var _ events.Emitter = (*Bridge)(nil)
