package exec

import (
    "bufio"
    "context"
    "io"
    osexec "os/exec"
    "time"
)

// LocalRunner is a minimal Runner implementation backed by the standard
// library's os/exec. It streams stdout/stderr as chunks and emits a final
// EventExit when the command terminates.
type LocalRunner struct{}

// NewLocalRunner constructs a new LocalRunner.
func NewLocalRunner() *LocalRunner { return &LocalRunner{} }

// Start launches the process and returns an event stream and a cancel func.
//
// Behavior:
// - Spawns argv[0] with argv[1..] and the provided Cwd/Env.
// - Emits EventStdout/EventStderr with textual chunks (not necessarily lines).
// - Emits EventExit with the exit code when the process finishes.
// - cancel() attempts to terminate the process early.
func (r *LocalRunner) Start(parent context.Context, argv []string, opt Options) (<-chan Event, func() error, error) {
    if len(argv) == 0 {
        ch := make(chan Event)
        close(ch)
        return ch, func() error { return nil }, nil
    }

    // Honor timeout if provided.
    ctx := parent
    var cancelTimeout context.CancelFunc
    if opt.TimeoutSec > 0 {
        ctx, cancelTimeout = context.WithTimeout(parent, time.Duration(opt.TimeoutSec)*time.Second)
    }

    cmd := osexec.CommandContext(ctx, argv[0], argv[1:]...)
    if opt.Cwd != "" {
        cmd.Dir = opt.Cwd
    }
    if len(opt.Env) > 0 {
        cmd.Env = opt.Env
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        if cancelTimeout != nil {
            cancelTimeout()
        }
        return nil, nil, err
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        if cancelTimeout != nil {
            cancelTimeout()
        }
        return nil, nil, err
    }

    if err := cmd.Start(); err != nil {
        if cancelTimeout != nil {
            cancelTimeout()
        }
        return nil, nil, err
    }

    events := make(chan Event, 16)

    // Reader helper that streams chunks from r into events as type et.
    stream := func(r io.Reader, et EventType) {
        // Use a buffered reader to read fixed-size chunks; this avoids the
        // token-size limitation of bufio.Scanner and keeps implementation simple.
        br := bufio.NewReader(r)
        const chunk = 4096
        for {
            buf := make([]byte, chunk)
            n, err := br.Read(buf)
            if n > 0 {
                events <- Event{Type: et, Data: string(buf[:n])}
            }
            if err != nil {
                if err == io.EOF {
                    return
                }
                // On read error other than EOF, stop this stream.
                return
            }
        }
    }

    go stream(stdout, EventStdout)
    go stream(stderr, EventStderr)

    // Wait for process completion and emit exit code.
    go func() {
        // Wait respects context cancellation/timeout via CommandContext.
        err := cmd.Wait()
        code := 0
        if err != nil {
            // Best-effort extraction of exit status; if unavailable, leave 1.
            code = 1
            if exitErr, ok := err.(*osexec.ExitError); ok {
                // Process finished and produced non-zero exit code.
                if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok {
                    code = status.ExitStatus()
                }
            }
        }
        events <- Event{Type: EventExit, Code: code}
        close(events)
        if cancelTimeout != nil {
            cancelTimeout()
        }
    }()

    // Cancel function attempts to terminate the process by canceling the context.
    cancel := func() error {
        if cancelTimeout != nil {
            cancelTimeout()
        }
        // CommandContext cancellation sends SIGKILL on Unix-ish systems.
        return nil
    }

    return events, cancel, nil
}

