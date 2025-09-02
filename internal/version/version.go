package version

import "fmt"

// Version information is centralized here. In real releases this can be
// overridden at build time via -ldflags, e.g.:
//   go build -ldflags "-X codex-go/internal/version.Version=0.1.2 -X codex-go/internal/version.Commit=abc123 -X codex-go/internal/version.Date=2025-09-01"
// Keeping it simple makes it easier to learn and trace where the output comes from.
var Version = "0.1.0-dev"
var Commit = "unknown"
var Date = "" // optional build date string

// String returns a human-readable version line used by `codex version`.
func String() string {
    if Date != "" {
        return fmt.Sprintf("codex-go %s (%s, %s)", Version, Commit, Date)
    }
    return fmt.Sprintf("codex-go %s (%s)", Version, Commit)
}
