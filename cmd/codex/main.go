package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Internal packages for the learning scaffold. We keep them under
	// internal/ so the API surface can evolve freely without breaking users.
	iexec "codex-go/internal/exec"
	"codex-go/internal/server/mcp"
	"codex-go/internal/version"
)

// usage prints a minimal help message. We intentionally avoid pulling in
// external CLI deps (e.g., cobra) at this stage to keep setup friction low.
func usage() {
	fmt.Println("Usage:")
	fmt.Println("  codex [flags] version")
	fmt.Println("  codex [flags] mcp serve")
	fmt.Println("  codex [flags] run -- <cmd...>")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --cwd <dir>         Set working directory")
	fmt.Println("  --env <key=value>   Set environment variable (can be used multiple times)")
	fmt.Println("  --timeout <duration> Set timeout for command execution (e.g., 30s, 5m)")
}

// parseFlags parses global flags and returns remaining arguments
type GlobalFlags struct {
	cwd     string
	env     []string
	timeout time.Duration
}

func parseFlags(args []string) (GlobalFlags, []string, error) {
	var flags GlobalFlags
	var envFlags arrayFlags
	
	flagSet := flag.NewFlagSet("codex", flag.ContinueOnError)
	flagSet.StringVar(&flags.cwd, "cwd", "", "Set working directory")
	flagSet.Var(&envFlags, "env", "Set environment variable (key=value)")
	flagSet.DurationVar(&flags.timeout, "timeout", 0, "Set timeout for command execution")
	
	// Parse flags
	err := flagSet.Parse(args)
	if err != nil {
		return flags, nil, err
	}
	
	flags.env = envFlags
	return flags, flagSet.Args(), nil
}

// arrayFlags implements flag.Value for string slices
type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// applyGlobalFlags applies global flags to the environment
func applyGlobalFlags(flags GlobalFlags) error {
	// Change working directory
	if flags.cwd != "" {
		if err := os.Chdir(flags.cwd); err != nil {
			return fmt.Errorf("failed to change directory to %s: %v", flags.cwd, err)
		}
	}
	
	// Set environment variables
	for _, env := range flags.env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable format: %s (expected key=value)", env)
		}
		if err := os.Setenv(parts[0], parts[1]); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %v", parts[0], err)
		}
	}
	
	return nil
}

// main dispatches on the first CLI arg. The goal here is approachability:
// a few clear subcommands that we can evolve into a fuller CLI later.
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	// Parse global flags
	globalFlags, remainingArgs, err := parseFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "flag parsing error: %v\n", err)
		os.Exit(2)
	}
	
	if len(remainingArgs) == 0 {
		usage()
		os.Exit(2)
	}
	
	// Apply global flags
	if err := applyGlobalFlags(globalFlags); err != nil {
		fmt.Fprintf(os.Stderr, "flag application error: %v\n", err)
		os.Exit(1)
	}

	switch remainingArgs[0] {
	case "version":
		// Prints version string (optionally includes commit/date via -ldflags).
		fmt.Println(version.String())
	case "mcp":
		// Minimal stdio JSON loop. Initially only supports a ping method.
		if len(remainingArgs) >= 2 && remainingArgs[1] == "serve" {
			ctx := context.Background()
			// Apply timeout if specified
			if globalFlags.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, globalFlags.timeout)
				defer cancel()
			}
			if err := mcp.Serve(ctx, os.Stdin, os.Stdout); err != nil {
				// Errors go to stderr and a non‑zero exit code.
				fmt.Fprintf(os.Stderr, "mcp serve error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		fmt.Println("usage: codex mcp serve")
		os.Exit(2)
	case "run":
		// Minimal event-streaming runner: codex run -- <cmd...>
		// Example: codex run -- echo hello
		argv := remainingArgs[1:]
		if len(argv) > 0 && argv[0] == "--" {
			argv = argv[1:]
		}
		if len(argv) == 0 {
			fmt.Println("usage: codex run -- <cmd...>")
			os.Exit(2)
		}

		// Set up a context that cancels on Ctrl-C (SIGINT) or SIGTERM.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		
		// Apply timeout if specified
		if globalFlags.timeout > 0 {
			var timeoutCancel context.CancelFunc
			ctx, timeoutCancel = context.WithTimeout(ctx, globalFlags.timeout)
			defer timeoutCancel()
		}

		runner := iexec.NewLocalRunner()
		
		// Prepare options with environment variables
		opts := iexec.Options{}
		if len(globalFlags.env) > 0 {
			opts.Env = append(os.Environ(), globalFlags.env...)
		}
		
		events, cancel, err := runner.Start(ctx, argv, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run start error: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = cancel() }()

		// Stream events to the terminal.
		for ev := range events {
			switch ev.Type {
			case iexec.EventStdout:
				// Write stdout chunks as-is to stdout.
				fmt.Print(ev.Data)
			case iexec.EventStderr:
				// Write stderr chunks as-is to stderr.
				fmt.Fprint(os.Stderr, ev.Data)
			case iexec.EventExit:
				fmt.Fprintf(os.Stderr, "\n[exit %d]\n", ev.Code)
			}
		}
		os.Exit(0)
	default:
		usage()
		os.Exit(2)
	}
}
