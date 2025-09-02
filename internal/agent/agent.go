package agent

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "strings"

    "codex-go/internal/protocol"
)

// writeJSONLine marshals v to JSON and writes a newline-terminated frame.
func writeJSONLine(w io.Writer, v any) error {
    b, err := json.Marshal(v)
    if err != nil {
        return err
    }
    if _, err := w.Write(append(b, '\n')); err != nil {
        return err
    }
    return nil
}

// textFromUserInput extracts concatenated text items from a user_input op.
func textFromUserInput(op protocol.Op) string {
    var parts []string
    for _, it := range op.Items {
        if strings.ToLower(it.Type) == "text" && it.Text != "" {
            parts = append(parts, it.Text)
        }
    }
    return strings.TrimSpace(strings.Join(parts, " "))
}

// Serve implements the Phase 1 minimal protocol loop over a line-delimited
// JSON stream. For each Submission:
// - user_input => task_started, agent_message, task_complete
// - interrupt  => error("interrupted")
func Serve(ctx context.Context, r io.Reader, w io.Writer) error {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        line := scanner.Bytes()

        var sub protocol.Submission
        if err := json.Unmarshal(line, &sub); err != nil {
            // For invalid JSON, emit a protocol-level error without id binding.
            // Keep the loop alive for subsequent frames.
            _ = writeJSONLine(w, map[string]string{"error": "invalid json"})
            continue
        }

        switch sub.Op.Type {
        case protocol.OpUserInput:
            // 1. task_started
            _ = writeJSONLine(w, protocol.Event{ID: sub.ID, Msg: protocol.EventMsg{Type: protocol.EventTaskStarted}})

            // 2. agent_message (minimal – echo or static reply)
            text := textFromUserInput(sub.Op)
            reply := "Hi there"
            if text != "" {
                reply = fmt.Sprintf("You said: %s", text)
            }
            _ = writeJSONLine(w, protocol.Event{ID: sub.ID, Msg: protocol.EventMsg{Type: protocol.EventAgentMessage, Text: reply}})

            // 3. task_complete
            _ = writeJSONLine(w, protocol.Event{ID: sub.ID, Msg: protocol.EventMsg{Type: protocol.EventTaskComplete}})

        case protocol.OpInterrupt:
            // Emit an error for this submission. In later phases, this would
            // target the currently running task's id.
            _ = writeJSONLine(w, protocol.Event{ID: sub.ID, Msg: protocol.EventMsg{Type: protocol.EventError, Message: "interrupted"}})

        default:
            _ = writeJSONLine(w, protocol.Event{ID: sub.ID, Msg: protocol.EventMsg{Type: protocol.EventError, Message: "unsupported op"}})
        }
    }
    return scanner.Err()
}

