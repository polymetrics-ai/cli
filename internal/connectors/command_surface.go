package connectors

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
	Name    string
	Type    string
	Summary string
	Values  []string
	MapsTo  string
}

type CommandSurfaceCommand struct {
	Path          string
	Summary       string
	Intent        string
	Availability  string
	Stream        string
	Write         string
	SourceCLIPath string
	SourceURL     string
	Flags         []CommandSurfaceFlag
	Examples      []string
	Risk          string
	Approval      string
	Notes         string
}

type CommandSurfaceHelpTopic struct {
	Name    string
	Summary string
}

type CommandSurfaceProvider interface {
	CommandSurface() *CommandSurface
}
