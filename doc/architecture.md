# Go 实现架构设计

> 本文档定义了 Go 版本 claude-code-acp 的架构设计和目录结构。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              ACP Client (Zed)                           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ NDJSON over stdin/stdout
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Go ACP Adapter                                  │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Protocol  │  │   Session   │  │    Tools    │  │   Settings  │   │
│  │    Layer    │  │   Manager   │  │  Converter  │  │   Manager   │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                        MCP Server                                │   │
│  │   (Read, Write, Edit, Bash, BashOutput, KillShell)              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Claude Agent SDK (Go)                         │   │
│  │                   github.com/schlunsen/claude-agent-sdk-go       │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Subprocess / API
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Claude Code CLI                                 │
│                      (Anthropic Claude API)                              │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 目录结构

```
claude-code-acp-go/
├── cmd/
│   └── claude-code-acp/
│       └── main.go              # CLI 入口点
│
├── internal/
│   ├── acp/
│   │   ├── agent.go             # 核心 ACP 代理实现
│   │   ├── agent_test.go
│   │   ├── session.go           # 会话管理
│   │   ├── session_test.go
│   │   ├── permissions.go       # 权限控制
│   │   ├── permissions_test.go
│   │   ├── protocol.go          # ACP 协议类型定义
│   │   ├── protocol_test.go
│   │   ├── prompts.go           # Prompt 处理
│   │   ├── prompts_test.go
│   │   ├── notifications.go     # 通知发送
│   │   └── doc.go
│   │
│   ├── tools/
│   │   ├── converter.go         # 工具格式转换
│   │   ├── converter_test.go
│   │   ├── types.go             # 工具类型定义
│   │   ├── read.go              # Read 工具处理
│   │   ├── write.go             # Write 工具处理
│   │   ├── edit.go              # Edit 工具处理
│   │   ├── bash.go              # Bash 工具处理
│   │   └── doc.go
│   │
│   ├── mcp/
│   │   ├── server.go            # MCP 服务器实现
│   │   ├── server_test.go
│   │   ├── tools.go             # MCP 工具注册
│   │   └── doc.go
│   │
│   ├── settings/
│   │   ├── manager.go           # 设置管理器
│   │   ├── manager_test.go
│   │   ├── permissions.go       # 权限规则
│   │   ├── permissions_test.go
│   │   ├── sources.go           # 设置来源
│   │   └── doc.go
│   │
│   ├── transport/
│   │   ├── ndjson.go            # NDJSON 编解码
│   │   ├── ndjson_test.go
│   │   ├── stdio.go             # stdio 传输
│   │   └── doc.go
│   │
│   └── utils/
│       ├── pushable.go          # Pushable 流实现
│       ├── pushable_test.go
│       ├── encoding.go          # 路径编码等
│       └── doc.go
│
├── pkg/
│   └── api/
│       ├── client.go            # 公开客户端 API
│       └── options.go           # 配置选项
│
├── e2e/
│   ├── compat/
│   │   └── compat_test.go       # 兼容性测试
│   ├── conformance/
│   │   └── conformance_test.go  # 协议合规测试
│   └── zed/
│       └── zed_test.go          # Zed 集成测试
│
├── golden/
│   ├── initialize-basic.jsonl   # 黄金测试文件
│   ├── session-new.jsonl
│   └── ...
│
├── testdata/
│   ├── test.png                 # 测试图片
│   └── test-files/              # 测试文件
│
├── doc/
│   ├── README.md
│   ├── analysis.md
│   ├── sdk-comparison.md
│   ├── test-suite.md
│   ├── implementation-guide.md
│   └── architecture.md
│
├── scripts/
│   ├── record-golden.ts         # 录制黄金文件
│   └── validate-types.ts        # 验证类型映射
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 核心组件设计

### 1. Agent (internal/acp/agent.go)

```go
package acp

import (
    "context"
    "sync"

    sdk "github.com/schlunsen/claude-agent-sdk-go"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClaudeAcpAgent 实现 ACP Agent 接口
type ClaudeAcpAgent struct {
    mu                sync.RWMutex
    sessions          map[string]*Session
    client            *AgentSideConnection
    toolUseCache      ToolUseCache
    clientCapabilities *ClientCapabilities
    logger            *Logger
    mcpServer         *ACPMCPServer
}

// Session 表示一个 ACP 会话
type Session struct {
    ID             string
    Cwd            string
    Query          *sdk.Query
    Input          *Pushable[sdk.UserMessage]
    Cancelled      bool
    PermissionMode PermissionMode
    Settings       *settings.Manager
    Terminals      map[string]*Terminal
}

// Terminal 表示一个终端会话
type Terminal struct {
    ID          string
    Handle      *TerminalHandle
    Status      TerminalStatus
    LastOutput  *TerminalOutputResponse
}

type TerminalStatus string

const (
    TerminalStarted  TerminalStatus = "started"
    TerminalExited   TerminalStatus = "exited"
    TerminalKilled   TerminalStatus = "killed"
    TerminalTimedOut TerminalStatus = "timedOut"
    TerminalAborted  TerminalStatus = "aborted"
)

// NewClaudeAcpAgent 创建新的 ACP 代理
func NewClaudeAcpAgent(opts ...AgentOption) *ClaudeAcpAgent {
    agent := &ClaudeAcpAgent{
        sessions:     make(map[string]*Session),
        toolUseCache: make(ToolUseCache),
        logger:       NewLogger(os.Stderr),
    }

    for _, opt := range opts {
        opt(agent)
    }

    return agent
}

// Run 启动代理，监听 stdin/stdout
func (a *ClaudeAcpAgent) Run(ctx context.Context) error {
    transport := transport.NewNDJSONTransport(os.Stdin, os.Stdout)
    a.client = NewAgentSideConnection(transport, a)
    return a.client.Run(ctx)
}
```

### 2. Protocol Types (internal/acp/protocol.go)

```go
package acp

// 从 ACP Schema 精确映射的类型

// InitializeRequest 初始化请求
type InitializeRequest struct {
    ProtocolVersion    int                `json:"protocolVersion"`
    ClientCapabilities ClientCapabilities `json:"clientCapabilities"`
    ClientInfo         *Implementation    `json:"clientInfo,omitempty"`
    Meta               map[string]any     `json:"_meta,omitempty"`
}

// InitializeResponse 初始化响应
type InitializeResponse struct {
    ProtocolVersion   int               `json:"protocolVersion"`
    AgentCapabilities AgentCapabilities `json:"agentCapabilities"`
    AgentInfo         *Implementation   `json:"agentInfo,omitempty"`
    AuthMethods       []AuthMethod      `json:"authMethods,omitempty"`
    Meta              map[string]any    `json:"_meta,omitempty"`
}

// ClientCapabilities 客户端能力
type ClientCapabilities struct {
    FS       FileSystemCapability `json:"fs,omitempty"`
    Terminal bool                 `json:"terminal,omitempty"`
    Meta     map[string]any       `json:"_meta,omitempty"`
}

// FileSystemCapability 文件系统能力
type FileSystemCapability struct {
    ReadTextFile  bool `json:"readTextFile,omitempty"`
    WriteTextFile bool `json:"writeTextFile,omitempty"`
    Meta          map[string]any `json:"_meta,omitempty"`
}

// AgentCapabilities 代理能力
type AgentCapabilities struct {
    LoadSession        bool               `json:"loadSession,omitempty"`
    PromptCapabilities PromptCapabilities `json:"promptCapabilities,omitempty"`
    MCPCapabilities    MCPCapabilities    `json:"mcpCapabilities,omitempty"`
    SessionCapabilities SessionCapabilities `json:"sessionCapabilities,omitempty"`
    Meta               map[string]any     `json:"_meta,omitempty"`
}

// PromptCapabilities Prompt 能力
type PromptCapabilities struct {
    Image           bool `json:"image,omitempty"`
    Audio           bool `json:"audio,omitempty"`
    EmbeddedContext bool `json:"embeddedContext,omitempty"`
    Meta            map[string]any `json:"_meta,omitempty"`
}

// MCPCapabilities MCP 能力
type MCPCapabilities struct {
    HTTP bool `json:"http,omitempty"`
    SSE  bool `json:"sse,omitempty"`
    Meta map[string]any `json:"_meta,omitempty"`
}

// ContentBlock 内容块
type ContentBlock struct {
    Type     string          `json:"type"`
    Text     string          `json:"text,omitempty"`
    Image    *ImageContent   `json:"image,omitempty"`
    Audio    *AudioContent   `json:"audio,omitempty"`
    Resource *EmbeddedResource `json:"resource,omitempty"`
    ResourceLink *ResourceLink `json:"resourceLink,omitempty"`
    Meta     map[string]any  `json:"_meta,omitempty"`
}

// SessionUpdate 会话更新
type SessionUpdate struct {
    SessionUpdate string `json:"sessionUpdate"`

    // 各种更新类型的字段
    Content       *TextContent     `json:"content,omitempty"`
    ToolCallID    string           `json:"toolCallId,omitempty"`
    Kind          string           `json:"kind,omitempty"`
    Title         string           `json:"title,omitempty"`
    Status        string           `json:"status,omitempty"`
    RawInput      any              `json:"rawInput,omitempty"`
    RawOutput     any              `json:"rawOutput,omitempty"`
    Content_      []ToolCallContent `json:"content,omitempty"`
    Locations     []ToolCallLocation `json:"locations,omitempty"`
    Entries       []PlanEntry      `json:"entries,omitempty"`
    CurrentModeID string           `json:"currentModeId,omitempty"`
    AvailableCommands []AvailableCommand `json:"availableCommands,omitempty"`
    ConfigOptions []SessionConfigOption `json:"configOptions,omitempty"`
    Meta          map[string]any   `json:"_meta,omitempty"`
}
```

### 3. Tools Converter (internal/tools/converter.go)

```go
package tools

// ToolInfo ACP 工具信息
type ToolInfo struct {
    Title     string              `json:"title"`
    Kind      string              `json:"kind"`
    Content   []ToolCallContent   `json:"content,omitempty"`
    Locations []ToolCallLocation  `json:"locations,omitempty"`
}

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

// ToolInfoFromToolUse 从 SDK 工具使用转换为 ACP ToolInfo
func ToolInfoFromToolUse(toolUse sdk.ToolUse) ToolInfo {
    switch toolUse.Name {
    case "Read", "mcp__acp__Read":
        return convertReadTool(toolUse)
    case "Write", "mcp__acp__Write":
        return convertWriteTool(toolUse)
    case "Edit", "mcp__acp__Edit":
        return convertEditTool(toolUse)
    case "Bash", "mcp__acp__Bash":
        return convertBashTool(toolUse)
    case "Glob":
        return convertGlobTool(toolUse)
    case "Grep":
        return convertGrepTool(toolUse)
    case "Task":
        return convertTaskTool(toolUse)
    case "TodoWrite":
        return convertTodoWriteTool(toolUse)
    default:
        return ToolInfo{
            Title: toolUse.Name,
            Kind:  string(ToolKindOther),
        }
    }
}

func convertReadTool(toolUse sdk.ToolUse) ToolInfo {
    input := toolUse.Input.(map[string]any)
    filePath := input["file_path"].(string)

    title := fmt.Sprintf("Read %s", filePath)
    offset, hasOffset := input["offset"].(float64)
    limit, hasLimit := input["limit"].(float64)

    if hasOffset && hasLimit {
        title = fmt.Sprintf("Read %s (%d - %d)", filePath, int(offset)+1, int(offset+limit))
    } else if hasOffset {
        title = fmt.Sprintf("Read %s (from line %d)", filePath, int(offset)+1)
    } else if hasLimit {
        title = fmt.Sprintf("Read %s (1 - %d)", filePath, int(limit))
    }

    return ToolInfo{
        Title: title,
        Kind:  string(ToolKindRead),
        Locations: []ToolCallLocation{
            {Path: filePath, Line: int(offset)},
        },
    }
}

func convertWriteTool(toolUse sdk.ToolUse) ToolInfo {
    input := toolUse.Input.(map[string]any)
    filePath := input["file_path"].(string)
    content := input["content"].(string)

    return ToolInfo{
        Title: fmt.Sprintf("Write %s", filePath),
        Kind:  string(ToolKindEdit),
        Content: []ToolCallContent{
            {
                Type:   "diff",
                Path:   filePath,
                OldText: nil,
                NewText: content,
            },
        },
        Locations: []ToolCallLocation{
            {Path: filePath},
        },
    }
}

func convertBashTool(toolUse sdk.ToolUse) ToolInfo {
    input := toolUse.Input.(map[string]any)
    command := input["command"].(string)

    return ToolInfo{
        Title: fmt.Sprintf("`%s`", command),
        Kind:  string(ToolKindExecute),
    }
}
```

### 4. MCP Server (internal/mcp/server.go)

```go
package mcp

import (
    "context"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ACPMCPServer 为 ACP 提供 MCP 工具
type ACPMCPServer struct {
    server *mcp.Server
    agent  *acp.ClaudeAcpAgent
}

// NewACPMCPServer 创建新的 ACP MCP 服务器
func NewACPMCPServer(agent *acp.ClaudeAcpAgent) *ACPMCPServer {
    s := &ACPMCPServer{
        server: mcp.NewServer(&mcp.Implementation{
            Name:    "acp-tools",
            Version: "1.0.0",
        }, nil),
        agent: agent,
    }

    s.registerTools()
    return s
}

func (s *ACPMCPServer) registerTools() {
    // Read 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "Read",
        Description: "Read file via ACP client",
    }, s.handleRead)

    // Write 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "Write",
        Description: "Write file via ACP client",
    }, s.handleWrite)

    // Edit 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "Edit",
        Description: "Edit file via ACP client",
    }, s.handleEdit)

    // Bash 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "Bash",
        Description: "Execute command via ACP client",
    }, s.handleBash)

    // BashOutput 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "BashOutput",
        Description: "Get output from background shell",
    }, s.handleBashOutput)

    // KillShell 工具
    mcp.AddTool(s.server, &mcp.Tool{
        Name:        "KillShell",
        Description: "Kill background process",
    }, s.handleKillShell)
}

type ReadInput struct {
    FilePath string `json:"file_path" jsonschema:"required,path to the file"`
    Offset   int    `json:"offset,omitempty,line number to start from"`
    Limit    int    `json:"limit,omitempty,maximum number of lines"`
}

type ReadOutput struct {
    Content string `json:"content"`
}

func (s *ACPMCPServer) handleRead(ctx context.Context, req *mcp.CallToolRequest, input ReadInput) (
    *mcp.CallToolResult, ReadOutput, error,
) {
    // 调用 ACP 客户端读取文件
    result, err := s.agent.Client().ReadTextFile(ctx, acp.ReadTextFileRequest{
        SessionID: getSessionID(ctx),
        Path:      input.FilePath,
        Offset:    input.Offset,
        Limit:     input.Limit,
    })

    if err != nil {
        return nil, ReadOutput{}, err
    }

    return nil, ReadOutput{Content: result.Content}, nil
}
```

### 5. Settings Manager (internal/settings/manager.go)

```go
package settings

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// Manager 管理 Claude Code 设置
type Manager struct {
    cwd        string
    sources    map[SourceType]*ClaudeCodeSettings
    merged     *ClaudeCodeSettings
}

// SourceType 设置来源类型
type SourceType string

const (
    SourceUser     SourceType = "user"
    SourceProject  SourceType = "project"
    SourceLocal    SourceType = "local"
    SourceManaged  SourceType = "managed"
)

// ClaudeCodeSettings Claude Code 设置
type ClaudeCodeSettings struct {
    Permissions *PermissionSettings `json:"permissions,omitempty"`
    Env         map[string]string   `json:"env,omitempty"`
    Model       string              `json:"model,omitempty"`
}

// PermissionSettings 权限设置
type PermissionSettings struct {
    Allow              []string `json:"allow,omitempty"`
    Deny               []string `json:"deny,omitempty"`
    Ask                []string `json:"ask,omitempty"`
    AdditionalDirectories []string `json:"additionalDirectories,omitempty"`
    DefaultMode        string   `json:"defaultMode,omitempty"`
}

// NewManager 创建新的设置管理器
func NewManager(cwd string) *Manager {
    return &Manager{
        cwd:     cwd,
        sources: make(map[SourceType]*ClaudeCodeSettings),
        merged:  &ClaudeCodeSettings{},
    }
}

// LoadAll 加载所有设置来源
func (m *Manager) LoadAll() error {
    // 1. 用户设置
    m.load(SourceUser, m.userSettingsPath())

    // 2. 项目设置
    m.load(SourceProject, m.projectSettingsPath())

    // 3. 本地项目设置
    m.load(SourceLocal, m.localSettingsPath())

    // 4. 企业托管设置
    m.load(SourceManaged, m.managedSettingsPath())

    // 合并设置
    m.merge()

    return nil
}

func (m *Manager) load(source SourceType, path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err // 文件不存在是正常的
    }

    var settings ClaudeCodeSettings
    if err := json.Unmarshal(data, &settings); err != nil {
        return err
    }

    m.sources[source] = &settings
    return nil
}

func (m *Manager) merge() {
    // 按优先级顺序合并: user < project < local < managed
    order := []SourceType{SourceUser, SourceProject, SourceLocal, SourceManaged}

    for _, source := range order {
        if settings, ok := m.sources[source]; ok {
            m.mergeSettings(settings)
        }
    }
}

func (m *Manager) mergeSettings(other *ClaudeCodeSettings) {
    if other.Permissions != nil {
        if m.merged.Permissions == nil {
            m.merged.Permissions = &PermissionSettings{}
        }
        // 合并权限设置...
    }

    if other.Env != nil {
        if m.merged.Env == nil {
            m.merged.Env = make(map[string]string)
        }
        for k, v := range other.Env {
            m.merged.Env[k] = v
        }
    }

    if other.Model != "" {
        m.merged.Model = other.Model
    }
}

// CheckPermission 检查工具权限
func (m *Manager) CheckPermission(toolName string, input map[string]any) PermissionDecision {
    if m.merged.Permissions == nil {
        return PermissionAsk
    }

    // 检查 deny 列表
    for _, rule := range m.merged.Permissions.Deny {
        if m.matchRule(rule, toolName, input) {
            return PermissionDeny
        }
    }

    // 检查 allow 列表
    for _, rule := range m.merged.Permissions.Allow {
        if m.matchRule(rule, toolName, input) {
            return PermissionAllow
        }
    }

    // 检查 ask 列表
    for _, rule := range m.merged.Permissions.Ask {
        if m.matchRule(rule, toolName, input) {
            return PermissionAsk
        }
    }

    return PermissionAsk
}

func (m *Manager) matchRule(rule, toolName string, input map[string]any) bool {
    // 解析规则格式: ToolName 或 ToolName(args)
    // 例如: "Read", "Read(./.env)", "Bash(npm run:*)"
    // ...
}
```

### 6. Pushable Stream (internal/utils/pushable.go)

```go
package utils

import (
    "context"
    "sync"
)

// Pushable 是一个可推送的异步可迭代对象
type Pushable[T any] struct {
    ch     chan T
    closed bool
    mu     sync.Mutex
}

// NewPushable 创建新的 Pushable
func NewPushable[T any](bufferSize int) *Pushable[T] {
    return &Pushable[T]{
        ch: make(chan T, bufferSize),
    }
}

// Push 添加元素
func (p *Pushable[T]) Push(item T) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.closed {
        return
    }

    p.ch <- item
}

// End 结束流
func (p *Pushable[T]) End() {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.closed {
        return
    }

    p.closed = true
    close(p.ch)
}

// Channel 返回底层 channel
func (p *Pushable[T]) Channel() <-chan T {
    return p.ch
}

// Iter 返回迭代器
func (p *Pushable[T]) Iter(ctx context.Context) <-chan T {
    return p.ch
}
```

---

## 依赖关系

```
cmd/claude-code-acp
       │
       ▼
   internal/acp ◄─────────────────────────────┐
       │                                      │
       ├──► internal/transport                │
       │         │                            │
       │         └──► NDJSON 编解码           │
       │                                      │
       ├──► internal/tools                    │
       │         │                            │
       │         └──► 工具格式转换            │
       │                                      │
       ├──► internal/mcp                      │
       │         │                            │
       │         └──► MCP 服务器              │
       │                                      │
       ├──► internal/settings                 │
       │         │                            │
       │         └──► 设置管理                │
       │                                      │
       └──► internal/utils                    │
                 │                            │
                 └──► 通用工具                │
                                            │
       ┌────────────────────────────────────┘
       │
       ▼
  外部依赖:
  ├── github.com/schlunsen/claude-agent-sdk-go
  └── github.com/modelcontextprotocol/go-sdk
```

---

## Makefile

```makefile
# Makefile

.PHONY: all build test clean lint

# 默认目标
all: build

# 构建
build:
	go build -o bin/claude-code-acp ./cmd/claude-code-acp

# 运行测试
test:
	go test ./... -v

# 运行测试 (带覆盖率)
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# 运行竞态检测
test-race:
	go test -race ./...

# 运行 E2E 测试
test-e2e:
	go test -tags=e2e ./e2e/... -v

# 运行兼容性测试
test-compat:
	go test ./e2e/compat/... -v -timeout 30m

# 代码检查
lint:
	golangci-lint run

# 格式化
fmt:
	go fmt ./...

# 清理
clean:
	rm -rf bin/
	go clean

# 安装依赖
deps:
	go mod download
	go mod tidy

# 生成类型 (从 ACP Schema)
generate:
	go generate ./...

# 录制黄金文件
record-golden:
	cd scripts && npm install && npm run record-golden

# 验证类型映射
validate-types:
	cd scripts && npm run validate-types
```

---

## go.mod

```go
module github.com/your-org/claude-code-acp-go

go 1.24

require (
    // Claude Agent SDK (非官方社区移植)
    github.com/schlunsen/claude-agent-sdk-go v0.1.0

    // MCP SDK (官方)
    github.com/modelcontextprotocol/go-sdk v0.2.0

    // 测试
    github.com/stretchr/testify v1.9.0
)

require (
    // 间接依赖...
)
```

---

## 实现路线图

### 阶段 1: 基础设施 (1周)

- [ ] 项目结构初始化
- [ ] 类型定义 (从 ACP Schema)
- [ ] NDJSON 传输层
- [ ] 基础测试框架

### 阶段 2: 协议层 (2周)

- [ ] `initialize` 方法
- [ ] `session/new` 方法
- [ ] `session/load` 方法
- [ ] `session/cancel` 方法
- [ ] 黄金文件测试

### 阶段 3: Prompt 处理 (2周)

- [ ] Prompt 转换
- [ ] 流式响应
- [ ] 通知发送
- [ ] Stop reasons

### 阶段 4: 工具系统 (2周)

- [ ] 工具转换器
- [ ] 权限请求
- [ ] MCP 服务器
- [ ] ACP 工具实现

### 阶段 5: 高级功能 (1周)

- [ ] 设置管理
- [ ] 模式切换
- [ ] Plan 支持
- [ ] 斜杠命令

### 阶段 6: 验证与优化 (1周)

- [ ] 兼容性测试
- [ ] Zed 集成测试
- [ ] 性能优化
- [ ] 文档完善

**总计: 9周**

---

*文档生成日期: 2026-02-13*