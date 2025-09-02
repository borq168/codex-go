package mcp

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
)

// request is a minimal shape for line-delimited JSON requests to MCP.
// In a full JSON-RPC implementation we'd parse id/version, but for learning
// we start with a tiny method switch that keeps I/O simple and observable.
type request struct {
    Method string `json:"method"`
    // Optional text payload for echo requests.
    Text   string `json:"text,omitempty"`
}

// pong is the success payload we return when Method=="ping".
type pong struct {
    Result string `json:"result"`
}

// errResp is a small error envelope for malformed input or unknown methods.
type errResp struct {
    Error string `json:"error"`
}

// Serve implements a minimal JSON-over-stdio handler.
// Each input line is expected to be a JSON object with at least a "method" field.
// Supported: {"method":"ping"} -> {"result":"pong"}\n
// The function is streaming: it reads until EOF/caller closes stdin. This makes
// it easy to later evolve into a proper JSON-RPC loop without changing the
// high-level control flow in cmd/codex.
func Serve(ctx context.Context, r io.Reader, w io.Writer) error {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        line := scanner.Bytes()

        var req request
        if err := json.Unmarshal(line, &req); err != nil {
            // For invalid JSON, reply with a one-line error and keep the loop alive.
            if _, werr := fmt.Fprintln(w, `{"error":"invalid json"}`); werr != nil {
                return werr
            }
            continue
        }

        switch req.Method {
        case "ping":
            // Happy path: reply with pong.
            b, _ := json.Marshal(pong{Result: "pong"})
            if _, err := w.Write(append(b, '\n')); err != nil {
                return err
            }
        case "echo":
            // Minimal echo route: return an agent_message with the provided text.
            // Shape mirrors a tiny slice of our EventMsg for learning purposes.
            type agentMsg struct {
                Type string `json:"type"`
                Text string `json:"text,omitempty"`
            }
            if req.Text == "" {
                if _, err := fmt.Fprintln(w, `{"error":"missing text"}`); err != nil {
                    return err
                }
                continue
            }
            b, _ := json.Marshal(agentMsg{Type: "agent_message", Text: req.Text})
            if _, err := w.Write(append(b, '\n')); err != nil {
                return err
            }
        default:
            // Unrecognized method – respond with a small error message.
            b, _ := json.Marshal(errResp{Error: "method not implemented"})
            if _, err := w.Write(append(b, '\n')); err != nil {
                return err
            }
        }
    }
    return scanner.Err()
}
