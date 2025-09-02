package exec

import "context"

// Options controls how a command should be executed.
// We keep the shape intentionally small so it's easy to extend later
// (e.g., by adding resource limits, sandbox knobs, etc.).
type Options struct {
	// Cwd is the working directory for the process. Empty means inherit.
	Cwd string
	// Env is the environment as a list of KEY=VALUE entries. Empty means inherit.
	Env []string
	// TimeoutSec, if > 0, enforces a soft timeout for the process lifetime.
	TimeoutSec int
}

// EventType describes the kind of stream event emitted by a running process.
type EventType int

const (
	// EventStdout is a chunk of data read from stdout.
	EventStdout EventType = iota
	// EventStderr is a chunk of data read from stderr.
	EventStderr
	// EventExit indicates the process has terminated; Code holds the exit status.
	EventExit
)

// Event is a single item in the execution event stream.
// For stdout/stderr, Data contains a text chunk (not necessarily line-aligned).
// For exit, Code is populated.
type Event struct {
	Type EventType
	Data string
	Code int
}

// Runner abstracts process execution behind a streaming interface.
// Start should spawn the process and return a receive-only Event channel,
// a cancel func (to terminate the process), and an error if startup failed.
type Runner interface {
	Start(ctx context.Context, argv []string, opt Options) (<-chan Event, func() error, error)
}
