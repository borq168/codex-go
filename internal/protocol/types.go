package protocol

// Minimal protocol v1 core for a first learning iteration.
// SQ/EQ 模型（最小实现）：
// - UI -> Agent: Submission { id, op }
// - Agent -> UI: Event { id, msg }
// - Op/EventMsg 使用 "type" 作为判别字段；仅保留最少必要的几类。

// Submission: UI 发送给 Agent 的一条请求。id 用于回溯匹配后续 Event。
type Submission struct {
    ID string `json:"id"`
    Op Op    `json:"op"`
}

// Op: 提交的具体操作（最小子集）。
// - "user_input": items=[{type:"text", text:"..."}, ...]
// - "interrupt": 无额外字段
type Op struct {
    Type  string      `json:"type"`           // "user_input" | "interrupt"
    Items []InputItem `json:"items,omitempty"` // 仅当 type=="user_input" 时使用
}

const (
    OpUserInput = "user_input"
    OpInterrupt = "interrupt"
)

// InputItem: 用户输入项（最小实现只支持文本）。
type InputItem struct {
    Type string `json:"type"`           // 固定为 "text"
    Text string `json:"text,omitempty"` // 文本内容
}

// Event: Agent 发送给 UI 的响应消息。id 与 Submission.id 对应。
type Event struct {
    ID  string   `json:"id"`
    Msg EventMsg `json:"msg"`
}

// EventMsg: Agent -> UI 的事件（最小子集）。
// - "task_started": 开始处理一次用户输入
// - "agent_message": Agent 的文本输出（一次或多次）
// - "task_complete": 本次处理完成
// - "error": 出错信息
type EventMsg struct {
    Type string `json:"type"` // "task_started" | "agent_message" | "task_complete" | "error"

    // agent_message / error
    Text    string `json:"text,omitempty"`    // agent_message 文本
    Message string `json:"message,omitempty"` // error 文本
}

const (
    EventTaskStarted  = "task_started"
    EventAgentMessage = "agent_message"
    EventTaskComplete = "task_complete"
    EventError        = "error"
)

// 示例 JSON（最小）：
// Submission (user_input):
// {"id":"sub-1","op":{"type":"user_input","items":[{"type":"text","text":"Hello"}]}}
// Submission (interrupt):
// {"id":"sub-2","op":{"type":"interrupt"}}
// Event 序列:
// {"id":"sub-1","msg":{"type":"task_started"}}
// {"id":"sub-1","msg":{"type":"agent_message","text":"Hi there"}}
// {"id":"sub-1","msg":{"type":"task_complete"}}
