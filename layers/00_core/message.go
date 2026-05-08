package core

// Role identifies the sender of a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single message in a conversation. Fields are populated
// depending on the Role:
//   - system/user: Content only
//   - assistant: Content and/or ToolCalls
//   - tool: Content, ToolCallID, and Name (result of a tool execution)
type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall // assistant requesting tool execution
	ToolCallID string     // tool result referencing original call
	Name       string     // tool name for tool result messages
}

// NewSystemMessage creates a system prompt message.
func NewSystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// NewUserMessage creates a user message.
func NewUserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// NewAssistantMessage creates an assistant response message.
func NewAssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// NewToolResultMessage creates a message carrying a tool execution result.
func NewToolResultMessage(callID, name, content string) Message {
	return Message{
		Role:       RoleTool,
		Content:    content,
		ToolCallID: callID,
		Name:       name,
	}
}
