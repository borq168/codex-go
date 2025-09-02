# codex-go (learning scaffold)

A minimal Go skeleton mirroring key responsibilities from codex-rs for learning.

## Layout
- cmd/codex: CLI entrypoint
- internal/server/mcp: Minimal stdio JSON handler (ping only)
- internal/version: Version info
- internal/protocol: Protocol types (placeholder)
- internal/exec: Execution interfaces (placeholder)

## Quick start
Prerequisite: Go 1.22+ installed and in PATH.
```
cd ~/Projects/0_Work/based/codex-go
# build
go build ./cmd/codex
# version
./codex version
# mcp ping (type a line then Enter)
printf '{"method":"ping"}\n' | ./codex mcp serve

# run (stream stdout/stderr and exit)
./codex run -- echo hello
```
