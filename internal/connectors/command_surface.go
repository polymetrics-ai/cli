package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// RequestContractExecutionPolicyVersion identifies the provider-neutral PM
// resource envelope applied to generated request inputs.
const RequestContractExecutionPolicyVersion = "pm-request-contract-bounds-v1"

// ExactNumber preserves a declaration-owned JSON number as its source lexeme.
// CLI values and bounds must never pass through float64: provider IDs and
// decimal minima commonly exceed its exact range. It marshals as a JSON number
// (not a quoted string) so generated CLI/docs surfaces stay source-compatible.
type ExactNumber string

func (n ExactNumber) String() string { return string(n) }

func (n *ExactNumber) UnmarshalJSON(raw []byte) error {
	value, err := decodeExactNumber(raw)
	if err != nil {
		return err
	}
	*n = ExactNumber(value)
	return nil
}

func (n ExactNumber) MarshalJSON() ([]byte, error) {
	if _, err := decodeExactNumber([]byte(n)); err != nil {
		return nil, err
	}
	return []byte(n), nil
}

func decodeExactNumber(raw []byte) (json.Number, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("exact number: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return "", fmt.Errorf("exact number: must contain one value")
		}
		return "", fmt.Errorf("exact number: %w", err)
	}
	number, ok := value.(json.Number)
	if !ok {
		return "", fmt.Errorf("exact number: must be a JSON number")
	}
	return number, nil
}

// CommandSurface is docs/help metadata for a provider-style connector command
// tree. It is deliberately descriptive: execution is still controlled by the
// connector's streams, write actions, and future approved dispatch paths.
type CommandSurface struct {
	Tagline     string
	Usage       string
	SourceCLI   *CommandSurfaceSource
	Groups      []CommandSurfaceGroup
	GlobalFlags []CommandSurfaceFlag
	Commands    []CommandSurfaceCommand
	HelpTopics  []CommandSurfaceHelpTopic
}

type CommandSurfaceSource struct {
	Name      string
	Docs      string
	Reference string
	Source    string
}

type CommandSurfaceGroup struct {
	ID       string
	Title    string
	Commands []string
}

type CommandSurfaceFlag struct {
	Name       string
	Type       string
	Summary    string
	Values     []string
	MapsTo     string
	Format     string
	AllowEmpty *bool
	Minimum    *ExactNumber
	Maximum    *ExactNumber
	Required   bool
	Repeatable bool
	// EnvOnly requires the CLI to receive this command value through the
	// declared --from-env field=ENV channel. The resolved value is still
	// validated by the ordinary typed flag path, but it never appears in argv.
	EnvOnly bool
	// AllowBareString is valid only for a bounded, declaration-owned JSON
	// record field. It lets a source-declared string arm retain ordinary CLI
	// text syntax while the complete union remains schema-validated.
	AllowBareString bool
	// MaxItems/MinItems bound a string_array flag's item count. This is a second,
	// independent bound alongside the body schema's maxItems: the schema bound
	// only fires after the flag has been expanded into a body, whereas this one
	// can name the flag the user actually typed.
	MaxItems int
	MinItems int
	// MaxBytes bounds the exact percent-encoded value for path/query targets.
	// It is projected from the source operation parameter or from the engine's
	// conservative fallback and enforced again by the executor.
	MaxBytes int
	// MaxBytesOrigin and MaxBytesPolicyVersion identify MaxBytes as a PM
	// execution limit. They prevent generated help and inspection output from
	// misreporting an implementation resource ceiling as provider semantics.
	MaxBytesOrigin        string
	MaxBytesPolicyVersion string
}

// RequestExecutionLimit is the safe inspection projection of one generated
// command input's finite PM byte envelope.
type RequestExecutionLimit struct {
	Command       string `json:"command,omitempty"`
	Flag          string `json:"flag"`
	MapsTo        string `json:"maps_to,omitempty"`
	PolicyVersion string `json:"policy_version"`
	Origin        string `json:"origin"`
	Unit          string `json:"unit"`
	Effective     int    `json:"effective"`
}

type CommandSurfaceConstraint struct {
	Kind          string
	Fields        []string
	Left          string
	Right         string
	Op            string
	ValueType     string
	LeftFallback  string
	RightFallback string
	Message       string
}

type CommandSurfaceCommand struct {
	Path            string
	Summary         string
	Intent          string
	Availability    string
	Stream          string
	Write           string
	Operation       string
	SourceOperation string
	SourceCLIPath   string
	SourceURL       string
	Flags           []CommandSurfaceFlag
	Constraints     []CommandSurfaceConstraint
	Examples        []string
	APISurface      []CommandSurfaceEndpointRef
	OutputPolicy    string
	// RedactFields is retained for bundle compatibility. Commandrunner does not
	// use it to mutate connector-command records or errors, or forward it to
	// executor requests.
	//
	// It is also the third declaration site for a redact list, alongside a
	// write action's redact_fields and an operation's
	// sensitive_policy.redact_fields. Reverse-plan withholding deliberately
	// consults neither this field nor a merge of the three: two sources
	// feeding one path is how a withhold set silently resolves to the wrong
	// list. Wiring it in needs a deliberate decision, not a default.
	RedactFields []string
	Risk         string
	Approval     string
	Notes        string
}

type CommandSurfaceEndpointRef struct {
	Method string
	Path   string
}

type CommandSurfaceHelpTopic struct {
	Name    string
	Summary string
}

type CommandSurfaceProvider interface {
	CommandSurface() *CommandSurface
}

// RequestExecutionLimitsOf returns finite command input byte limits without
// reading connector credentials. The command surface has already computed the
// effective runtime cap, including the engine fallback for source-silent
// path/query parameters.
func RequestExecutionLimitsOf(c Connector) []RequestExecutionLimit {
	provider, ok := c.(CommandSurfaceProvider)
	if !ok {
		return nil
	}
	surface := provider.CommandSurface()
	if surface == nil {
		return nil
	}
	limits := make([]RequestExecutionLimit, 0)
	appendLimit := func(command string, flag CommandSurfaceFlag) {
		if flag.MaxBytes <= 0 || flag.MaxBytesOrigin == "" || flag.MaxBytesPolicyVersion == "" {
			return
		}
		limits = append(limits, RequestExecutionLimit{
			Command:       command,
			Flag:          flag.Name,
			MapsTo:        flag.MapsTo,
			PolicyVersion: flag.MaxBytesPolicyVersion,
			Origin:        flag.MaxBytesOrigin,
			Unit:          commandSurfaceMaxBytesUnit(flag),
			Effective:     flag.MaxBytes,
		})
	}
	for _, flag := range surface.GlobalFlags {
		appendLimit("", flag)
	}
	for _, command := range surface.Commands {
		for _, flag := range command.Flags {
			appendLimit(command.Path, flag)
		}
	}
	return limits
}

func commandSurfaceMaxBytesUnit(flag CommandSurfaceFlag) string {
	location, _, ok := strings.Cut(strings.TrimSpace(flag.MapsTo), ".")
	if ok && (location == "path" || location == "query") {
		return "encoded_bytes"
	}
	return "bytes"
}
