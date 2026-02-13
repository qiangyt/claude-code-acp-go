package tools

// ToolKind 工具类型
type ToolKind string

const (
	ToolKindRead       ToolKind = "read"
	ToolKindEdit       ToolKind = "edit"
	ToolKindMove       ToolKind = "move"
	ToolKindSearch     ToolKind = "search"
	ToolKindExecute    ToolKind = "execute"
	ToolKindThink      ToolKind = "think"
	ToolKindFetch      ToolKind = "fetch"
	ToolKindSwitchMode ToolKind = "switch_mode"
	ToolKindOther      ToolKind = "other"
)

// ToolCallLocation 工具调用位置
type ToolCallLocation struct {
	Path string `json:"path,omitempty"`
	Line int    `json:"line,omitempty"`
}

// ToolCallContent 工具调用内容
type ToolCallContent struct {
	Type    string  `json:"type,omitempty"`
	Path    string  `json:"path,omitempty"`
	OldText *string `json:"oldText,omitempty"`
	NewText string  `json:"newText,omitempty"`
}

// ToolInfo ACP 工具信息
type ToolInfo struct {
	Title     string             `json:"title,omitempty"`
	Kind      ToolKind           `json:"kind,omitempty"`
	Content   []ToolCallContent  `json:"content,omitempty"`
	Locations []ToolCallLocation `json:"locations,omitempty"`
}
