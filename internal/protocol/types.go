package protocol

// This package defines the core protocol types that mirror the Rust
// codex-rs "protocol" crate at a high level. The design follows a
// Submission Queue (SQ) / Event Queue (EQ) model:
//
// - UI sends Submissions (Op) to the agent.
// - Agent responds with Events (EventMsg) back to the UI.
//
// We intentionally keep the initial set minimal and human-readable to support
// incremental learning. Over time this can be extended to more variants.

// Submission represents a single request from the UI to the agent.
// The UI provides a unique id to correlate follow-up Events.
type Submission struct {
    ID string `json:"id"`
    Op Op    `json:"op"`
}

// Op is a tagged union of possible submission payloads.
// We encode it as a struct with a Type discriminator for simplicity.
// This mirrors serde's #[serde(tag = "type")] approach in Rust.
type Op struct {
    Type string `json:"type"`

    // user_input fields
    Items []InputItem `json:"items,omitempty"`

    // exec_approval fields
    ApprovalID string         `json:"id,omitempty"`
    Decision   ReviewDecision `json:"decision,omitempty"`
}

// Well-known Op.Type values (subset).
const (
    OpInterrupt   = "interrupt"
    OpUserInput   = "user_input"
    OpExecApproval = "exec_approval"
)

// InputItem is a user-provided content item. We support a minimal set
// for now (text and image by URL, or a local file path to be converted later).
type InputItem struct {
    Type string `json:"type"` // "text" | "image" | "local_image"

    // Text content when Type=="text".
    Text string `json:"text,omitempty"`

    // ImageURL when Type=="image" (data: URI or remote URL).
    ImageURL string `json:"image_url,omitempty"`

    // Local file path when Type=="local_image".
    Path string `json:"path,omitempty"`
}

// Event represents a single message from the agent back to the UI that
// correlates to a Submission ID.
type Event struct {
    ID  string  `json:"id"`
    Msg EventMsg `json:"msg"`
}

// EventMsg is a tagged union for agent->UI messages. We start with a compact
// subset that is enough to demonstrate task lifecycle and messaging.
type EventMsg struct {
    Type string `json:"type"` // e.g., "error", "task_started", "task_complete", "agent_message", "exec_approval_request"

    // error
    ErrorMessage string `json:"message,omitempty"`

    // task_started
    ModelContextWindow *uint64 `json:"model_context_window,omitempty"`

    // task_complete
    LastAgentMessage string `json:"last_agent_message,omitempty"`

    // agent_message
    AgentText string `json:"text,omitempty"`

    // exec_approval_request
    CallID string   `json:"call_id,omitempty"`
    Command []string `json:"command,omitempty"`
    Cwd     string   `json:"cwd,omitempty"`
    Reason  string   `json:"reason,omitempty"`
}

// Well-known EventMsg.Type values (subset).
const (
    EventError               = "error"
    EventTaskStarted         = "task_started"
    EventTaskComplete        = "task_complete"
    EventAgentMessage        = "agent_message"
    EventExecApprovalRequest = "exec_approval_request"
)

// ReviewDecision mirrors a small subset of the Rust enum used when the user
// approves or denies an execution request.
type ReviewDecision string

const (
    DecisionApproved           ReviewDecision = "approved"
    DecisionApprovedForSession ReviewDecision = "approved_for_session"
    DecisionDenied             ReviewDecision = "denied"
    DecisionAbort              ReviewDecision = "abort"
)
